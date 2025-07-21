package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

type PaymentSchema struct {
	ID                       string `gorm:"column:id;primaryKey;type:uuid;"`
	PaymentID                string `gorm:"column:payment_id;unique;check:payment_id<>'';"`
	ClientTransactionID      string `gorm:"column:client_transaction_id;unique;check:client_transaction_id<>'';"`
	IdempotencyID            string `gorm:"column:idempotency_id;check:idempotency_id<>'';"`
	PaymentReference         string `gorm:"column:payment_reference;check:payment_reference<>'';"`
	Amount                   string `gorm:"column:amount;"`
	Currency                 string `gorm:"column:currency;"`
	SourceAccountNumber      string `gorm:"column:source_account_number;"`
	DestinationAccountNumber string `gorm:"column:external_account_number;"`
	BeneficiaryAccountNumber string `gorm:"column:beneficiary_account_number;"`
	Description              string `gorm:"column:description;"`

	Bank   string `gorm:"column:bank;"`
	Type   string `gorm:"column:type;"`
	Status string `gorm:"column:status;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;"`
}

func (PaymentSchema) TableName() string {
	return "payment_requests"
}

func (schema PaymentSchema) ToEntity() payments.Payment {
	// TODO: Revisit these fields
	return payments.Payment{
		BankCode:                 schema.Bank,
		PaymentID:                schema.PaymentID,
		ClientTransactionID:      schema.ClientTransactionID,
		IdempotencyID:            schema.IdempotencyID,
		PaymentReference:         schema.PaymentReference,
		AccountNumber:            schema.SourceAccountNumber,
		ExternalAccountNumber:    schema.DestinationAccountNumber,
		BeneficiaryAccountNumber: schema.BeneficiaryAccountNumber,
		Amount:                   schema.Amount,
		Description:              schema.Description,
	}
}

func (schema *PaymentSchema) BeforeCreate(tx *gorm.DB) (err error) {
	// generate uuid v7 id for the primary key
	schema.ID = uuid.Must(uuid.NewV7()).String()

	return
}

func (schema *PaymentSchema) FindOptions(opts payments.OptionsFindOnePayment) {
	// configure find options
	if opts.PaymentID != nil {
		schema.PaymentID = *opts.PaymentID
	}
	if opts.IdempotencyID != nil {
		schema.IdempotencyID = *opts.IdempotencyID
	}
	if opts.TransactionID != nil {
		schema.ClientTransactionID = *opts.TransactionID
	}
	if opts.PaymentReference != nil {
		schema.PaymentReference = *opts.PaymentReference
	}
}

func NewRepository(db *gorm.DB) Repository {
	return Repository{db: db}
}

type Repository struct {
	db *gorm.DB
}

func (repo Repository) AddPayment(ctx context.Context, payment payments.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, payment).Msg("saving payment")

	record := PaymentSchema{
		PaymentID:           payment.PaymentID,
		ClientTransactionID: payment.ClientTransactionID,
		IdempotencyID:       payment.IdempotencyID,
		PaymentReference:    payment.PaymentReference,
		Amount:              payment.Amount,
		//Currency:                 "",
		SourceAccountNumber:      payment.AccountNumber,
		DestinationAccountNumber: payment.ExternalAccountNumber,
		BeneficiaryAccountNumber: payment.BeneficiaryAccountNumber,
		Description:              payment.Description,
		Bank:                     payment.BankCode,
		//Type:                     "",
		//Status:                   "",
		//CreatedAt:                time.Time{},
		//UpdatedAt:                time.Time{},
	}

	result := repo.db.WithContext(ctx).Create(&record)
	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error saving record")
		return Error{Err: err}
	}
	l.Debug().Msg("saved record")

	return nil
}

func (repo Repository) FindOnePayment(ctx context.Context, opts payments.OptionsFindOnePayment) (payments.Payment, error) {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, opts).Msg("find options")

	// configure find options
	where := PaymentSchema{}
	where.FindOptions(opts)
	l.Info().Any(logger.LData, where).Msg("query params")

	var record PaymentSchema
	result := repo.db.WithContext(ctx).Where(where).First(&record)
	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error fetching record")
		return payments.Payment{}, Error{Err: err}
	}
	l.Info().Any(logger.LData, record).Msg("fetched record")

	return record.ToEntity(), nil
}
