package payments

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

func NewWalletService(provider Provider, repository Repository) WalletService {
	return WalletService{provider: provider, repository: repository}
}

type WalletService struct {
	provider   Provider
	repository Repository
}

func (service WalletService) Charge(ctx context.Context, req WalletPayment) (Payment, error) {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, req).Msg("charge params")

	// create new payment and save it
	payment := Payment{}
	err := service.repository.AddPayment(ctx, payment)
	if err != nil {
		return Payment{}, err
	}

	// get wallet api for this payment request
	api := service.provider.GetWalletApi(req)
	if api == nil {
		return Payment{}, errors.New("api not configured")
	}

	err = api.C2B(ctx, req)
	if err != nil {
		return Payment{}, err
	}

	return payment, nil
}
