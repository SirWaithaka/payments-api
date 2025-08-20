package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/api/rest/requests"
	"github.com/SirWaithaka/payments-api/internal/api/rest/responses"
	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

func NewMpesaHandlers(service mpesa.Service) MpesaHandlers {
	return MpesaHandlers{service}
}

type MpesaHandlers struct {
	service mpesa.Service
}

func (handler MpesaHandlers) Charge(c *gin.Context) {
	l := zerolog.Ctx(c.Request.Context())
	l.Debug().Msg("mpesa charge request")

	var params requests.RequestMpesaPayment
	if err := c.ShouldBindBodyWithJSON(&params); err != nil {
		handleRequestParsingError(c, err)
		return
	}

	payment, err := handler.service.Charge(c.Request.Context(), mpesa.PaymentRequest{
		IdempotencyID:         params.IdempotencyID,
		ClientTransactionID:   params.TransactionID,
		Amount:                params.Amount,
		ExternalAccountNumber: params.ExternalAccountID,
		Description:           params.Description,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	l.Debug().Any(logger.LData, payment).Msg("payment")

	c.JSON(http.StatusOK, responses.MpesaPaymentResponse{
		PaymentID:     payment.PaymentID,
		TransactionID: payment.ClientTransactionID,
		Status:        payment.Status.String(),
	})
}

func (handler MpesaHandlers) Payout(c *gin.Context) {
	l := zerolog.Ctx(c.Request.Context())
	l.Debug().Msg("mpesa payout request")

	var params requests.RequestMpesaPayment
	if err := c.ShouldBindBodyWithJSON(&params); err != nil {
		handleRequestParsingError(c, err)
		return
	}

	payment, err := handler.service.Payout(c.Request.Context(), mpesa.PaymentRequest{
		IdempotencyID:         params.IdempotencyID,
		ClientTransactionID:   params.TransactionID,
		Amount:                params.Amount,
		ExternalAccountNumber: params.ExternalAccountID,
		Description:           params.Description,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	l.Debug().Any(logger.LData, payment).Msg("payment")

	c.JSON(http.StatusOK, responses.MpesaPaymentResponse{
		PaymentID:     payment.PaymentID,
		TransactionID: payment.ClientTransactionID,
		Status:        payment.Status.String(),
	})
}

func (handler MpesaHandlers) Transfer(c *gin.Context) {
	l := zerolog.Ctx(c.Request.Context())
	l.Debug().Msg("mpesa transfer request")

	var params requests.RequestMpesaPayment
	if err := c.ShouldBindBodyWithJSON(&params); err != nil {
		handleRequestParsingError(c, err)
		return
	}

	payment, err := handler.service.Transfer(c.Request.Context(), mpesa.PaymentRequest{
		IdempotencyID:         params.IdempotencyID,
		ClientTransactionID:   params.TransactionID,
		Amount:                params.Amount,
		ExternalAccountType:   mpesa.ToAccountType(params.ExternalAccountType),
		ExternalAccountNumber: params.ExternalAccountID,
		Beneficiary:           params.Beneficiary,
		Description:           params.Description,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	l.Debug().Any(logger.LData, payment).Msg("payment")

	c.JSON(http.StatusOK, responses.MpesaPaymentResponse{
		PaymentID:     payment.PaymentID,
		TransactionID: payment.ClientTransactionID,
		Status:        payment.Status.String(),
	})

}

func (handler MpesaHandlers) PaymentStatus(c *gin.Context) {
	l := zerolog.Ctx(c.Request.Context())
	l.Debug().Msg("mpesa payment status request")

	var params requests.RequestMpesaPaymentStatus
	if err := c.ShouldBindBodyWithJSON(&params); err != nil {
		handleRequestParsingError(c, err)
		return
	}

	opts := mpesa.OptionsFindPayment{}
	if params.PaymentID != "" {
		opts.PaymentID = &params.PaymentID
	}
	if params.TransactionID != "" {
		opts.ClientTransactionID = &params.TransactionID
	}
	if params.PaymentReference != "" {
		opts.PaymentReference = &params.PaymentReference
	}

	payment, err := handler.service.Status(c.Request.Context(), opts)
	if err != nil {
		_ = c.Error(err)
		return
	}
	l.Debug().Any(logger.LData, payment).Msg("payment")

	c.JSON(http.StatusOK, payment)
}
