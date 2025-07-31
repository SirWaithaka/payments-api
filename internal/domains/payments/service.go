package payments

import (
	"context"
	"errors"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

func NewWalletService(provider Provider, repository Repository) WalletService {
	return WalletService{provider: provider, repository: repository}
}

type WalletService struct {
	provider   Provider
	repository Repository
}

func (service WalletService) Charge(ctx context.Context, req WalletPayment) (requests.Payment, error) {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, req).Msg("charge params")

	// create new payment and save it
	payment := requests.Payment{
		BankCode:            req.BankCode,
		PaymentID:           ulid.Make().String(),
		ClientTransactionID: req.ClientTransactionID,
		IdempotencyID:       req.IdempotencyID,
		SourceAccountNumber: req.DestinationAccountNumber,
		// TODO: Set Account number to MPESA short
		DestinationAccountNumber: "",
		Beneficiary:              req.Beneficiary,
		Amount:                   req.Amount,
		Description:              req.Description,
	}
	err := service.repository.AddPayment(ctx, payment)
	if err != nil {
		return requests.Payment{}, err
	}

	// get wallet api for this payment request
	api := service.provider.GetWalletApi(req)
	if api == nil {
		return requests.Payment{}, errors.New("api not configured")
	}

	err = api.C2B(ctx, req)
	if err != nil {
		return requests.Payment{}, err
	}

	return payment, nil
}

func (service WalletService) Payout(ctx context.Context, req WalletPayment) (requests.Payment, error) {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, req).Msg("payout params")

	// create new payment and save it
	payment := requests.Payment{}
	err := service.repository.AddPayment(ctx, payment)
	if err != nil {
		return requests.Payment{}, err
	}

	// get wallet api for this payment request
	api := service.provider.GetWalletApi(req)
	if api == nil {
		return requests.Payment{}, errors.New("api not configured")
	}

	err = api.B2C(ctx, req)
	if err != nil {
		return requests.Payment{}, err
	}

	return payment, nil
}
