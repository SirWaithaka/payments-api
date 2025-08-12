package quikk

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/clients/quikk"
	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
	"github.com/SirWaithaka/payments-api/internal/services/hooks"
	"github.com/SirWaithaka/payments-api/request"
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
		CustomerType: "msisdn",
		ShortCode:    api.shortcode.ShortCode,
		PostedAt:     time.Now().Format(time.RFC3339),
	}
	l.Debug().Any(logger.LData, payload).Msg("request payload")

	// initialize request recorder
	recorder := hooks.NewRequestRecorder(api.requestRepo)

	// create an instance of request and add the request recorder hook
	requestID := xid.New().String()
	out := &quikk.ResponseDefault{}
	req, _ := api.client.ChargeRequest(payload, requestID, request.WithServiceName(serviceName.String()))
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

	return errors.New("not implemented")
}

func (api QuikkApi) B2B(ctx context.Context, paymentID string, payment mpesa.PaymentRequest) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling b2b payment")

	return errors.New("not implemented")
}

func (api QuikkApi) Status(ctx context.Context, payment mpesa.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("handling transaction status")

	return errors.New("not implemented")
}
