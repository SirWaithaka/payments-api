package webhooks_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SirWaithaka/payments-api/pkg/events"
	"github.com/SirWaithaka/payments-api/src/domains/requests"
	"github.com/SirWaithaka/payments-api/src/domains/webhooks"
	"github.com/SirWaithaka/payments-api/src/repositories/postgres"
	"github.com/SirWaithaka/payments-api/testdata"
)

type MockPublisher struct {
	calls uint
}

func (m *MockPublisher) Publish(ctx context.Context, event events.EventType) error {
	m.calls++
	return nil
}

func TestWebhookService_Confirm(t *testing.T) {
	defer testdata.ResetTables(inf)

	repository := postgres.NewWebhookRepository(inf.Storage.PG)
	publisher := &MockPublisher{}

	service := webhooks.NewService(repository, nil, publisher)

	// fake webhook body
	body := `{"ResultCode": "0"}`
	err := service.Confirm(t.Context(), requests.NewWebhookResult("test", "action", strings.NewReader(body)))
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	// check that the webhook was saved
	var record postgres.WebhookRequestSchema
	result := inf.Storage.PG.First(&record)
	if err = result.Error; err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	assert.Equal(t, body, record.Payload.String())
	assert.Equal(t, uint(1), publisher.calls)

}
