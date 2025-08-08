package mpesa

import (
	"context"

	"github.com/SirWaithaka/payments-api/internal/domains/requests"
)

type PaymentType string

const (
	PaymentTypeWalletCharge   PaymentType = "charge"
	PaymentTypeWalletPayout   PaymentType = "payout"
	PaymentTypeWalletTransfer PaymentType = "transfer"
)

func (p PaymentType) String() string {
	return string(p)
}

func ToPaymentType(s string) PaymentType {
	switch s {
	case string(PaymentTypeWalletCharge):
		return PaymentTypeWalletCharge
	case string(PaymentTypeWalletPayout):
		return PaymentTypeWalletPayout
	case string(PaymentTypeWalletTransfer):
		return PaymentTypeWalletTransfer
	default:
		return "unknown"
	}
}

type Payment struct {
	// generate unique id for the payment request
	PaymentID string
	// should be one of charge, transfer or payout
	Type PaymentType
	// client generated id for the payment which can be used for reconciliation. Need not be unique
	ClientTransactionID string
	// unique idempotency identifier. Duplicates are rejected
	IdempotencyID string
	// payment reference from the payment processor
	PaymentReference string
	// amount to be charged
	Amount string
	// for charge payments, this is the account of customer that will be charged
	SourceAccountNumber string
	// for charge payments, this is the account where funds will be credited
	DestinationAccountNumber string
	// for transfers, this is the account number of the beneficiary.
	// Not required for other payment types
	Beneficiary string
	// payment description
	Description string
	// shortcode id
	ShortCodeID string
	// payment status
	Status requests.Status
}

type PaymentRequest struct {
	IdempotencyID         string
	ClientTransactionID   string
	Amount                string
	ExternalAccountNumber string
	Beneficiary           string
	Description           string
}

type ShortCode struct {
	ShortCodeID       string
	ShortCode         string           // business pay bill or buy goods account
	Service           requests.Partner // service can be either daraja or quikk
	Type              string           // types of payment the shortcode can be used for
	InitiatorName     string           // daraja api initiator name
	InitiatorPassword string           // daraja api initiator password
	Passphrase        string           // (optional) passphrase for c2b transfers
	Key               string           // daraja app consumer key or quikk app key
	Secret            string           // daraja app consumer secret or quikk app secret
	CallbackURL       string           // callback url for shortcode async responses
}

type OptionsFindPayment struct {
	PaymentID           *string
	ClientTransactionID *string
	IdempotencyID       *string
	PaymentReference    *string
}

type OptionsUpdatePayment struct {
	Status           *requests.Status
	PaymentReference *string
}

type OptionsFindShortCodes struct {
	ShortCodeID *string
	Service     *requests.Partner
	Type        *string
	ShortCode   *string
}

type Repository interface {
	Add(context.Context, Payment) error
	FindOne(ctx context.Context, opts OptionsFindPayment) (Payment, error)
	Update(ctx context.Context, id string, opts OptionsUpdatePayment) error
}

type ShortCodeRepository interface {
	Add(ctx context.Context, shortcode ShortCode) error
	FindOne(ctx context.Context, opts OptionsFindShortCodes) (ShortCode, error)
	FindMany(ctx context.Context, opts OptionsFindShortCodes) ([]ShortCode, error)
}

type API interface {
	B2B(ctx context.Context, paymentID string, req PaymentRequest) error
	Status(ctx context.Context, payment Payment) error
}

type Provider interface {
	GetMpesaApi(ShortCode) API
}

type Service interface {
	Transfer(ctx context.Context, request PaymentRequest) (Payment, error)
	Status(ctx context.Context, opts OptionsFindPayment) (Payment, error)
}
