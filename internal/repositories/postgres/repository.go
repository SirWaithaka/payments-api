package postgres

import (
	"context"

	"github.com/SirWaithaka/payments-api/internal/domains/payments"
)

func NewRepository() Repository {
	return Repository{}
}

type Repository struct{}

func (repo Repository) AddPayment(ctx context.Context, payment payments.Payment) error {
	return nil
}

func (repo Repository) FindOnePayment(ctx context.Context, opts payments.OptionsFindOnePayment) (payments.Payment, error) {
	return payments.Payment{}, nil
}
