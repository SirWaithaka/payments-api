package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/api/rest/requests"
)

type PaymentHandlers struct{}

func (handler PaymentHandlers) Deposit(c *gin.Context) {
	l := zerolog.Ctx(c)
	l.Info().Msg("deposit request")

	var params requests.RequestPayment
	if err := c.ShouldBindJSON(&params); err != nil {
		err = c.Error(err)
		l.Error().Err(err).Msg("error parsing request")
		return
	}

	c.Status(http.StatusOK)
}
