package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
)

// ErrorHandler provides a customer error handling mechanism for gin framework
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		l := zerolog.Ctx(c.Request.Context())
		l.Debug().Msg("handling error")

		c.Next()

		if len(c.Errors) < 1 {
			return
		}
		l.Error().Any("errors", c.Errors).Msg("gin errors")

		// pick one error from c.Errors
		err := c.Errors.Last()

		// set header for error responses format
		c.Header("content-type", "application/json")

		// check headers are not written and status code is not default 200
		if c.Writer.Size() < 1 && c.Writer.Status() != http.StatusOK {
			c.AbortWithStatusJSON(c.Writer.Status(), gin.H{
				"error": err.Error(),
			})
			return
		}

		switch e := err.Err.(type) {
		case postgres.Error:
			if e.NotFound() {
				c.AbortWithStatus(http.StatusNotFound)
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		case validator.ValidationErrors:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "validation failed",
				"details": e,
			})
		case interface{ Temporary() bool }:
			if e.Temporary() {
				c.AbortWithStatus(http.StatusServiceUnavailable)
			}

		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": e.Error(),
			})
		}

	}
}
