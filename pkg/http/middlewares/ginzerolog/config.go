package ginzerolog

import (
	"os"

	"github.com/rs/zerolog"
)

var (
	defaultLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()
)

type Config struct {
	// Add custom zerolog logger.
	//
	// Optional. Default: zerolog.New(os.Stderr).With().Timestamp().Logger()
	Logger *zerolog.Logger

	// Skip logging for these uri
	//
	// Optional. Default: nil
	SkipURIs []string
}

func configDefault(cfg Config) Config {

	// set logger to default if nil
	if cfg.Logger == nil {
		cfg.Logger = &defaultLogger
	}

	return cfg
}
