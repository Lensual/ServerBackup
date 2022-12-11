package storage

import (
	"context"
	"fmt"
	"io"
)

type Storage interface {
	CreateFile(ctx context.Context, path string) (io.WriteCloser, error)
}

type StorageFactory interface {
	Name() string
	Instance(opt map[string]interface{}) (Storage, error)
}

var registered = make(map[string]StorageFactory)

func Register(f StorageFactory) {
	if f == nil {
		panic("cannot register a nil Storage")
	}
	if f.Name() == "" {
		panic("cannot register Storage with empty string result for Name()")
	}
	registered[f.Name()] = f
}

func New(driver string, opt map[string]interface{}) (Storage, error) {
	f, ok := registered[driver]
	if !ok {
		return nil, fmt.Errorf("storage driver \"%s\" not found", driver)
	}
	return f.Instance(opt)
}
