package di

import (
	"github.com/SirWaithaka/payments-api/internal/config"
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	"github.com/SirWaithaka/payments-api/internal/domains/webhooks"
	"github.com/SirWaithaka/payments-api/internal/events"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/services"
	"github.com/SirWaithaka/payments-api/internal/storage"
)

type DI struct {
	Cfg       *config.Config
	Publisher events.Publisher

	Wallets payments.WalletService
	Webhook webhooks.Service
}

func New(cfg config.Config, db *storage.Database, pub events.Publisher) *DI {
	paymentsRepository := postgres.NewPaymentsRepository(db.PG)
	requestsRepository := postgres.NewRequestRepository(db.PG)
	webhooksRepository := postgres.NewWebhookRepository(db.PG)

	apiProvider := services.NewProvider(requestsRepository, webhooksRepository)

	walletsService := payments.NewWalletService(apiProvider, paymentsRepository)
	webhooksService := webhooks.NewService(webhooksRepository, requestsRepository, paymentsRepository, apiProvider, pub)

	return &DI{
		Cfg:       &cfg,
		Publisher: pub,
		Wallets:   walletsService,
		Webhook:   webhooksService,
	}
}
