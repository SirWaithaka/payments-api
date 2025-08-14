package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

type MpesaPaymentSchema struct {
	ID                       string  `gorm:"column:id;primaryKey;type:uuid;"`
	PaymentID                string  `gorm:"column:payment_id;not null"`
	Type                     string  `gorm:"column:type;not null"`
	Status                   string  `gorm:"column:status;not null"`
	ClientTransactionID      string  `gorm:"column:client_transaction_id;not null"`
	IdempotencyID            string  `gorm:"column:idempotency_id;not null"`
	PaymentReference         *string `gorm:"column:payment_reference;"`
	Amount                   string  `gorm:"column:amount;not null"`
	SourceAccountNumber      string  `gorm:"column:source_account_number;not null"`
	DestinationAccountNumber string  `gorm:"column:destination_account_number;not null"`
	Beneficiary              *string `gorm:"column:beneficiary;"`
	Description              *string `gorm:"column:description;check:description<>'';"`

	ShortCodeID *string `gorm:"column:shortcode_id;"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;"`
}

func (MpesaPaymentSchema) TableName() string {
	return "mpesa_payments"
}

func (schema MpesaPaymentSchema) ToEntity() mpesa.Payment {
	payment := mpesa.Payment{
		PaymentID:                schema.PaymentID,
		Type:                     mpesa.ToPaymentType(schema.Type),
		ClientTransactionID:      schema.ClientTransactionID,
		IdempotencyID:            schema.IdempotencyID,
		Amount:                   schema.Amount,
		SourceAccountNumber:      schema.SourceAccountNumber,
		DestinationAccountNumber: schema.DestinationAccountNumber,
		Status:                   requests.ToStatus(schema.Status),
	}

	// check if pointer values are nil
	if schema.PaymentReference != nil {
		payment.PaymentReference = *schema.PaymentReference
	}
	if schema.Beneficiary != nil {
		payment.Beneficiary = *schema.Beneficiary
	}
	if schema.ShortCodeID != nil {
		payment.ShortCodeID = *schema.ShortCodeID
	}
	if schema.Description != nil {
		payment.Description = *schema.Description
	}

	return payment
}

func (schema *MpesaPaymentSchema) BeforeCreate(tx *gorm.DB) (err error) {
	// generate uuid v7 id for the primary key
	schema.ID = uuid.Must(uuid.NewV7()).String()

	sch := *schema

	// validate that nullable strings should be nil instead of empty
	if sch.PaymentReference != nil && *sch.PaymentReference == "" {
		schema.PaymentReference = nil
	}
	if sch.Beneficiary != nil && *sch.Beneficiary == "" {
		schema.Beneficiary = nil
	}
	if sch.ShortCodeID != nil && *sch.ShortCodeID == "" {
		schema.ShortCodeID = nil
	}
	if sch.Description != nil && *sch.Description == "" {
		schema.Description = nil
	}

	return
}

func (schema *MpesaPaymentSchema) FindOptions(opts mpesa.OptionsFindPayment) {
	// by default, gorm ignores zero value struct properties in the where clause

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

func NewMpesaPaymentsRepository(db *gorm.DB) MpesaPaymentsRepository {
	return MpesaPaymentsRepository{db}
}

type MpesaPaymentsRepository struct {
	db *gorm.DB
}

func (repository MpesaPaymentsRepository) Add(ctx context.Context, payment mpesa.Payment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, payment).Msg("saving payment")

	record := MpesaPaymentSchema{
		PaymentID:                payment.PaymentID,
		Type:                     string(payment.Type),
		Status:                   string(payment.Status),
		ClientTransactionID:      payment.ClientTransactionID,
		IdempotencyID:            payment.IdempotencyID,
		PaymentReference:         &payment.PaymentReference,
		Amount:                   payment.Amount,
		SourceAccountNumber:      payment.SourceAccountNumber,
		DestinationAccountNumber: payment.DestinationAccountNumber,
		Beneficiary:              &payment.Beneficiary,
		Description:              &payment.Description,
		ShortCodeID:              &payment.ShortCodeID,
	}

	result := repository.db.WithContext(ctx).Create(&record)
	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error saving record")
		return Error{Err: err}
	}
	l.Debug().Msg("saved record")

	return nil

}

func (repository MpesaPaymentsRepository) FindOne(ctx context.Context, opts mpesa.OptionsFindPayment) (mpesa.Payment, error) {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, opts).Msg("find options")

	// configure find options
	where := MpesaPaymentSchema{}
	where.FindOptions(opts)
	l.Info().Any(logger.LData, where).Msg("query params")

	var record MpesaPaymentSchema
	result := repository.db.WithContext(ctx).Where(where).First(&record)
	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error fetching record")
		return mpesa.Payment{}, Error{Err: err}
	}
	l.Info().Any(logger.LData, record).Msg("record found")

	return record.ToEntity(), nil
}

func (repository MpesaPaymentsRepository) Update(ctx context.Context, id string, opts mpesa.OptionsUpdatePayment) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, opts).Msg("update options")

	values := MpesaPaymentSchema{}
	if opts.Status != nil {
		values.Status = string(*opts.Status)
	}
	if opts.PaymentReference != nil {
		values.PaymentReference = opts.PaymentReference
	}

	result := repository.db.WithContext(ctx).
		Where(MpesaPaymentSchema{PaymentID: id}).
		Updates(values)

	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error updating record")
		return Error{Err: err}
	}
	l.Info().Msg("record updated")

	return nil
}
