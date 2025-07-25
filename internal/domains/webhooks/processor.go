package webhooks

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/events"
	pkgevents "github.com/SirWaithaka/payments-api/internal/pkg/events"
	"github.com/SirWaithaka/payments-api/internal/pkg/events/payloads"
	"github.com/SirWaithaka/payments-api/internal/pkg/events/subjects"
)

func NewProcessor(provider Provider, publisher events.Publisher) Processor {
	return Processor{provider, publisher}
}

type Processor struct {
	provider  Provider
	publisher events.Publisher
}

func (processor Processor) Process(ctx context.Context, result *WebhookResult) error {
	l := zerolog.Ctx(ctx)
	l.Info().Msg("processing webhook")

	client := processor.provider.GetWebhookClient(result.Service)

	// immediately saves webhook to storage
	err := client.Process(ctx, result)
	if err != nil || result.Data == nil {
		// if error we do nothing and just return
		err = errors.Join(errors.New("error processing webhook"), err)
		l.Warn().Err(err).Msg("error transforming webhook")
		return err
	}

	// publish webhook event
	event := pkgevents.NewEvent(subjects.WebhookReceived, payloads.WebhookReceived[WebhookResult]{Content: *result})
	err = processor.publisher.Publish(ctx, event)
	if err != nil {
		l.Error().Err(err).Msg("error publishing event")
		return err
	}
	l.Debug().Msg("webhook event published")

	return nil
}
