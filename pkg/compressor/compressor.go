package compressor

import (
	"context"
	"fmt"
	"io"
)

type Compressor interface {
	CompressFile(ctx context.Context, r io.Reader, w io.Writer) error
	Rename(origin string) string
}

type CompressorFactory interface {
	Name() string
	Instance(opt map[string]interface{}) (Compressor, error)
}

var registered = make(map[string]CompressorFactory)

func Register(f CompressorFactory) {
	if f == nil {
		panic("cannot register a nil Compressor")
	}
	if f.Name() == "" {
		panic("cannot register Compressor with empty string result for Name()")
	}
	registered[f.Name()] = f
}

func New(driver string, opt map[string]interface{}) (Compressor, error) {
	f, ok := registered[driver]
	if !ok {
		return nil, fmt.Errorf("compressor driver \"%s\" not found", driver)
	}
	return f.Instance(opt)
}
