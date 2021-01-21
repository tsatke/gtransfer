package gtransfer

import (
	"os"

	"github.com/rs/zerolog"
)

var (
	log = zerolog.New(
		zerolog.ConsoleWriter{
			Out: os.Stdout,
		},
	).With().
		Timestamp().
		Logger()
)
