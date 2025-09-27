package handlers

import (
	"bytes"
	"context"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/pkg/events"
	"github.com/SirWaithaka/payments-api/pkg/events/payloads"
	"github.com/SirWaithaka/payments-api/pkg/logger"
	"github.com/SirWaithaka/payments-api/src/domains/requests"
	"github.com/SirWaithaka/payments-api/src/domains/webhooks"
)

func NewHandler(service webhooks.Service) Handler {
	return Handler{
		webhook: service,
	}
}

// Handler that declares methods that handle events
type Handler struct {
	webhook webhooks.Service
}

func (handler Handler) PaymentCompleted() (events.EventMessage, func(ctx context.Context) error) {
	evt := &events.Event[payloads.PaymentCompleted]{}

	fn := func(ctx context.Context) error {
		l := zerolog.Ctx(ctx)
		l.Info().Any(logger.LData, evt).Msg("processing payment completed event")
		return nil
	}

	return evt, fn
}

func (handler Handler) WebhookReceived() (events.EventMessage, func(ctx context.Context) error) {
	evt := &events.Event[payloads.WebhookReceived[payloads.Bytes]]{}

	fn := func(ctx context.Context) error {
		l := zerolog.Ctx(ctx)
		l.Info().Msgf("webhook received event: %s - %s", evt.Payload.Service, evt.Payload.Action)

		result := requests.NewWebhookResult(evt.Payload.Service, evt.Payload.Action, bytes.NewReader(evt.Payload.Content))
		if err := handler.webhook.Process(ctx, result); err != nil {
			l.Error().Err(err).Msg("error processing webhook")
			return err
		}

		return nil
	}

	return evt, fn
}
