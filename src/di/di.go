package di

import (
	"github.com/SirWaithaka/payments-api/src/config"
	"github.com/SirWaithaka/payments-api/src/domains/mpesa"
	"github.com/SirWaithaka/payments-api/src/domains/webhooks"
	"github.com/SirWaithaka/payments-api/src/events"
	"github.com/SirWaithaka/payments-api/src/repositories/postgres"
	"github.com/SirWaithaka/payments-api/src/services"
	"github.com/SirWaithaka/payments-api/src/storage"
)

type DI struct {
	Cfg       *config.Config
	Publisher events.Publisher

	Mpesa   mpesa.Service
	Webhook webhooks.Service
}

func New(cfg config.Config, db *storage.Database, pub events.Publisher) *DI {
	requestsRepository := postgres.NewRequestRepository(db.PG)
	webhooksRepository := postgres.NewWebhookRepository(db.PG)
	shortcodeRepository := postgres.NewShortCodeRepository(db.PG)
	mpesaPaymentsRepository := postgres.NewMpesaPaymentsRepository(db.PG)

	apiProvider := services.NewProvider(cfg, requestsRepository, webhooksRepository)

	mpesaService := mpesa.NewService(mpesaPaymentsRepository, shortcodeRepository, requestsRepository, apiProvider, pub)
	webhooksService := webhooks.NewService(webhooksRepository, mpesaService, pub)

	return &DI{
		Cfg:       &cfg,
		Publisher: pub,
		Mpesa:     mpesaService,
		Webhook:   webhooksService,
	}
}
