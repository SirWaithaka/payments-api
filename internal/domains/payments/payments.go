package payments

import (
	"context"
	"time"
)

const (
	darajaEndpoint = "http://localhost:9002/daraja"
)

type Payment struct {
	BankCode string
	// automatically generated unique id
	PaymentID string
	// client's transaction reference
	ClientTransactionID string
	// client generated unique id identifying the payment request
	IdempotencyID string
	// payment reference from the payment processor
	PaymentReference string
	// source account
	SourceAccountNumber string
	// destination account
	DestinationAccountNumber string
	// (optional) - this can be an account number, account name
	// of the beneficiary for the payment
	Beneficiary string
	// amount for payment
	Amount string
	// short description for payment
	Description string
	// status of the payment, "received", "pending", "completed", "failed", "refunded"
	Status string
}

type WalletPayment struct {
	// generate unique id for the payment request
	PaymentID string
	// should be one of charge, transfer or payout
	Type string
	// code identifying the wallet provider
	BankCode string
	// external identifier for the transfer which can be used for reconciliation. Need not be unique
	TransactionID string
	// unique idempotency identifier. Duplicates are rejected
	IdempotencyID string
	// amount to be charged
	Amount string
	// for charge payments, this is the account of customer that will be charged
	// for payouts and transfers, this is the destination account
	ExternalAccountNumber string
	// for transfers, this is the account number of the beneficiary.
	// Not required for other payment types
	BeneficiaryAccountNumber string
	// payment description
	Description string
}

type Request struct {
	RequestID  string // unique request id
	PaymentID  string // foreign id tied to the original payment request
	ExternalID string // request id we get back from partner from response
	Partner    string
	Status     string
	Latency    time.Duration
	Response   map[string]any
	CreatedAt  time.Time

	Payment *Payment
}

type Wallet interface {
	Charge(context.Context, WalletPayment) (Payment, error)
	Payout(context.Context, WalletPayment) (Payment, error)
	Transfer(context.Context, WalletPayment) (Payment, error)
}

type WalletApi interface {
	C2B(context.Context, WalletPayment) error
	B2C(context.Context, WalletPayment) error
	B2B(context.Context, WalletPayment) error
	Status(context.Context, Payment) error
}

type Provider interface {
	GetWalletApi(WalletPayment) WalletApi
}

type OptionsFindOnePayment struct {
	PaymentID        *string
	TransactionID    *string
	PaymentReference *string
	IdempotencyID    *string
}

type OptionsFindOneRequest struct {
	RequestID  *string
	ExternalID *string
}

type OptionsUpdateRequest struct {
	ExternalID *string
	Status     *string
	Response   map[string]any
}

type Repository interface {
	AddPayment(context.Context, Payment) error
	FindOnePayment(ctx context.Context, opts OptionsFindOnePayment) (Payment, error)
}

type RequestRepository interface {
	Add(ctx context.Context, req Request) error
	FindOneRequest(ctx context.Context, opts OptionsFindOneRequest) (Request, error)
	UpdateRequest(ctx context.Context, id string, opts OptionsUpdateRequest) error
}
