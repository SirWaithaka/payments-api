package request

import (
	"log"
	"os"
)

type LogLevel uint

// AtLeast returns true if this LogLevel is at least high enough to satisfy v.
func (l LogLevel) AtLeast(v LogLevel) bool {
	return l >= v
}

func (l LogLevel) Equals(v LogLevel) bool {
	return l == v
}

const (
	// LogSilent state used to disable all logging. This is the default state
	LogSilent LogLevel = iota * 0x1000

	// LogError state used to log when service requests fail
	// to build, send, validate, or unmarshal.
	LogError

	// LogDebug state can be used for debug output to inspect requests
	// made and responses received.
	LogDebug
)

// Debug Logging Sub Levels
const (
	// LogDebugWithHTTPBody state used to log HTTP request and response HTTP bodies
	// in addition to the headers and path. This should be used to see the body
	// content of requests and responses made. Will also enable LogDebug.
	LogDebugWithHTTPBody LogLevel = LogDebug | (1 << iota)

	// LogDebugWithRequestRetries state used to log when service requests will
	// be retried. This should be used to log when you want to log when service
	// requests are being retried. Will also enable LogDebug.
	LogDebugWithRequestRetries
)

type Logger interface {
	Log(...any)
}

// A LoggerFunc is a convenience type to convert a function taking a variadic
// list of arguments and wrap it so the Logger interface can be used.
type LoggerFunc func(...any)

func (f LoggerFunc) Log(args ...any) {
	f(args...)
}

// NewDefaultLogger returns a Logger which will write log messages to stdout, and
// use same formatting runes as the stdlib log.Logger
func NewDefaultLogger() Logger {
	return &defaultLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Llongfile),
	}
}

// A defaultLogger provides a minimalistic logger satisfying the Logger interface.
type defaultLogger struct {
	logger *log.Logger
}

// Log logs the parameters to the stdlib logger. See log.Println.
func (l defaultLogger) Log(args ...interface{}) {
	l.logger.Println(args...)
}
