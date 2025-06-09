package middlewares

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
)

// ErrorHandler provides a customer error handling mechanism for gin framework
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		l := zerolog.Ctx(c)
		l.Info().Msg("handling error")

		c.Next()

		if len(c.Errors) < 1 {
			return
		}

		l.Error().Any("errors", c.Errors).Msg("gin errors")

		// pick one error from c.Errors
		err := c.Errors.Last()

		// check if err.Err is a postgres error
		e := postgres.Error{}
		if errors.Is(err.Err, e) {
			if e.NotFound() {
				c.AbortWithStatus(http.StatusNotFound)
			}
			return
		}

		// default for any other type of error
		c.AbortWithStatus(http.StatusInternalServerError)

	}
}
