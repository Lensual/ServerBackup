package worker

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"

	"github.com/Lensual/ServerBackup/config"
	"github.com/Lensual/ServerBackup/pkg/backuper"
	"github.com/Lensual/ServerBackup/pkg/compressor"
	"github.com/Lensual/ServerBackup/pkg/storage"
)

type Worker struct {
	taskName string
	taskConf config.TaskTemplate
	logger   *log.Logger
}

func New(opts ...func(*Worker)) *Worker {
	w := &Worker{}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func WithTaskName(taskName string) func(*Worker) {
	return func(w *Worker) {
		w.taskName = taskName
	}
}

func WithTaskConfig(taskConf config.TaskTemplate) func(*Worker) {
	return func(w *Worker) {
		w.taskConf = taskConf
	}
}

func WithLogger(logger *log.Logger) func(*Worker) {
	return func(w *Worker) {
		w.logger = logger
	}
}

// Implement cron.Job
func (w *Worker) Run() {
	err := func(w *Worker) error {
		var bk backuper.Backuper
		var st storage.Storage
		var cpr compressor.Compressor

		var err error

		bk, err = w.dependBackuper()
		if err != nil {
			return err
		}
		st, err = w.dependStorage()
		if err != nil {
			return err
		}
		cpr, err = w.dependCompressor()
		if err != nil {
			return err
		}

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := bk.Run(ctx)
			if err != nil {
				w.logger.Printf("%v", err)
				cancel()
			}
		}()

		for {
			var file *backuper.File
			file, err = bk.RecvFile(ctx)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					w.logger.Printf("%v", err)
					cancel()
				}
				break
			}

			file.Path = cpr.Rename(file.Path)

			writeCloser, err := st.CreateFile(ctx, file.Path)
			if err != nil {
				w.logger.Printf("%v", err)
				cancel()
				return err
			}

			wg.Add(1)
			go func(ctx context.Context, rc io.ReadCloser, wc io.WriteCloser) {
				defer wg.Done()
				defer rc.Close()
				defer wc.Close()
				err := cpr.CompressFile(ctx, rc, wc)
				if err != nil {
					w.logger.Printf("%v", err)
					return
				}
			}(ctx, file.ReadCloser, writeCloser)
		}

		wg.Wait()

		return nil
	}(w)
	if err != nil {
		w.logger.Printf("Task Failed: %v %v", w.taskName, err)
		return
	}
	w.logger.Printf("Task Done: %v", w.taskName)
}

func (w *Worker) dependBackuper() (backuper.Backuper, error) {
	driver := w.taskConf.Backup.Driver
	opt, _ := w.taskConf.Backup.Option[driver].(map[string]interface{})
	return backuper.New(driver, w.taskName, w.logger, opt)
}

func (w *Worker) dependStorage() (storage.Storage, error) {
	driver := w.taskConf.Storage.Driver
	opt, _ := w.taskConf.Storage.Option[driver].(map[string]interface{})
	return storage.New(driver, opt)
}

func (w *Worker) dependCompressor() (compressor.Compressor, error) {
	driver := w.taskConf.Compress.Driver
	opt, _ := w.taskConf.Compress.Option[driver].(map[string]interface{})
	return compressor.New(driver, opt)
}
