package di

import (
	"github.com/SirWaithaka/payments-api/internal/config"
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
)

type DI struct {
	Cfg      *config.Config
	Payments payments.Service
}

func New(cfg config.Config) *DI {

	paymentsService := payments.NewService()
	return &DI{
		Cfg:      &cfg,
		Payments: paymentsService,
	}
}
