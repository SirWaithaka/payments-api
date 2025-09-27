package webhooks

import (
	"context"
	"io"
	"time"

	"github.com/SirWaithaka/payments-api/src/domains/requests"
)

type WebhookRequest struct {
	ID        string
	Action    string
	Partner   string
	Payload   io.Reader
	CreatedAt time.Time
}

type Repository interface {
	Add(ctx context.Context, partner, action string, payload []byte) error
	Find(ctx context.Context, id string) (WebhookRequest, error)
}

type Service interface {
	Confirm(ctx context.Context, result *requests.WebhookResult) error
	Process(ctx context.Context, result *requests.WebhookResult) error
}
