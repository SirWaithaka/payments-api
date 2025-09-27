package services

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/gorequest"
)

func NewLogger(logger *zerolog.Logger, lvl gorequest.LogLevel) Logger {
	return Logger{logger: logger, level: lvl}
}

type Logger struct {
	logger *zerolog.Logger
	level  gorequest.LogLevel
}

func (l Logger) Log(v ...any) {
	msg := strings.Replace(fmt.Sprint(v...), "\n", " ", -1)
	switch {
	case l.level == gorequest.LogSilent:
		return
	case l.level == gorequest.LogError:
		l.logger.Error().Msg(msg)
	case l.level.AtLeast(gorequest.LogDebug):
		l.logger.Debug().Msg(msg)
	default:
		return
	}
}
