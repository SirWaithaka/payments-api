package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

func NewWebhookHandlers() WebhookHandlers {
	return WebhookHandlers{}
}

type WebhookHandlers struct{}

func (handler WebhookHandlers) Daraja(c *gin.Context) {
	l := zerolog.Ctx(c.Request.Context())
	l.Info().Msg("daraja webhook received")

	// get path param
	action := c.Param("action")

	l.Debug().Str(logger.LData, action).Msg("daraja action received")

}
