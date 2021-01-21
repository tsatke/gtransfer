package gtransfer

import (
	"archive/tar"
	"fmt"
	"io"

	"github.com/spf13/afero"
)

type FileWriter struct {
	underlying *tar.Writer
}

func NewFileWriter(to io.Writer) *FileWriter {
	return &FileWriter{
		underlying: tar.NewWriter(to),
	}
}

func (w *FileWriter) WriteFile(file afero.File) error {
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", file.Name(), err)
	}
	header := &tar.Header{
		Name: file.Name(),
		Size: stat.Size(),
	}
	if err := w.underlying.WriteHeader(header); err != nil {
		return fmt.Errorf("header: %w", err)
	}
	if _, err := io.Copy(w.underlying, file); err != nil {
		return fmt.Errorf("file: %w", err)
	}
	return nil
}

func (w *FileWriter) Close() error {
	return w.underlying.Close()
}
