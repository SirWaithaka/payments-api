package payments

import (
	"context"

	"github.com/SirWaithaka/payments-api/internal/domains/requests"
)

const (
	darajaEndpoint = "http://localhost:9002/daraja"
)

type WalletPayment struct {
	// generate unique id for the payment request
	PaymentID string
	// should be one of charge, transfer or payout
	Type string
	// code identifying the wallet provider
	BankCode string
	// client generated id for the payment which can be used for reconciliation. Need not be unique
	ClientTransactionID string
	// unique idempotency identifier. Duplicates are rejected
	IdempotencyID string
	// amount to be charged
	Amount string
	// for charge payments, this is the account of customer that will be charged
	// for payouts and transfers, this is the destination account
	DestinationAccountNumber string
	// for transfers, this is the account number of the beneficiary.
	// Not required for other payment types
	Beneficiary string
	// payment description
	Description string
}

type Wallet interface {
	Charge(context.Context, WalletPayment) (requests.Payment, error)
	Payout(context.Context, WalletPayment) (requests.Payment, error)
	Transfer(context.Context, WalletPayment) (requests.Payment, error)
}

type WalletApi interface {
	C2B(context.Context, WalletPayment) error
	B2C(context.Context, WalletPayment) error
	B2B(context.Context, WalletPayment) error
	Status(context.Context, requests.Payment) error
}

type Provider interface {
	GetWalletApi(WalletPayment) WalletApi
}

type Repository interface {
	AddPayment(context.Context, requests.Payment) error
	FindOnePayment(ctx context.Context, opts requests.OptionsFindOnePayment) (requests.Payment, error)
	UpdatePayment(ctx context.Context, id string, opts requests.OptionsUpdatePayment) error
}
