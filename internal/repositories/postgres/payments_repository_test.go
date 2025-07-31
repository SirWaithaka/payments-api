package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/oklog/ulid/v2"

	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/testdata"
)

func TestRequestRepository_AddPayment(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testdata.ResetTables(inf) })

	repo := postgres.NewPaymentsRepository(inf.Storage.PG)

	t.Run("test that empty payment instance cant be saved", func(t *testing.T) {
		testcases := []struct {
			input requests.Payment
			name  string
		}{
			{
				name:  "test reference column",
				input: requests.Payment{
					//Reference: ulid.Make().String(),
				},
			},
			{
				name:  "test status column",
				input: requests.Payment{
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
		record := requests.Payment{
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

func TestRequestRepository_FindOnePayment(t *testing.T) {
	ctx := context.Background()

	repo := postgres.NewPaymentsRepository(inf.Storage.PG)

	t.Run("test that it finds records", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		record := requests.Payment{
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

		testcases := []struct {
			name  string
			input requests.OptionsFindOnePayment
		}{
			{name: "test that it finds by payment id", input: requests.OptionsFindOnePayment{PaymentID: &record.PaymentID}},
			{name: "test that it finds by client transaction id", input: requests.OptionsFindOnePayment{ClientTransactionID: &record.ClientTransactionID}},
			{name: "test that it finds by idempotency id", input: requests.OptionsFindOnePayment{IdempotencyID: &record.IdempotencyID}},
			{name: "test that it finds by payment reference", input: requests.OptionsFindOnePayment{PaymentReference: &record.PaymentReference}},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				payment, err := repo.FindOnePayment(ctx, tc.input)
				if err != nil {
					t.Errorf("expected nil error, got %v", err)
				}

				assert.Equal(t, record.PaymentID, payment.PaymentID)
				assert.Equal(t, record.ClientTransactionID, payment.ClientTransactionID)
				assert.Equal(t, record.IdempotencyID, payment.IdempotencyID)
				assert.Equal(t, record.PaymentReference, payment.PaymentReference)
			})
		}
	})

	t.Run("test that it returns not found error when no record is found", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		_, err := repo.FindOnePayment(ctx, requests.OptionsFindOnePayment{PaymentID: Ptr("fake_id")})
		if err == nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// check error is not found
		var e interface{ NotFound() bool }
		if !errors.As(err, &e) {
			t.Errorf("expected not found error, got %T %v", err, err)
		}

	})
}

func TestRequestRepository_UpdatePayment(t *testing.T) {
	ctx := context.Background()

	repo := postgres.NewPaymentsRepository(inf.Storage.PG)

	testcases := []struct {
		name  string
		input requests.OptionsUpdatePayment
	}{
		{
			name: "test all values provided",
			input: requests.OptionsUpdatePayment{
				Status:           Ptr(requests.StatusSucceeded),
				PaymentReference: Ptr(ulid.Make().String()),
			},
		},
	}

	for _, tc := range testcases {

		t.Run(tc.name, func(t *testing.T) {
			// create and save payment record

			payment := requests.Payment{
				PaymentID:           ulid.Make().String(),
				ClientTransactionID: ulid.Make().String(),
				IdempotencyID:       ulid.Make().String(),
				PaymentReference:    ulid.Make().String(),
				Status:              "status",
			}

			err := repo.AddPayment(ctx, payment)
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			err = repo.UpdatePayment(ctx, payment.PaymentID, tc.input)
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			// pull the record and check values
			record := postgres.PaymentSchema{}
			r := inf.Storage.PG.Where(postgres.PaymentSchema{PaymentID: payment.PaymentID}).First(&record)
			if r.Error != nil {
				t.Errorf("expected nil error, got %v", r.Error)
			}

			assert.Equal(t, string(*tc.input.Status), record.Status)
			assert.Equal(t, tc.input.PaymentReference, record.PaymentReference)
		})
	}
}
