package gtransfer

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

type Client struct {
	addr string
}

func Dial(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

func (c *Client) DownloadInto(fs afero.Fs) error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	decompressor, err := gzip.NewReader(conn)
	if err != nil {
		return fmt.Errorf("new gzip reader: %w", err)
	}
	tarReader := tar.NewReader(decompressor)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			log.Info().
				Msg("download done")
			break
		} else if err != nil {
			log.Error().
				Err(err).
				Msg("read header")
			break
		}

		if err := fs.MkdirAll(filepath.Dir(header.Name), 0700); err != nil {
			return fmt.Errorf("mkdirall %s: %w", filepath.Dir(header.Name), err)
		}

		file, err := fs.OpenFile(header.Name, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		if err != nil {
			return fmt.Errorf("create %s: %w", header.Name, err)
		}

		start := time.Now()
		if _, err := io.CopyN(file, tarReader, header.Size); err != nil {
			return fmt.Errorf("read %s: %w", file.Name(), err)
		}
		duration := time.Since(start)
		log.Info().
			Str("path", header.Name).
			Str("size", humanReadableByteSize(header.Size)).
			Stringer("took", duration).
			Msg("receive file")
	}

	_ = conn.Close()

	return nil
}
