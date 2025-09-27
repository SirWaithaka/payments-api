package postgres_test

import (
	"context"
	"testing"

	"github.com/SirWaithaka/payments-api/src/repositories/postgres"
	"github.com/SirWaithaka/payments-api/testdata"
)

func TestWebhookRepository_Add(t *testing.T) {
	ctx := context.Background()

	repo := postgres.NewWebhookRepository(inf.Storage.PG)

	t.Run("test that it saves record", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		payload := []byte(`{"name":"john doe","amount":10}`)
		err := repo.Add(ctx, "fake_partner", "fake_action", payload)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		var records []postgres.WebhookRequestSchema
		result := inf.Storage.PG.Find(&records)
		if result.Error != nil {
			t.Errorf("expected nil error, got %v", result.Error)
		}

	})

	t.Run("test that nil payload saves as nil", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		err := repo.Add(ctx, "fake_partner", "fake_action", nil)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		var records []postgres.WebhookRequestSchema
		result := inf.Storage.PG.Find(&records)
		if result.Error != nil {
			t.Errorf("expected nil error, got %v", result.Error)
		}

		if len(records) == 0 {
			t.Errorf("expected non-empty result, got empty")
		}

	})
}
