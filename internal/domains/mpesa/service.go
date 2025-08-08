package mpesa

import (
	"context"
	"errors"

	"github.com/oklog/ulid/v2"

	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/pkg/types"
)

func NewMpesaService(repository Repository, shortCodeRepository ShortCodeRepository, provider Provider) MpesaService {
	return MpesaService{
		repository:          repository,
		shortCodeRepository: shortCodeRepository,
		provider:            provider,
	}
}

type MpesaService struct {
	repository          Repository
	shortCodeRepository ShortCodeRepository
	provider            Provider
}

func (service MpesaService) Transfer(ctx context.Context, req PaymentRequest) (Payment, error) {

	// get shortcode details for this payment type
	shortcodes, err := service.shortCodeRepository.FindMany(ctx, OptionsFindShortCodes{
		Service: types.Pointer(requests.PartnerDaraja),
		Type:    types.Pointer(PaymentTypeWalletTransfer.String()),
	})
	if err != nil {
		return Payment{}, err
	}

	if len(shortcodes) == 0 {
		return Payment{}, errors.New("no shortcodes configured for payment type")
	}

	shortcode := shortcodes[0]

	// create a new payment and save it
	payment := Payment{
		PaymentID:           ulid.Make().String(),
		Type:                PaymentTypeWalletTransfer,
		ClientTransactionID: req.ClientTransactionID,
		IdempotencyID:       req.IdempotencyID,
		Amount:              req.Amount,
		SourceAccountNumber: req.ExternalAccountNumber,
		Beneficiary:         req.Beneficiary,
		Description:         req.Description,
		ShortCodeID:         shortcode.ShortCodeID,
	}

	// saving will fail if payment with the same idempotency id already exists
	err = service.repository.Add(ctx, payment)
	if err != nil {
		return Payment{}, err
	}

	// get client api for this payment request
	api := service.provider.GetMpesaApi(shortcode)
	if api == nil {
		return Payment{}, errors.New("api not configured")
	}

	// make http request to payment processor api
	err = api.B2B(ctx, payment.PaymentID, req)
	if err != nil {
		return Payment{}, err
	}

	// update payment status
	err = service.repository.Update(ctx, payment.PaymentID, OptionsUpdatePayment{Status: types.Pointer(requests.StatusSent)})
	if err != nil {
		return Payment{}, err
	}

	return payment, nil
}

func (service MpesaService) Status(ctx context.Context, opts OptionsFindPayment) (Payment, error) {

	// find payment
	payment, err := service.repository.FindOne(ctx, opts)
	if err != nil {
		return Payment{}, err
	}

	// check if the payment lifecycle is complete
	if payment.Status.Final() {
		return payment, nil
	}

	// if payment status is not final, fetch true status from external api
	// get shortcode details of the payment
	shortcode, err := service.shortCodeRepository.FindOne(ctx, OptionsFindShortCodes{ShortCodeID: &payment.ShortCodeID})
	if err != nil {
		return Payment{}, err
	}

	// get client api for this payment request
	if shortcode.Service != requests.PartnerDaraja {
		return Payment{}, errors.New("api not configured")
	}

	api := service.provider.GetMpesaApi(shortcode)
	if api == nil {
		return Payment{}, errors.New("api not configured")
	}

	// make http request to payment processor api
	if err = api.Status(ctx, payment); err != nil {
		return Payment{}, err
	}

	// refetch payment, in case the status is updated synchronously from the previous call
	return service.repository.FindOne(ctx, opts)
}
