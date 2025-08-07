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

	// create a new payment and save it
	payment := requests.Payment{
		BankCode:            req.BankCode,
		PaymentID:           ulid.Make().String(),
		ClientTransactionID: req.ClientTransactionID,
		IdempotencyID:       req.IdempotencyID,
		SourceAccountNumber: req.ExternalAccountNumber,
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
	api := service.provider.GetWalletApi(req.BankCode, req.Type)
	if api == nil {
		return requests.Payment{}, errors.New("api not configured")
	}

	// set payment id in request
	req.PaymentID = payment.PaymentID
	err = api.C2B(ctx, req)
	if err != nil {
		return requests.Payment{}, err
	}

	return payment, nil
}

func (service WalletService) Payout(ctx context.Context, req WalletPayment) (requests.Payment, error) {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, req).Msg("payout params")

	// create a new payment and save it
	payment := requests.Payment{
		BankCode:            req.BankCode,
		PaymentID:           ulid.Make().String(),
		ClientTransactionID: req.ClientTransactionID,
		IdempotencyID:       req.IdempotencyID,
		// TODO: Set Account number to internal wallet account number
		SourceAccountNumber:      "",
		DestinationAccountNumber: req.ExternalAccountNumber,
		Beneficiary:              req.Beneficiary,
		Amount:                   req.Amount,
		Description:              req.Description,
	}
	err := service.repository.AddPayment(ctx, payment)
	if err != nil {
		return requests.Payment{}, err
	}

	// get wallet api for this payment request
	api := service.provider.GetWalletApi(req.BankCode, req.Type)
	if api == nil {
		return requests.Payment{}, errors.New("api not configured")
	}

	// set payment id in request
	req.PaymentID = payment.PaymentID
	err = api.B2C(ctx, req)
	if err != nil {
		return requests.Payment{}, err
	}

	return payment, nil
}

func (service WalletService) Transfer(ctx context.Context, req WalletPayment) (requests.Payment, error) {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, req).Msg("transfer params")

	// create a new payment and save it
	payment := requests.Payment{
		BankCode:            req.BankCode,
		PaymentID:           ulid.Make().String(),
		ClientTransactionID: req.ClientTransactionID,
		IdempotencyID:       req.IdempotencyID,
		// TODO: Set Account number to internal wallet account number
		SourceAccountNumber:      "",
		DestinationAccountNumber: req.ExternalAccountNumber,
		Beneficiary:              req.Beneficiary,
		Amount:                   req.Amount,
		Description:              req.Description,
	}
	err := service.repository.AddPayment(ctx, payment)
	if err != nil {
		return requests.Payment{}, err
	}

	// get wallet api for this payment request
	api := service.provider.GetWalletApi(req.BankCode, req.Type)
	if api == nil {
		return requests.Payment{}, errors.New("api not configured")
	}

	// set payment id in request
	req.PaymentID = payment.PaymentID
	err = api.B2B(ctx, req)
	if err != nil {
		return requests.Payment{}, err
	}

	return payment, nil
}

func (service WalletService) Status(ctx context.Context, req WalletPayment) (requests.Payment, error) {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, req).Msg("status params")

	// find payment
	payment, err := service.repository.FindOnePayment(ctx, requests.OptionsFindOnePayment{
		PaymentID:           &req.PaymentID,
		ClientTransactionID: &req.ClientTransactionID,
	})
	if err != nil {
		return requests.Payment{}, err
	}

	// check if the payment lifecycle is complete
	if payment.Status.Final() {
		return payment, nil
	}

	// if payment status is not final, fetch true status from external api
	api := service.provider.GetWalletApi(payment.BankCode, RequestTypePaymentStatus)
	if api == nil {
		return requests.Payment{}, errors.New("api not configured")
	}

	// make http request to payment processor api
	if err = api.Status(ctx, payment); err != nil {
		return requests.Payment{}, err
	}

	// refetch payment, in case the status is updated synchronously from the previous call
	return service.repository.FindOnePayment(ctx, requests.OptionsFindOnePayment{PaymentID: &req.PaymentID})

}
