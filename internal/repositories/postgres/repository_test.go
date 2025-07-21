package postgres_test

import (
	"context"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/testdata"
)

func TestRequestRepository_AddPayment(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testdata.ResetTables(inf) })

	repo := postgres.NewRepository(inf.Storage.PG)

	t.Run("test that empty payment instance cant be saved", func(t *testing.T) {
		testcases := []struct {
			input payments.Payment
			name  string
		}{
			{
				name:  "test reference column",
				input: payments.Payment{
					//Reference: ulid.Make().String(),
				},
			},
			{
				name:  "test status column",
				input: payments.Payment{
					//Reference: ulid.Make().String(),
				},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				err := repo.AddPayment(ctx, tc.input)
				if err == nil {
					t.Errorf("expected non-nil error")
				}

			})
		}
	})

	t.Run("test that a valid payment request is saved successfully", func(t *testing.T) {
		record := payments.Payment{
			PaymentID:           ulid.Make().String(),
			ClientTransactionID: ulid.Make().String(),
			IdempotencyID:       ulid.Make().String(),
			PaymentReference:    ulid.Make().String(),
			Status:              "status",
		}
		err := repo.AddPayment(ctx, record)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		var records []postgres.PaymentSchema
		result := inf.Storage.PG.Find(&records)
		if result.Error != nil {
			t.Errorf("expected nil error, got %v", result.Error)
		}

		if len(records) != 1 {
			t.Errorf("expected 1 record, got %d", len(records))
		}
	})
}
