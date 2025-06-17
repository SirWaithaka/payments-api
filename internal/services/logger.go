package services

import (
	"fmt"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/request"
)

type Logger struct {
	logger *zerolog.Logger
	level  request.LogLevel
}

func (l Logger) Log(msg ...any) {
	switch {
	case l.level == request.LogSilent:
		return
	case l.level.AtLeast(request.LogDebug):
		l.logger.Debug().Msg(fmt.Sprint(msg...))
	case l.level == request.LogError:
		l.logger.Error().Msg(fmt.Sprint(msg...))
	default:
		return
	}
}
