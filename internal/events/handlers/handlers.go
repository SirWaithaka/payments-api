package handlers

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/domains/webhooks"
	"github.com/SirWaithaka/payments-api/internal/pkg/events"
	"github.com/SirWaithaka/payments-api/internal/pkg/events/payloads"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
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
	evt := &events.Event[payloads.WebhookReceived[requests.WebhookResult]]{}

	fn := func(ctx context.Context) error {
		l := zerolog.Ctx(ctx)
		l.Info().Any(logger.LData, evt).Msg("processing webhook received event")

		if err := handler.webhook.Process(ctx, &evt.Payload.Content); err != nil {
			l.Error().Err(err).Msg("error processing webhook")
			return err
		}

		return nil
	}

	return evt, fn
}
