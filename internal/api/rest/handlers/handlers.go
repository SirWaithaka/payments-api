package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// logs the error and writes status 400
func handleRequestParsingError(c *gin.Context, err error) {
	l := zerolog.Ctx(c.Request.Context())
	err = c.Error(err)
	l.Error().Err(err).Msg("error parsing request")
	c.Status(http.StatusBadRequest)
}
