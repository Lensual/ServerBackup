package backuper

import "io"

type File struct {
	Path       string
	ReadCloser io.ReadCloser
}
