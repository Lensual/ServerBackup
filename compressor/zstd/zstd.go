package pigzcmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Lensual/ServerBackup/pkg/compressor"
	zskp "github.com/klauspost/compress/zstd"
	"github.com/mitchellh/mapstructure"
)

const Name = "zstd"

func init() {
	compressor.Register(zstdFactory{})
}

type zstdFactory struct{}

func (f zstdFactory) Name() string {
	return Name
}

func (f zstdFactory) Instance(opt map[string]interface{}) (compressor.Compressor, error) {
	cpr := zstd{}
	err := mapstructure.Decode(&opt, &cpr.conf)
	if err != nil {
		return nil, err
	}
	return &cpr, err
}

type zstd struct {
	conf ZstdConfig
}

type ZstdConfig struct {
	Level int `mapstructure:"level"`
}

func (cpr *zstd) CompressFile(ctx context.Context, r io.Reader, w io.Writer) error {
	enc, err := zskp.NewWriter(w,
		zskp.WithEncoderLevel(zskp.EncoderLevelFromZstd(cpr.conf.Level)),
	)
	if err != nil {
		return err
	}
	defer enc.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_, err = io.CopyN(enc, r, 32*1024)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
	}
	err = enc.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (cpr *zstd) Rename(origin string) string {
	return fmt.Sprintf("%s.zst", origin)
}
