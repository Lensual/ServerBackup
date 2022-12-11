package local

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/Lensual/ServerBackup/pkg/storage"
	"github.com/mitchellh/mapstructure"
)

const Name = "local"

func init() {
	storage.Register(localFactory{})
}

type local struct {
	conf localConfig
}

type localConfig struct {
	Path string `mapstructure:"path"`
}

type localFactory struct{}

func (f localFactory) Name() string {
	return Name
}

func (f localFactory) Instance(opt map[string]interface{}) (storage.Storage, error) {
	st := local{}
	err := mapstructure.Decode(&opt, &st.conf)
	if err != nil {
		return nil, err
	}
	return &st, err
}

func (st *local) CreateFile(ctx context.Context, path string) (io.WriteCloser, error) {
	filePath := filepath.Join(st.conf.Path, path)
	dirPath := filepath.Dir(filePath)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}
