package gtransfer

import (
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/afero"
)

type Server struct {
	addr      string
	listener  net.Listener
	fs        afero.Fs
	listening chan struct{}
}

func NewServer(addr string, fs afero.Fs) *Server {
	return &Server{
		addr:      addr,
		fs:        fs,
		listening: make(chan struct{}),
	}
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) Listening() <-chan struct{} {
	return s.listening
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.listener = lis
	log.Info().
		Str("addr", s.Addr()).
		Msg("listening")

	close(s.listening)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}
		go s.handleConn(conn)
	}
}

func (s *Server) IP() net.IP {
	return s.listener.Addr().(*net.TCPAddr).IP
}

func (s *Server) Stop() {
	_ = s.listener.Close()
}

func (s *Server) handleConn(conn net.Conn) {
	compressor := gzip.NewWriter(conn)
	fileWriter := NewFileWriter(compressor)
	var failedErr []error
	totalStart := time.Now()
	if err := afero.Walk(s.fs, "", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		file, openErr := s.fs.OpenFile(path, os.O_RDONLY, 0600)
		if openErr != nil {
			failedErr = append(failedErr, fmt.Errorf("open %s: %w", info.Name(), err))
			return nil
		}
		start := time.Now()
		err = fileWriter.WriteFile(file)
		if err != nil {
			return fmt.Errorf("write %s: %w", info.Name(), err)
		}
		duration := time.Since(start)
		log.Info().
			Str("path", path).
			Stringer("took", duration).
			Str("size", humanReadableByteSize(info.Size())).
			Msg("write file")

		_ = file.Close()

		return nil
	}); err != nil {
		log.Error().
			Err(err).
			IPAddr("client", conn.RemoteAddr().(*net.TCPAddr).IP).
			Msg("error while sending file system")
	}

	for i, err := range failedErr {
		log.Error().
			Err(err).
			Msgf("error %d", i)
	}

	log.Info().
		IPAddr("client", conn.RemoteAddr().(*net.TCPAddr).IP).
		Stringer("took", time.Since(totalStart)).
		Msg("upload done")
	_ = fileWriter.Close()
	_ = compressor.Close()
	_ = conn.Close()
}
