package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/api/rest/requests"
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
)

func NewPaymentHandlers(service payments.WalletService) PaymentHandlers {
	return PaymentHandlers{service}
}

type PaymentHandlers struct {
	service payments.WalletService
}

func (handler PaymentHandlers) Charge(c *gin.Context) {
	l := zerolog.Ctx(c.Request.Context())
	l.Debug().Msg("wallet charge request")

	var params requests.RequestWalletPayment
	if err := c.ShouldBindBodyWithJSON(&params); err != nil {
		handleRequestParsingError(c, err)
		return
	}

	payment, err := handler.service.Charge(c.Request.Context(), payments.WalletPayment{
		Type:                  "CHARGE",
		BankCode:              params.BankCode,
		ExternalAccountNumber: params.ExternalAccountID,
		Amount:                params.Amount,
		Description:           params.Description,
		TransactionID:         params.TransactionID,
		IdempotencyID:         params.IdempotencyID,
	})
	if err != nil {
		err = c.Error(err)
		return
	}
	l.Debug().Interface("payment", payment).Msg("payment")

	c.JSON(http.StatusOK, payment)
}

func (handler PaymentHandlers) Payout(c *gin.Context) {
	l := zerolog.Ctx(c.Request.Context())
	l.Debug().Msg("wallet payout request")

	var params requests.RequestWalletPayment
	if err := c.ShouldBindBodyWithJSON(&params); err != nil {
		handleRequestParsingError(c, err)
		return
	}

	payment, err := handler.service.Payout(c.Request.Context(), payments.WalletPayment{
		Type:                  "PAYOUT",
		BankCode:              params.BankCode,
		TransactionID:         params.TransactionID,
		IdempotencyID:         params.IdempotencyID,
		Amount:                params.Amount,
		Description:           params.Description,
		ExternalAccountNumber: params.ExternalAccountID,
	})
	if err != nil {
		err = c.Error(err)
		return
	}
	l.Debug().Interface("payment", payment).Msg("payment")

	c.JSON(http.StatusOK, payment)
}
