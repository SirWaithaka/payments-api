package di

import (
	"github.com/SirWaithaka/payments-api/internal/config"
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/services"
)

type DI struct {
	Cfg *config.Config

	Wallets payments.WalletService
}

func New(cfg config.Config) *DI {
	paymentsRepository := postgres.NewRepository()

	apiProvider := services.NewProvider()

	walletsService := payments.NewWalletService(apiProvider, paymentsRepository)

	return &DI{
		Cfg:     &cfg,
		Wallets: walletsService,
	}
}
