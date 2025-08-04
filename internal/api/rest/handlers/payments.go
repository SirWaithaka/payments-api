package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/api/rest/requests"
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

func NewPaymentHandlers(service payments.Wallet) PaymentHandlers {
	return PaymentHandlers{service}
}

type PaymentHandlers struct {
	service payments.Wallet
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
		Type:                  payments.RequestTypeWalletCharge,
		BankCode:              params.BankCode,
		ExternalAccountNumber: params.ExternalAccountID,
		Amount:                params.Amount,
		Description:           params.Description,
		ClientTransactionID:   params.TransactionID,
		IdempotencyID:         params.IdempotencyID,
	})
	if err != nil {
		err = c.Error(err)
		return
	}
	l.Debug().Any(logger.LData, payment).Msg("payment")

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
		Type:                  payments.RequestTypeWalletPayout,
		BankCode:              params.BankCode,
		ClientTransactionID:   params.TransactionID,
		IdempotencyID:         params.IdempotencyID,
		Amount:                params.Amount,
		Description:           params.Description,
		ExternalAccountNumber: params.ExternalAccountID,
	})
	if err != nil {
		err = c.Error(err)
		return
	}
	l.Debug().Any(logger.LData, payment).Msg("payment")

	c.JSON(http.StatusOK, payment)
}
