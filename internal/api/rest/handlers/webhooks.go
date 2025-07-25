package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/domains/webhooks"
)

func NewWebhookHandlers(processor webhooks.WebhookProcessor) WebhookHandlers {
	return WebhookHandlers{
		webhook: processor,
	}
}

type WebhookHandlers struct {
	webhook webhooks.WebhookProcessor
}

func (handler WebhookHandlers) Daraja(c *gin.Context) {
	l := zerolog.Ctx(c.Request.Context())
	l.Info().Msg("daraja webhook received")

	// get path param
	action := c.Param("action")

	// TODO: Make partner argument into type
	err := handler.webhook.Process(c.Request.Context(), webhooks.NewWebhookResult("daraja", action, c.Request.Body))
	if err != nil {
		l.Warn().Err(err).Msg("error processing webhook")
		c.String(http.StatusAccepted, "accepted")
		return
	}

	c.String(http.StatusOK, "OK")

}
