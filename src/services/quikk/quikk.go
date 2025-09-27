package quikk

import (
	"context"
	"strconv"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/pkg/logger"
	"github.com/SirWaithaka/payments-api/src/domains/mpesa"
	"github.com/SirWaithaka/payments-api/src/domains/requests"
	"github.com/SirWaithaka/payments-api/src/services/hooks"
	"github.com/SirWaithaka/payments/quikk"

	"github.com/SirWaithaka/gorequest"
)

const (
	serviceName = requests.PartnerQuikk
)

type ResponseDefault quikk.ResponseDefault

func (response ResponseDefault) ExternalID() string { return response.Data.ID }

// QUIKK MPESA API SERVICE

// NewQuikkApi creates a new instance of QuikkApi
func NewQuikkApi(client *quikk.Client, shortcode mpesa.ShortCode, repo requests.Repository) QuikkApi {
	return QuikkApi{client: client, shortcode: shortcode, requestRepo: repo}
}

type QuikkApi struct {
	client      *quikk.Client
	shortcode   mpesa.ShortCode
	requestRepo requests.Repository
}

func (api QuikkApi) C2B(ctx context.Context, paymentID string, payment mpesa.PaymentRequest) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling c2b payment")

	amount, err := strconv.ParseFloat(payment.Amount, 64)
	if err != nil {
		l.Error().Err(err).Msg("error parsing amount")
		// TODO: return a customer payment error
		return err
	}

	payload := quikk.RequestCharge{
		Amount:       amount,
		CustomerNo:   payment.ExternalAccountNumber,
		Reference:    paymentID,
		CustomerType: mpesa.AccountTypeMSISDN.String(),
		ShortCode:    api.shortcode.ShortCode,
		PostedAt:     time.Now().Format(time.RFC3339),
	}
	l.Debug().Any(logger.LData, payload).Msg("request payload")

	// initialize request recorder
	recorder := hooks.NewRequestRecorder(api.requestRepo)

	// create an instance of request and add the request recorder hook
	requestID := xid.New().String()
	out := &ResponseDefault{}
	req, _ := api.client.ChargeRequest(payload, requestID, gorequest.WithServiceName(serviceName.String()))
	req.WithContext(ctx)
	req.Data = out
	req.Hooks.Send.PushFrontHook(recorder.RecordRequest(paymentID, requestID))
	req.Hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

	if err = req.Send(); err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, out).Msg("c2b response")

	return nil
}

func (api QuikkApi) B2C(ctx context.Context, paymentID string, payment mpesa.PaymentRequest) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling b2c payment")

	amount, err := strconv.ParseFloat(payment.Amount, 64)
	if err != nil {
		l.Error().Err(err).Msg("error parsing amount")
		// TODO: return a customer payment error
		return err
	}

	payload := quikk.RequestPayout{
		Amount:        amount,
		RecipientNo:   payment.ExternalAccountNumber,
		RecipientType: mpesa.AccountTypeMSISDN.String(),
		ShortCode:     api.shortcode.ShortCode,
		PostedAt:      time.Now().Format(time.RFC3339),
	}
	l.Debug().Any(logger.LData, payload).Msg("request payload")

	// initialize request recorder
	recorder := hooks.NewRequestRecorder(api.requestRepo)

	// create an instance of request and add the request recorder hook
	requestID := xid.New().String()
	out := &ResponseDefault{}
	req, _ := api.client.PayoutRequest(payload, requestID, gorequest.WithServiceName(serviceName.String()))
	req.WithContext(ctx)
	req.Data = out
	req.Hooks.Send.PushFrontHook(recorder.RecordRequest(paymentID, requestID))
	req.Hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

	if err = req.Send(); err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, out).Msg("b2c response")

	return nil

}

func (api QuikkApi) B2B(ctx context.Context, paymentID string, payment mpesa.PaymentRequest) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling b2b payment")

	amount, err := strconv.ParseFloat(payment.Amount, 64)
	if err != nil {
		l.Error().Err(err).Msg("error parsing amount")
		// TODO: return a customer payment error
		return err
	}

	payload := quikk.RequestTransfer{
		Amount:            amount,
		RecipientNo:       payment.ExternalAccountNumber,
		AccountNo:         payment.Beneficiary,
		ShortCode:         api.shortcode.ShortCode,
		RecipientType:     "short_code",
		RecipientCategory: payment.ExternalAccountType.String(),
		PostedAt:          time.Now().Format(time.RFC3339),
	}
	l.Debug().Any(logger.LData, payload).Msg("request payload")

	// initialize request recorder
	recorder := hooks.NewRequestRecorder(api.requestRepo)

	// create an instance of request and add the request recorder hook
	requestID := xid.New().String()
	out := &ResponseDefault{}
	req, _ := api.client.TransferRequest(payload, requestID, gorequest.WithServiceName(serviceName.String()))
	req.WithContext(ctx)
	req.Data = out
	req.Hooks.Send.PushFrontHook(recorder.RecordRequest(paymentID, requestID))
	req.Hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

	if err = req.Send(); err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, out).Msg("b2b response")

	return nil

}

func (api QuikkApi) Status(ctx context.Context, payment mpesa.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling transaction status")

	payload := quikk.RequestTransactionStatus{
		ShortCode: api.shortcode.ShortCode,
	}

	// default using the mpesa ref to check status
	if payment.PaymentReference != "" {
		payload.Reference = payment.PaymentReference
		payload.ReferenceType = "txn_id"
	} else {
		// if not using mpesa ref, we need to pull the latest request
		// made for payment to check the response id or the resource id
		req, err := api.requestRepo.FindOne(ctx, requests.OptionsFindRequest{PaymentID: &payment.PaymentID})
		if err != nil {
			l.Error().Err(err).Msg("error fetching request")
			return err
		}

		// use response id
		payload.Reference = req.ExternalID
		payload.ReferenceType = "response_id"
	}
	l.Debug().Any(logger.LData, payload).Msg("request payload")

	requestID := xid.New().String()
	res, err := api.client.TransactionSearch(ctx, payload, requestID)
	if err != nil {
		l.Error().Err(err).Msg("client error")
		return err
	}
	l.Debug().Any(logger.LData, res).Msg("transaction status")

	return nil
}
