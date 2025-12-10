package mpesa

import (
	"context"

	"github.com/SirWaithaka/payments-api/src/domains/requests"
)

type PaymentType string

const (
	PaymentTypeCharge   PaymentType = "charge"
	PaymentTypePayout   PaymentType = "payout"
	PaymentTypeTransfer PaymentType = "transfer"
)

func (p PaymentType) String() string {
	return string(p)
}

func (p PaymentType) Valid() bool {
	return p == PaymentTypeCharge || p == PaymentTypePayout || p == PaymentTypeTransfer
}

func ToPaymentType(s string) PaymentType {
	switch s {
	case string(PaymentTypeCharge):
		return PaymentTypeCharge
	case string(PaymentTypePayout):
		return PaymentTypePayout
	case string(PaymentTypeTransfer):
		return PaymentTypeTransfer
	default:
		return "unknown"
	}
}

type AccountType string

const (
	AccountTypeMSISDN  AccountType = "msisdn"
	AccountTypePaybill AccountType = "paybill"
	AccountTypeTill    AccountType = "till"
)

func (a AccountType) String() string { return string(a) }

func ToAccountType(s string) AccountType {
	switch s {
	case string(AccountTypeMSISDN):
		return AccountTypeMSISDN
	case string(AccountTypePaybill):
		return AccountTypePaybill
	case string(AccountTypeTill):
		return AccountTypeTill
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
	// for charge payments, this is the account of the customer that will be charged
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
	ExternalAccountType   AccountType
	ExternalAccountNumber string
	Beneficiary           string
	Description           string
}

type ReversalRequest struct {
	PaymentID           string
	IdempotencyID       string
	ClientTransactionID string
	Amount              string
	PaymentReference    string
}

type ShortCode struct {
	ShortCodeID       string
	Environment       string           // enum of sandbox, production
	ShortCode         string           // business pay bill or buy goods account
	Priority          uint             // low value means higher priority, min=1
	Service           requests.Partner // service can be either daraja or quikk
	Type              PaymentType      // types of payment the shortcode can be used for
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
	C2B(ctx context.Context, paymentID string, req PaymentRequest) error
	B2C(ctx context.Context, paymentID string, req PaymentRequest) error
	B2B(ctx context.Context, paymentID string, req PaymentRequest) error
	Status(ctx context.Context, payment Payment) error
}

type Provider interface {
	GetMpesaApi(ShortCode) API
	GetWebhookProcessor(requests.Partner) requests.WebhookProcessor
}

type Service interface {
	Charge(ctx context.Context, request PaymentRequest) (Payment, error)
	Payout(ctx context.Context, request PaymentRequest) (Payment, error)
	Transfer(ctx context.Context, request PaymentRequest) (Payment, error)
	Status(ctx context.Context, opts OptionsFindPayment) (Payment, error)
	ProcessWebhook(ctx context.Context, result *requests.WebhookResult) error
}

type ShortCodeService interface {
	Add(ctx context.Context, shortcode ShortCode) error
}
