package rsyncmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Lensual/ServerBackup/pkg/backuper"
	"github.com/mitchellh/mapstructure"
)

const Name = "rsyncmd"

func init() {
	backuper.Register(rsyncmdFactory{})
}

type rsyncmdFactory struct{}

func (f rsyncmdFactory) Name() string {
	return Name
}

func (f rsyncmdFactory) Instance(taskName string, logger *log.Logger, opt map[string]interface{}) (backuper.Backuper, error) {
	bk := rsyncmd{
		taskName: taskName,
		logger:   logger,
		fileChan: make(chan *backuper.File),
	}
	err := mapstructure.Decode(&opt, &bk.conf)
	if err != nil {
		return nil, err
	}
	bk.taskPath = filepath.Join(bk.conf.CachePath, taskName)
	return &bk, err
}

type rsyncmd struct {
	taskName string
	logger   *log.Logger
	conf     rsyncConfig
	fileChan chan *backuper.File
	taskPath string
}

type rsyncConfig struct {
	//允许增量备份几次，达到后执行全量备份
	Incremental int       `mapstructure:"incremental"`
	CachePath   string    `mapstructure:"cachePath"`
	Host        string    `mapstructure:"host"`
	Port        int       `mapstructure:"port"`
	User        string    `mapstructure:"user"`
	UseSsh      bool      `mapstructure:"useSsh"`
	Ssh         sshConfig `mapstructure:"ssh"`
	Target      string    `mapstructure:"target"`
	Include     []string  `mapstructure:"include"`
	Exclude     []string  `mapstructure:"exclude"`
	BwLimit     int       `mapstructure:"bwlimit"`
	Nice        int       `mapstructure:"nice"`
	Option      []string  `mapstructure:"option"`
}

type sshConfig struct {
	Port                  int    `mapstructure:"port"`
	PrivateKey            string `mapstructure:"privKey"`
	StrictHostKeyChecking bool   `mapstructure:"strictHostKeyChecking"`
}

func (bk *rsyncmd) Run(ctx context.Context) error {
	defer close(bk.fileChan)

	err := bk.rsync(ctx)
	if err != nil {
		return err
	}

	err = bk.tar(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (bk *rsyncmd) RecvFile(ctx context.Context) (file *backuper.File, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case file := <-bk.fileChan:
		if file == nil {
			return nil, io.EOF
		}
		return file, nil
	}
}

func (bk *rsyncmd) wait(ctx context.Context, proc *os.Process) error {
	var err error
	var procStat *os.ProcessState
	waitDone := make(chan struct{})
	go func() {
		defer close(waitDone)
		procStat, err = proc.Wait()
	}()

	select {
	case <-waitDone:
		if err != nil {
			return err
		}
		if !procStat.Success() {
			return errors.New(procStat.String())
		}
		return nil
	case <-ctx.Done():
		proc.Kill()
		return ctx.Err()
	}
}

func (bk *rsyncmd) rsync(ctx context.Context) error {
	env := os.Environ()
	procAttr := &os.ProcAttr{
		Env: env,
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}

	procArgs := bk.buildRsyncProcArgs()
	procArgsStr := ""
	for _, v := range procArgs {
		procArgsStr += v + " "
	}
	log.Println(procArgsStr)

	proc, err := os.StartProcess("/usr/bin/rsync", procArgs, procAttr)
	if err != nil {
		return err
	}

	//TODO only linux
	_ = syscall.Setpriority(syscall.PRIO_PROCESS, proc.Pid, bk.conf.Nice)

	err = bk.wait(ctx, proc)

	return err
}

func (bk *rsyncmd) tar(ctx context.Context) error {

	now := time.Now().Format(time.RFC3339)

	isIncremental := false
	//try incremental backup
	lock, err := func() (*TarinfoLock, error) {
		//read lockfile
		lock := &TarinfoLock{}
		lockData, err := os.ReadFile(filepath.Join(bk.taskPath, "tarinfo.lock"))
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(lockData, &lock)
		if err != nil {
			return nil, err
		}
		if lock.IncrementalTimes >= bk.conf.Incremental {
			return nil, errors.New("incremental times reached")
		}

		//try incremental backup
		bk.logger.Printf("run incremental backup: incremental times unreached")
		prevTarinfo, err := os.OpenFile(filepath.Join(bk.taskPath, lock.Current), os.O_RDONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("open prevTarinfo failed: %v", err)
		}
		lock.Current = fmt.Sprintf("%s_tarinfo", now)
		currentTarinfo, err := os.OpenFile(filepath.Join(bk.taskPath, lock.Current), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return nil, fmt.Errorf("open currentTarinfo failed: %v", err)
		}
		_, err = io.Copy(currentTarinfo, prevTarinfo)
		if err != nil {
			return nil, err
		}

		isIncremental = true
		lock.IncrementalTimes++

		return lock, nil
	}()

	//full backup
	if err != nil {
		bk.logger.Printf("run full backup: %v\n", err)
		lock = &TarinfoLock{}
		lock.Origin = fmt.Sprintf("%s_tarinfo", now)
		lock.Current = lock.Origin
		lock.IncrementalTimes = 0
	}

	//save lockfile
	err = func() error {
		lockData, err := json.Marshal(&lock)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(bk.taskPath, "tarinfo.lock"), lockData, 0644)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		bk.logger.Printf("saving tarinfo.lock failed: %v\n", err)
	}

	tarinfoPath := filepath.Join(bk.taskPath, lock.Current)

	procArgs := []string{
		"tar",
		"-c",
		"-v",
		"-p",
		"-f", "-",
		"-g", tarinfoPath,
		"-C", bk.buildDstParam(),
		".",
	}
	procArgsStr := ""
	for _, v := range procArgs {
		procArgsStr += v + " "
	}
	log.Println(procArgsStr)

	r, w, err := os.Pipe() //TODO close
	if err != nil {
		return err
	}
	procAttr := &os.ProcAttr{
		Env: os.Environ(),
		Files: []*os.File{
			os.Stdin,
			w,
			os.Stderr,
		},
	}
	proc, err := os.StartProcess("/usr/bin/tar", procArgs, procAttr)
	if err != nil {
		w.Close()
		r.Close()
		return err
	}

	//output tar file
	tarName := now
	if isIncremental {
		tarName += "_incremental.tar"
	} else {
		tarName += "_full.tar"
	}
	bk.fileChan <- &backuper.File{
		Path: filepath.Join(
			bk.taskName,
			tarName,
		),
		ReadCloser: r,
	}

	err = bk.wait(ctx, proc)
	w.Close()
	if err != nil {
		return err
	}

	//output tarinfo
	tarinfo, err := os.OpenFile(tarinfoPath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	bk.fileChan <- &backuper.File{
		Path: filepath.Join(
			bk.taskName,
			lock.Current,
		),
		ReadCloser: tarinfo,
	}
	return err
}

func (bk *rsyncmd) buildRsyncProcArgs() []string {
	args := []string{
		"rsync", //argv[0]
		bk.buildPortParam(),
		// "-a", //-rlptgoD
		"--recursive", //递归
		"--links",     //保留软连接
		"--perms",     //保留权限
		"--times",     //保留时间
		"--group",     //保留所有组
		"--owner",     //保留所有者
		"--devices",   //保留设备文件信息
		"--specials",  //preserve special files

		"--sparse",             //稀疏文件处理
		"--checksum",           //checksum
		bk.buildBwLimitParam(), //带宽限制 KBPS
		"--delete",             // 删除那些DST中SRC没有的文件
		// "--block-size=SIZE", //TODO
		"--partial", //保留那些因故没有完全传输的文件，以是加快随后的再次传输
		bk.buildSshParam(),
		"--compress", //传输时压缩处理
	}
	args = append(args, bk.buildIncludeParam()...)
	args = append(args, bk.buildExcludeParam()...)
	args = append(args, []string{
		bk.buildSrcParam(),
		bk.buildDstParam(),
	}...)
	args = append(args, bk.conf.Option...)
	for k, v := range args {
		if v == "" {
			args = append(args[:k], args[k+1:]...)
		}
	}
	return args
}

func (bk *rsyncmd) buildSrcParam() string {
	return fmt.Sprintf("%s@%s:%s", bk.conf.User, bk.conf.Host, bk.conf.Target)
}

func (bk *rsyncmd) buildDstParam() string {
	rsyncCachePath := filepath.Join(bk.taskPath, "rsync")
	os.MkdirAll(rsyncCachePath, 0755)
	return rsyncCachePath
}

func (bk *rsyncmd) buildPortParam() string {
	if bk.conf.Port == 0 {
		return ""
	}
	return fmt.Sprintf("--port=%d", bk.conf.Port)
}

func (bk *rsyncmd) buildIncludeParam() []string {
	if len(bk.conf.Include) == 0 {
		return nil
	}
	inc := make([]string, 0, len(bk.conf.Include))
	for _, v := range bk.conf.Include {
		inc = append(inc, fmt.Sprintf("--include=%s", v))
	}
	return inc
}

func (bk *rsyncmd) buildExcludeParam() []string {
	if len(bk.conf.Exclude) == 0 {
		return nil
	}
	exc := make([]string, 0, len(bk.conf.Exclude))
	for _, v := range bk.conf.Exclude {
		exc = append(exc, fmt.Sprintf("--exclude=%s", v))
	}
	return exc
}

func (bk *rsyncmd) buildBwLimitParam() string {
	if bk.conf.BwLimit == 0 {
		return ""
	}
	return fmt.Sprintf("--bwlimit=%d", bk.conf.BwLimit)
}

func (bk *rsyncmd) buildSshParam() string {
	if !bk.conf.UseSsh {
		return ""
	}
	strict := "yes"
	if !bk.conf.Ssh.StrictHostKeyChecking {
		strict = "false"
	}
	return fmt.Sprintf("-e ssh -p%d -i%s -o 'StrictHostKeyChecking %s'",
		bk.conf.Ssh.Port, bk.conf.Ssh.PrivateKey, strict)
}

type TarinfoLock struct {
	Origin           string
	Current          string
	IncrementalTimes int
}
