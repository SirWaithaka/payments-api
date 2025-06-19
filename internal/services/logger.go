package services

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/request"
)

func NewLogger(logger *zerolog.Logger) Logger {
	return Logger{logger: logger, level: request.LogError}
}

type Logger struct {
	logger *zerolog.Logger
	level  request.LogLevel
}

func (l Logger) Log(v ...any) {
	msg := strings.Replace(fmt.Sprint(v...), "\n", " ", -1)
	switch {
	case l.level == request.LogSilent:
		return
	case l.level == request.LogError:
		l.logger.Error().Msg(msg)
	case l.level.AtLeast(request.LogDebug):
		l.logger.Debug().Msg(msg)
	default:
		return
	}
}
