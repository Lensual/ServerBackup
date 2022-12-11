package rsyncmd_test

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/Lensual/ServerBackup/pkg/backuper"
)

var _rsyncmdConfig = map[string]interface{}{
	"cachePath": "./serverbackup-test/backuper/rsyncmd/cache",
	"host":      "localhost",
	"port":      873,
	"user":      "lensual",
	"useSsh":    true,
	"ssh": map[string]interface{}{
		"port":                  22,
		"privKey":               "~/.ssh/id_rsa",
		"strictHostKeyChecking": false,
	},
	"target":  "/var/www",
	"bwlimit": 0,
	"nice":    19,
	"option": []string{
		"-vv",
	},
}

func TestRun(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	bk, err := backuper.New("rsyncmd", "rsyncmd-test", log.Default(), _rsyncmdConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("./serverbackup-test")

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := bk.Run(ctx)
		if err != nil {
			t.Error(err)
			cancel()
		}
	}()

	for {
		file, err := bk.RecvFile(ctx)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				t.Error(err)
				cancel()
			}
			break
		}
		println(file.Path)
		wg.Add(1)
		go func(file *backuper.File) {
			defer wg.Done()
			defer file.ReadCloser.Close()
			b, err := io.ReadAll(file.ReadCloser)
			println(len(b))
			if err != nil {
				t.Error(err)
				cancel()
			}
		}(file)
	}

	wg.Wait()
}
