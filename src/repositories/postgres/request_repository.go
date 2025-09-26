package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/SirWaithaka/payments-api/pkg/logger"
	"github.com/SirWaithaka/payments-api/pkg/types"
	"github.com/SirWaithaka/payments-api/src/domains/requests"
)

type RequestSchema struct {
	ID         string    `gorm:"column:id;primaryKey;type:uuid;"`
	RequestID  string    `gorm:"column:request_id;unique;check:request_id<>'';"` // business id
	ExternalID *string   `gorm:"column:external_id;check:external_id<>'';"`
	Partner    string    `gorm:"column:partner;check:partner<>'';"`
	Status     *string   `gorm:"column:status;check:status<>'';"`
	Latency    int64     `gorm:"column:latency_ms"` // duration in milliseconds
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`

	Response datatypes.JSONMap `gorm:"column:response;type:json"`

	// define a belongsTo relationship
	PaymentID *string `gorm:"column:payment_id"`
}

func (RequestSchema) TableName() string {
	return "api_requests"
}

// BeforeCreate hook checks/validates that nullable values are not empty primitive types
func (schema *RequestSchema) BeforeCreate(tx *gorm.DB) (err error) {
	// create id
	schema.ID = uuid.Must(uuid.NewV7()).String()

	sch := *schema

	if sch.ExternalID != nil && *sch.ExternalID == "" {
		schema.ExternalID = nil
	}
	if sch.Status != nil && *sch.Status == "" {
		schema.Status = nil
	}
	if sch.PaymentID != nil && *sch.PaymentID == "" {
		schema.PaymentID = nil
	}

	return
}

func (schema RequestSchema) ToEntity() requests.Request {
	request := requests.Request{
		RequestID: schema.RequestID,
		Partner:   schema.Partner,
		Latency:   time.Duration(schema.Latency) * time.Millisecond,
		CreatedAt: schema.CreatedAt,
	}

	// check for null values
	if schema.ExternalID != nil {
		request.ExternalID = *schema.ExternalID
	}

	if schema.Status != nil {
		request.Status = requests.ToStatus(*schema.Status)
	}

	if schema.PaymentID != nil {
		request.PaymentID = (*schema.PaymentID)
	}

	if schema.Response != nil {
		request.Response = schema.Response
	}

	return request
}

func NewRequestRepository(db *gorm.DB) RequestRepository {
	return RequestRepository{db}
}

type RequestRepository struct {
	db *gorm.DB
}

func (repository RequestRepository) Add(ctx context.Context, req requests.Request) error {
	l := zerolog.Ctx(ctx)
	l.Info().Interface(logger.LData, req).Msg("saving api request ...")

	record := RequestSchema{
		RequestID:  req.RequestID,
		ExternalID: &req.ExternalID,
		Partner:    req.Partner,
		Status:     types.Pointer(req.Status.String()),
		Latency:    req.Latency.Milliseconds(),
		Response:   req.Response,
		PaymentID:  &req.PaymentID,
	}

	result := repository.db.WithContext(ctx).Model(RequestSchema{}).Create(&record)
	if result.Error != nil {
		l.Error().Err(result.Error).Msg("save error")
		return Error{Err: result.Error}
	}
	l.Info().Any(logger.LData, record).Msg("api request record saved")

	return nil
}

// FindOne selects the most recent record after ordering "created_at" in descending
func (repository RequestRepository) FindOne(ctx context.Context, opts requests.OptionsFindRequest) (requests.Request, error) {
	l := zerolog.Ctx(ctx)
	l.Info().Any(logger.LData, opts).Msg("fetch one api request")

	// configure find options
	where := RequestSchema{}
	if opts.RequestID != nil {
		where.RequestID = *opts.RequestID
	}
	if opts.ExternalID != nil {
		where.ExternalID = opts.ExternalID
	}
	if opts.PaymentID != nil {
		where.PaymentID = opts.PaymentID
	}

	var record RequestSchema
	result := repository.db.WithContext(ctx).
		Where(where).
		Take(&record).
		Order("created_at desc")
	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error fetching record")
		return requests.Request{}, Error{Err: err}
	}
	l.Info().Any(logger.LData, record).Msg("record fetched")

	return record.ToEntity(), nil

}

func (repository RequestRepository) UpdateRequest(ctx context.Context, id string, opts requests.OptionsUpdateRequest) error {
	l := zerolog.Ctx(ctx)
	l.Info().Msg("updating api request")

	values := RequestSchema{}
	if opts.Status != nil {
		values.Status = types.Pointer(opts.Status.String())
	}
	if opts.ExternalID != nil {
		values.ExternalID = opts.ExternalID
	}
	if opts.Response != nil {
		values.Response = opts.Response
	}
	if opts.Latency != nil {
		values.Latency = opts.Latency.Milliseconds()
	}

	// when using a struct to update, gorm will ignore zero values
	result := repository.db.WithContext(ctx).
		Where(RequestSchema{RequestID: id}).
		Updates(values)

	if err := result.Error; err != nil {
		l.Error().Err(err).Msg("error updating record")
		return Error{Err: err}
	}
	l.Info().Msg("record updated")

	return nil
}
