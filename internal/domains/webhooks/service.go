package webhooks

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/events"
	pkgevents "github.com/SirWaithaka/payments-api/internal/pkg/events"
	"github.com/SirWaithaka/payments-api/internal/pkg/events/payloads"
	"github.com/SirWaithaka/payments-api/internal/pkg/events/subjects"
)

func NewService(repository Repository, publisher events.Publisher) WebhookService {
	return WebhookService{repository, publisher}
}

type WebhookService struct {
	repository Repository
	publisher  events.Publisher
}

func (service WebhookService) Confirm(ctx context.Context, result *requests.WebhookResult) error {
	l := zerolog.Ctx(ctx)

	// TODO: Maybe we can validate against double webhooks before saving and publishing

	// save the webhook result
	err := service.repository.Add(ctx, result.Service, result.Action, result.Bytes())
	if err != nil {
		// I think we should fail if saving fails
		return err
	}

	// publish webhook event
	event := pkgevents.NewEvent(subjects.WebhookReceived, payloads.WebhookReceived[requests.WebhookResult]{Content: *result})
	err = service.publisher.Publish(ctx, event)
	if err != nil {
		l.Error().Err(err).Msg("error publishing event")
		return err
	}
	l.Debug().Msg("webhook event published")

	return nil

}
