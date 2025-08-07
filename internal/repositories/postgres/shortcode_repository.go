package postgres

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/SirWaithaka/payments-api/internal/domains/mpesa"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

type ShortcodeSchema struct {
	ShortCodeID       string  `gorm:"column:id;primaryKey;"`
	Service           string  `gorm:"column:service;check:service<>'';not null"`
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

func (ShortcodeSchema) TableName() string {
	return "mpesa_shortcodes"
}

func (schema ShortcodeSchema) ToEntity() mpesa.ShortCodeConfig {
	shortcode := mpesa.ShortCodeConfig{
		ShortCodeID: schema.ShortCodeID,
		ShortCode:   schema.ShortCode,
		Service:     schema.Service,
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

func (schema *ShortcodeSchema) BeforeCreate(tx *gorm.DB) (err error) {

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

func NewShortCodeRepository(db *gorm.DB) ShortCodeRepository {
	return ShortCodeRepository{db}
}

type ShortCodeRepository struct {
	db *gorm.DB
}

func (repository ShortCodeRepository) Add(ctx context.Context, shortcode mpesa.ShortCodeConfig) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, shortcode).Msg("saving shortcode")

	record := ShortcodeSchema{
		ShortCodeID:       shortcode.ShortCodeID,
		Service:           shortcode.Service,
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
		return err
	}
	l.Debug().Msg("saved record")

	return nil
}

func (repository ShortCodeRepository) Find(ctx context.Context, id string) (mpesa.ShortCodeConfig, error) {
	l := zerolog.Ctx(ctx)
	l.Info().Str(logger.LData, id).Msg("find shortcode by id")

	var record ShortcodeSchema
	result := repository.db.WithContext(ctx).Where(ShortcodeSchema{ShortCodeID: id}).First(&record)
	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error finding record")
		return mpesa.ShortCodeConfig{}, err
	}
	l.Info().Any(logger.LData, record).Msg("record found")

	return record.ToEntity(), nil
}
