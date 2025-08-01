package webhooks

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/events"
	pkgevents "github.com/SirWaithaka/payments-api/internal/pkg/events"
	"github.com/SirWaithaka/payments-api/internal/pkg/events/payloads"
	"github.com/SirWaithaka/payments-api/internal/pkg/events/subjects"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

func NewService(repository Repository, requestsRepo requests.Repository, paymentsRepo payments.Repository, provider requests.Provider, publisher events.Publisher) WebhookService {
	return WebhookService{
		repository:   repository,
		requestsRepo: requestsRepo,
		paymentsRepo: paymentsRepo,
		provider:     provider,
		publisher:    publisher,
	}
}

type WebhookService struct {
	repository   Repository
	requestsRepo requests.Repository
	paymentsRepo payments.Repository
	provider     requests.Provider
	publisher    events.Publisher
}

func (service WebhookService) Confirm(ctx context.Context, result *requests.WebhookResult) error {
	l := zerolog.Ctx(ctx)

	// TODO: Maybe validate against double webhooks before saving and publishing

	// save the webhook result
	err := service.repository.Add(ctx, result.Service, result.Action, result.Bytes())
	if err != nil {
		// I think we should fail if saving fails
		return err
	}

	// publish webhook event
	payload := payloads.WebhookReceived[[]byte]{
		Action:  result.Action,
		Service: result.Service,
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

	// get the specific client that should process the service's webhook
	client := service.provider.GetWebhookClient(result.Service)

	// use client to get necessary data to update payment
	opts, err := client.Process(ctx, result)
	if err != nil {
		// if error, do nothing and return
		l.Warn().Err(err).Msg("error transforming webhook")
		return err
	}

	// check if the webhook is tied to a request
	var in interface{ ExternalID() string }
	var ok bool
	if in, ok = result.Data.(interface{ ExternalID() string }); !ok {
		// TODO: do something else with webhook if its not registered
		l.Warn().Msg("webhook not registered")
		return nil
	}

	// fetch request
	extID := in.ExternalID()
	req, err := service.requestsRepo.FindOneRequest(ctx, requests.OptionsFindOneRequest{ExternalID: &extID})
	if err != nil {
		// TODO: do something if error is not found
		l.Error().Err(err).Msg("error fetching request")
		return err
	}

	// check if the request has a payment record attached, then update the payment
	if req.PaymentID == "" {
		l.Info().Msg("no payment details attached to request")
		return nil
	}

	// update payment record
	err = service.paymentsRepo.UpdatePayment(ctx, req.PaymentID, opts)
	if err != nil {
		return err
	}

	// publish webhook event
	event := pkgevents.NewEvent(subjects.PaymentCompleted, payloads.PaymentStatusUpdated{
		PaymentID: req.PaymentID,
	})
	err = service.publisher.Publish(ctx, event)
	if err != nil {
		l.Error().Err(err).Msg("error publishing event")
		return err
	}
	l.Debug().Msg("webhook event published")

	return nil
}
