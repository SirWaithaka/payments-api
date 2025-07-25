package postgres

import (
	"bytes"
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/SirWaithaka/payments-api/internal/domains/webhooks"
	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

type WebhookRequestSchema struct {
	ID      string         `gorm:"column:id; primaryKey; type:UUID;"`
	Action  string         `gorm:"column:action"`
	Partner string         `gorm:"column:partner"`
	Payload datatypes.JSON `gorm:"column:payload; type:JSONB"`

	CreatedAt time.Time `gorm:"column:created_at"`
}

func (WebhookRequestSchema) TableName() string {
	return "webhook_requests"
}

func (schema WebhookRequestSchema) ToEntity() webhooks.WebhookRequest {
	return webhooks.WebhookRequest{
		ID:        schema.ID,
		Action:    schema.Action,
		Partner:   schema.Partner,
		Payload:   bytes.NewReader(schema.Payload),
		CreatedAt: schema.CreatedAt,
	}
}

func NewWebhookRepository(db *gorm.DB) WebhookRepository {
	return WebhookRepository{db}
}

type WebhookRepository struct {
	db *gorm.DB
}

func (repo WebhookRepository) Add(ctx context.Context, partner, action string, payload []byte) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("saving webhook request ...")

	// copy payload
	buf := make([]byte, len(payload))
	copy(buf, payload)

	record := WebhookRequestSchema{
		ID:      uuid.Must(uuid.NewV7()).String(),
		Action:  action,
		Partner: partner,
		Payload: buf,
	}

	result := repo.db.WithContext(ctx).Model(WebhookRequestSchema{}).Create(&record)
	if result.Error != nil {
		l.Error().Err(result.Error).Msg("error saving record")
		return Error{Err: result.Error}
	}
	l.Debug().Any(logger.LData, record).Msg("record saved")

	return nil
}

func (repo WebhookRepository) Find(ctx context.Context, id string) (webhooks.WebhookRequest, error) {
	l := zerolog.Ctx(ctx)
	l.Info().Str(logger.LData, id).Msg("find webhook request by id")

	var record WebhookRequestSchema
	result := repo.db.WithContext(ctx).
		Where(WebhookRequestSchema{ID: id}).
		First(&record)

	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error fetching record")
		return webhooks.WebhookRequest{}, Error{Err: err}
	}
	l.Info().Any(logger.LData, record).Msg("record found")

	return record.ToEntity(), nil

}
