package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

type PaymentSchema struct {
	ID                       string  `gorm:"column:id;primaryKey;type:uuid;"`
	PaymentID                string  `gorm:"column:payment_id;unique;check:payment_id<>'';"`
	ClientTransactionID      string  `gorm:"column:client_transaction_id;unique;check:client_transaction_id<>'';"`
	IdempotencyID            string  `gorm:"column:idempotency_id;check:idempotency_id<>'';"`
	PaymentReference         *string `gorm:"column:payment_reference;check:payment_reference<>'';"`
	Amount                   string  `gorm:"column:amount;"`
	Currency                 string  `gorm:"column:currency;"`
	SourceAccountNumber      string  `gorm:"column:source_account_number;"`
	DestinationAccountNumber string  `gorm:"column:external_account_number;"`
	BeneficiaryAccountNumber string  `gorm:"column:beneficiary_account_number;"`
	Description              string  `gorm:"column:description;"`

	Bank   string `gorm:"column:bank;"`
	Type   string `gorm:"column:type;"`
	Status string `gorm:"column:status;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;"`
}

func (PaymentSchema) TableName() string {
	return "payment_requests"
}

func (schema PaymentSchema) ToEntity() requests.Payment {
	// TODO: Revisit these fields
	payment := requests.Payment{
		BankCode:                 schema.Bank,
		PaymentID:                schema.PaymentID,
		ClientTransactionID:      schema.ClientTransactionID,
		IdempotencyID:            schema.IdempotencyID,
		SourceAccountNumber:      schema.SourceAccountNumber,
		DestinationAccountNumber: schema.DestinationAccountNumber,
		Beneficiary:              schema.BeneficiaryAccountNumber,
		Amount:                   schema.Amount,
		Description:              schema.Description,
		Status:                   schema.Status,
	}

	// check if pointer values are nil
	if schema.PaymentReference != nil {
		payment.PaymentReference = *schema.PaymentReference
	}

	return payment
}

func (schema *PaymentSchema) BeforeCreate(tx *gorm.DB) (err error) {
	// generate uuid v7 id for the primary key
	schema.ID = uuid.Must(uuid.NewV7()).String()

	sch := *schema

	// validate that nullable strings should be nil instead of empty
	if sch.PaymentReference != nil && *sch.PaymentReference == "" {
		schema.PaymentReference = nil
	}

	return
}

func (schema *PaymentSchema) FindOptions(opts requests.OptionsFindOnePayment) {
	// configure find options
	if opts.PaymentID != nil {
		schema.PaymentID = *opts.PaymentID
	}
	if opts.IdempotencyID != nil {
		schema.IdempotencyID = *opts.IdempotencyID
	}
	if opts.ClientTransactionID != nil {
		schema.ClientTransactionID = *opts.ClientTransactionID
	}
	if opts.PaymentReference != nil {
		schema.PaymentReference = opts.PaymentReference
	}
}

func NewPaymentsRepository(db *gorm.DB) PaymentsRepository {
	return PaymentsRepository{db: db}
}

type PaymentsRepository struct {
	db *gorm.DB
}

func (repo PaymentsRepository) AddPayment(ctx context.Context, payment requests.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, payment).Msg("saving payment")

	record := PaymentSchema{
		PaymentID:           payment.PaymentID,
		ClientTransactionID: payment.ClientTransactionID,
		IdempotencyID:       payment.IdempotencyID,
		PaymentReference:    &payment.PaymentReference,
		Amount:              payment.Amount,
		//Currency:                 "",
		SourceAccountNumber:      payment.SourceAccountNumber,
		DestinationAccountNumber: payment.DestinationAccountNumber,
		BeneficiaryAccountNumber: payment.Beneficiary,
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

func (repo PaymentsRepository) FindOnePayment(ctx context.Context, opts requests.OptionsFindOnePayment) (requests.Payment, error) {
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
		return requests.Payment{}, Error{Err: err}
	}
	l.Info().Any(logger.LData, record).Msg("fetched record")

	return record.ToEntity(), nil
}

func (repo PaymentsRepository) UpdatePayment(ctx context.Context, id string, opts requests.OptionsUpdatePayment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, opts).Msg("updating payment")

	values := PaymentSchema{}
	if opts.Status != nil {
		values.Status = string(*opts.Status)
	}
	if opts.PaymentReference != nil {
		values.PaymentReference = opts.PaymentReference
	}

	result := repo.db.WithContext(ctx).
		Where(PaymentSchema{PaymentID: id}).
		Updates(values)

	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error updating record")
		return Error{Err: err}
	}
	l.Info().Msg("record updated")

	return nil
}
