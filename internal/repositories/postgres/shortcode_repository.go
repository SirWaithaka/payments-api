package postgres

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

type ShortCodeSchema struct {
	ShortCodeID       string  `gorm:"column:id;primaryKey;"`
	Service           string  `gorm:"column:service;check:service<>'';not null"`
	Type              string  `gorm:"column:type;check:type<>'';not null"`
	ShortCode         string  `gorm:"column:shortcode;check:shortcode<>'';not null"`
	InitiatorName     *string `gorm:"column:initiator_name;"`
	InitiatorPassword *string `gorm:"column:initiator_password;"`
	Passphrase        *string `gorm:"column:passphrase;"`
	Key               string  `gorm:"column:key;check:key<>'';not null"`
	Secret            string  `gorm:"column:secret;check:secret<>'';not null"`
	CallbackURL       string  `gorm:"column:callback_url;not null"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;"`
}

func (ShortCodeSchema) TableName() string {
	return "mpesa_shortcodes"
}

func (schema ShortCodeSchema) ToEntity() mpesa.ShortCode {
	shortcode := mpesa.ShortCode{
		ShortCodeID: schema.ShortCodeID,
		ShortCode:   schema.ShortCode,
		Service:     requests.ToPartner(schema.Service),
		Type:        schema.Type,
		Key:         schema.Key,
		Secret:      schema.Secret,
		CallbackURL: schema.CallbackURL,
	}

	// check if pointer values are nil
	if schema.InitiatorName != nil {
		shortcode.InitiatorName = *schema.InitiatorName
	}
	if schema.InitiatorPassword != nil {
		shortcode.InitiatorPassword = *schema.InitiatorPassword
	}
	if schema.Passphrase != nil {
		shortcode.Passphrase = *schema.Passphrase
	}

	return shortcode
}

func (schema *ShortCodeSchema) BeforeCreate(tx *gorm.DB) (err error) {

	// validate that nullable columns should be nil instead of zero values
	sch := *schema

	if sch.InitiatorName != nil && *sch.InitiatorName == "" {
		schema.InitiatorName = nil
	}
	if sch.InitiatorPassword != nil && *sch.InitiatorPassword == "" {
		schema.InitiatorPassword = nil
	}
	if sch.Passphrase != nil && *sch.Passphrase == "" {
		schema.Passphrase = nil
	}

	return
}

func (schema *ShortCodeSchema) FindOptions(opts mpesa.OptionsFindShortCodes) {
	// by default, gorm ignores zero value struct properties in the where clause

	// configure find options
	if opts.ShortCodeID != nil {
		schema.ShortCodeID = *opts.ShortCodeID
	}
	if opts.Service != nil {
		schema.Service = opts.Service.String()
	}
	if opts.Type != nil {
		schema.Type = *opts.Type
	}
	if opts.ShortCode != nil {
		schema.ShortCode = *opts.ShortCode
	}
}

func NewShortCodeRepository(db *gorm.DB) ShortCodeRepository {
	return ShortCodeRepository{db}
}

type ShortCodeRepository struct {
	db *gorm.DB
}

func (repository ShortCodeRepository) Add(ctx context.Context, shortcode mpesa.ShortCode) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, shortcode).Msg("saving shortcode")

	record := ShortCodeSchema{
		ShortCodeID:       shortcode.ShortCodeID,
		Service:           shortcode.Service.String(),
		Type:              shortcode.Type,
		ShortCode:         shortcode.ShortCode,
		InitiatorName:     &shortcode.InitiatorName,
		InitiatorPassword: &shortcode.InitiatorPassword,
		Passphrase:        &shortcode.Passphrase,
		Key:               shortcode.Key,
		Secret:            shortcode.Secret,
		CallbackURL:       shortcode.CallbackURL,
	}

	result := repository.db.WithContext(ctx).Create(&record)
	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error saving record")
		return Error{Err: err}
	}
	l.Debug().Msg("saved record")

	return nil
}

func (repository ShortCodeRepository) FindOne(ctx context.Context, opts mpesa.OptionsFindShortCodes) (mpesa.ShortCode, error) {
	l := zerolog.Ctx(ctx)
	l.Info().Any(logger.LData, opts).Msg("find shortcode by id")

	// build find options
	where := ShortCodeSchema{}
	where.FindOptions(opts)
	l.Info().Any(logger.LData, where).Msg("query params")

	var record ShortCodeSchema
	result := repository.db.WithContext(ctx).Where(where).First(&record)
	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error finding record")
		return mpesa.ShortCode{}, Error{Err: err}
	}
	l.Info().Any(logger.LData, record).Msg("record found")

	return record.ToEntity(), nil
}

func (repository ShortCodeRepository) FindMany(ctx context.Context, opts mpesa.OptionsFindShortCodes) ([]mpesa.ShortCode, error) {
	l := zerolog.Ctx(ctx)
	l.Info().Msg("find shortcodes")

	// build find options
	where := ShortCodeSchema{}
	where.FindOptions(opts)
	l.Info().Any(logger.LData, where).Msg("query params")

	var records []ShortCodeSchema
	result := repository.db.WithContext(ctx).Where(where).Find(&records)
	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error finding records")
		return []mpesa.ShortCode{}, Error{Err: err}
	}
	l.Info().Any(logger.LData, records).Msg("records found")

	var shortcodes []mpesa.ShortCode
	for _, record := range records {
		shortcodes = append(shortcodes, record.ToEntity())
	}

	return shortcodes, nil
}
