package webhooks

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/events"
	pkgevents "github.com/SirWaithaka/payments-api/internal/pkg/events"
	"github.com/SirWaithaka/payments-api/internal/pkg/events/payloads"
	"github.com/SirWaithaka/payments-api/internal/pkg/events/subjects"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

func NewService(repository Repository, mpesaService mpesa.Service, publisher events.Publisher) WebhookService {
	return WebhookService{
		repository:   repository,
		mpesaService: mpesaService,
		publisher:    publisher,
	}
}

type WebhookService struct {
	repository   Repository
	mpesaService mpesa.Service
	publisher    events.Publisher
}

func (service WebhookService) Confirm(ctx context.Context, result *requests.WebhookResult) error {
	l := zerolog.Ctx(ctx)

	// TODO: Maybe validate against double webhooks before saving and publishing

	// save the webhook result
	err := service.repository.Add(ctx, result.Service.String(), result.Action, result.Bytes())
	if err != nil {
		// I think we should fail if saving fails
		return err
	}

	// publish webhook event
	payload := payloads.WebhookReceived[[]byte]{
		Action:  result.Action,
		Service: result.Service.String(),
		Content: result.Bytes(),
	}
	event := pkgevents.NewEvent(subjects.WebhookReceived, payload)
	err = service.publisher.Publish(ctx, event)
	if err != nil {
		l.Error().Err(err).Msg("error publishing event")
		return err
	}
	l.Debug().Msg("webhook event published")

	return nil

}

// Process checks if the webhook received relates to any recorded payment request, if yes,
// the webhook is parsed then used to update the payment.
func (service WebhookService) Process(ctx context.Context, result *requests.WebhookResult) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, result).Msg("processing webhook")

	// check service name on webhook and call the appropriate domain service
	if result.Service == requests.PartnerDaraja || result.Service == requests.PartnerQuikk {
		return service.mpesaService.ProcessWebhook(ctx, result)
	}

	return nil
}
