package backuper

import (
	"context"
	"fmt"
	"log"
)

type Backuper interface {
	Run(ctx context.Context) error
	RecvFile(ctx context.Context) (file *File, err error)
}

type BackuperFactory interface {
	Name() string
	Instance(taskName string, logger *log.Logger, opt map[string]interface{}) (Backuper, error)
}

var registered = make(map[string]BackuperFactory)

func Register(f BackuperFactory) {
	if f == nil {
		panic("cannot register a nil Backuper")
	}
	if f.Name() == "" {
		panic("cannot register Backuper with empty string result for Name()")
	}
	registered[f.Name()] = f
}

func New(driver string, taskName string, logger *log.Logger, opt map[string]interface{}) (Backuper, error) {
	f, ok := registered[driver]
	if !ok {
		return nil, fmt.Errorf("backuper driver \"%s\" not found", driver)
	}
	return f.Instance(taskName, logger, opt)
}
