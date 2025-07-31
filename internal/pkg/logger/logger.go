package logger

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

const (
	LStage     = "stage"
	LData      = "data"
	LUserID    = "userId"
	LUrl       = "url"
	LMethod    = "method"
	LService   = "service"
	LContext   = "context"
	LRequestID = "requestID"
)

const (
	LModeSilent = "SILENT"
	LModeTrace  = "TRACE"
	LModeInfo   = "INFO"
	LModeWarn   = "WARN"
	LModeError  = "ERROR"
	LModeDebug  = "DEBUG"
)

type Config struct {
	LogMode string
	Service string
}

// DataParams can be used to pass data values to logger as interface
// e.g. zerolog.Info().Interface(LData, DataParams{"a": a, "b": b})
type DataParams map[string]interface{}

var once sync.Once

var (
	log *zerolog.Logger
	//DefaultContextLogger *zerolog.Logger
)

// New accepts configurations on logger and creates a new logger instance
// If no LogMode is set, it will default to debug level
func New(cfg *Config) zerolog.Logger {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339Nano

	if cfg == nil {
		return defaultLogger()
	}

	// configure logger
	// default output
	var output io.Writer = os.Stdout
	var level zerolog.Level
	switch cfg.LogMode {
	case LModeSilent:
		level = zerolog.Disabled
	case LModeTrace:
		level = zerolog.TraceLevel
	case LModeInfo:
		level = zerolog.InfoLevel
	case LModeWarn:
		level = zerolog.WarnLevel
	case LModeError:
		level = zerolog.ErrorLevel
	case LModeDebug:
		level = zerolog.DebugLevel
	default:
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		level = zerolog.DebugLevel
	}

	if cfg.Service == "" {
		cfg.Service = "generic"
	}

	l := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Str(LService, cfg.Service).
		Str(LStage, "processing").
		Caller().
		Logger()

	return l
}

// Get will return copy of global logger
func Get() zerolog.Logger {
	once.Do(func() {
		if log == nil {
			l := New(nil)
			log = &l
		}
	})
	return *log
}

// SetDefaultLogger will override global logger defined
func SetDefaultLogger(logger zerolog.Logger) {
	log = &logger
}

// Default will create a logger with sensible defaults
func defaultLogger() zerolog.Logger {
	return zerolog.New(os.Stdout).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Str(LStage, "processing").
		Caller().
		Logger()

}
