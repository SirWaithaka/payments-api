package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/api/rest/requests"
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
)

func NewPaymentHandlers(service payments.Service) PaymentHandlers {
	return PaymentHandlers{service}
}

type PaymentHandlers struct {
	service payments.Service
}

func (handler PaymentHandlers) Deposit(c *gin.Context) {
	l := zerolog.Ctx(c.Request.Context())
	l.Debug().Msg("deposit request")

	var params requests.RequestPayment
	if err := c.ShouldBindBodyWithJSON(&params); err != nil {
		handleRequestParsingError(c, err)
		return
	}

	err := handler.service.Transact(c.Request.Context(), payments.Payment{
		ExternalAccountNumber: params.ExternalAccountID,
		Amount:                params.Amount,
		Description:           params.Description,
		ExternalID:            params.ExternalID,
		ExternalUID:           params.ExternalUID,
	})
	if err != nil {
		err = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}
