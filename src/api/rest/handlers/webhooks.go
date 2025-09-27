package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/src/domains/requests"
	"github.com/SirWaithaka/payments-api/src/domains/webhooks"
)

func NewWebhookHandlers(service webhooks.Service) WebhookHandlers {
	return WebhookHandlers{
		service: service,
	}
}

type WebhookHandlers struct {
	service webhooks.Service
}

func (handler WebhookHandlers) Daraja(c *gin.Context) {
	l := zerolog.Ctx(c.Request.Context())
	l.Info().Msg("daraja webhook received")

	// get path param
	action := c.Param("action")

	// TODO: Make partner argument into type
	err := handler.service.Confirm(c.Request.Context(), requests.NewWebhookResult("daraja", action, c.Request.Body))
	if err != nil {
		l.Warn().Err(err).Msg("error processing webhook")
		c.String(http.StatusAccepted, "accepted")
		return
	}

	c.String(http.StatusOK, "OK")

}

func (handler WebhookHandlers) QuikkMpesa(c *gin.Context) {
	l := zerolog.Ctx(c.Request.Context())
	l.Info().Msg("quikk mpesa webhook received")

	// get path param
	action := c.Param("action")

	err := handler.service.Confirm(c.Request.Context(), requests.NewWebhookResult("quikk", action, c.Request.Body))
	if err != nil {
		l.Warn().Err(err).Msg("error processing webhook")
		c.String(http.StatusAccepted, "accepted")
		return
	}

	c.String(http.StatusOK, "OK")
}
