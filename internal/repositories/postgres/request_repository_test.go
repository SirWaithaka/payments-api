package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oklog/ulid/v2"

	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/testdata"
)

func Ptr[T any](s T) *T {
	return &s
}

func TestRequestRepository_AddRequest(t *testing.T) {
	defer testdata.ResetTables(inf)

	t.Run("test that it saves record", func(t *testing.T) {
		ctx := context.Background()

		successResponse := map[string]any{
			"MerchantRequestID":   "ed49-4afa-95de-fde98781ae6b37982271",
			"CheckoutRequestID":   "ws_CO_02062024212422901716772048",
			"ResponseCode":        "0",
			"ResponseDescription": "Success. Request accepted for processing",
			"CustomerMessage":     "Success. Request accepted for processing",
		}

		repo := postgres.NewRequestRepository(inf.Storage.PG)

		record := requests.Request{
			RequestID: ulid.Make().String(),
			Status:    "success",
			Partner:   "test",
			Latency:   100 * time.Millisecond,
			Response:  successResponse,
			CreatedAt: time.Now(),
		}
		err := repo.Add(ctx, record)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

	})

	t.Run("test that empty string values are not saved", func(t *testing.T) {
		ctx := context.Background()

		response := map[string]any{
			"ResponseCode": "0",
		}

		repo := postgres.NewRequestRepository(inf.Storage.PG)

		record := requests.Request{
			RequestID: ulid.Make().String(),
			Response:  response,
		}
		err := repo.Add(ctx, record)
		if err == nil {
			t.Errorf("expected non-nil error")
		}

		// confirm error is a check constraint violation
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) {
			if pgErr.Code != postgres.PgCodeCheckConstraintViolation {
				t.Errorf("expected check constraint violation error, got %s", pgErr.Code)
			}
		} else {
			t.Errorf("expected pg error, got %T %v", err, err)
		}

	})
}

func TestRequestRepository_FindOneRequest(t *testing.T) {
	ctx := context.Background()

	paymentsRepo := postgres.NewPaymentsRepository(inf.Storage.PG)
	repo := postgres.NewRequestRepository(inf.Storage.PG)

	t.Run("test that it appends payment details to api request", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		record := requests.Payment{
			PaymentID:           ulid.Make().String(),
			PaymentReference:    ulid.Make().String(),
			ClientTransactionID: ulid.Make().String(),
			IdempotencyID:       ulid.Make().String(),
		}

		apiRequest := requests.Request{
			RequestID:  ulid.Make().String(),
			ExternalID: ulid.Make().String(),
			Partner:    "fake_partner",
			Status:     "received",
			Latency:    1000,
			Response:   nil,
			PaymentID:  record.PaymentID,
		}

		err := paymentsRepo.AddPayment(ctx, record)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		err = repo.Add(ctx, apiRequest)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// now test the find
		request, err := repo.FindOneRequest(ctx, requests.OptionsFindOneRequest{RequestID: &apiRequest.RequestID})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		if request.Payment == nil {
			t.Errorf("expected non-nil value")
		}

		if request.Payment.PaymentReference != record.PaymentReference {
			t.Errorf("expected %s, got %s", request.Payment.PaymentReference, record.PaymentReference)
		}

	})

	t.Run("test that it finds by external id", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		apiRequest := requests.Request{
			RequestID:  ulid.Make().String(),
			ExternalID: ulid.Make().String(),
			Partner:    "fake_partner",
			Status:     "received",
			Latency:    1000,
			Response:   nil,
		}
		err := repo.Add(ctx, apiRequest)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		req, err := repo.FindOneRequest(t.Context(), requests.OptionsFindOneRequest{ExternalID: &apiRequest.ExternalID})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		assert.Equal(t, apiRequest.RequestID, req.RequestID)
		assert.Equal(t, apiRequest.ExternalID, req.ExternalID)

	})
}

func TestRequestRepository_UpdateRequest(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testdata.ResetTables(inf) })

	repo := postgres.NewRequestRepository(inf.Storage.PG)

	testcases := []struct {
		name  string
		input requests.OptionsUpdateRequest
	}{
		{
			name: "test all values provided",
			input: requests.OptionsUpdateRequest{
				ExternalID: Ptr(ulid.Make().String()),
				Status:     Ptr("status"),
				Response:   map[string]any{"key": "value"},
			},
		},
		{
			name:  "test externalID provided",
			input: requests.OptionsUpdateRequest{ExternalID: Ptr(ulid.Make().String())},
		},
		{
			name:  "test status provided",
			input: requests.OptionsUpdateRequest{Status: Ptr("status")},
		},
		{
			name: "test response provided",
			input: requests.OptionsUpdateRequest{
				Response: map[string]any{"key": "value"},
			},
		},
		{
			name: "test response map with multiple keys",
			input: requests.OptionsUpdateRequest{
				Response: map[string]any{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
		},
		{
			name: "test response map with nested values",
			input: requests.OptionsUpdateRequest{
				Response: map[string]any{
					"key": map[string]any{
						"nested": "value",
					},
				},
			},
		},
	}

	for _, tc := range testcases {

		t.Run(tc.name, func(t *testing.T) {
			// create and save a request record
			req := requests.Request{
				RequestID: ulid.Make().String(),
				Partner:   "daraja",
			}

			err := repo.Add(ctx, req)
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			err = repo.UpdateRequest(ctx, req.RequestID, tc.input)
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			// pull record and check values
			record := postgres.RequestSchema{}
			r := inf.Storage.PG.Where(postgres.RequestSchema{RequestID: req.RequestID}).First(&record)
			if r.Error != nil {
				t.Errorf("expected nil error, got %v", r.Error)
			}

			// assert values are equal to expected
			assert.Equal(t, tc.input.ExternalID, record.ExternalID)
			assert.Equal(t, tc.input.Status, record.Status)
			assert.Equal(t, tc.input.Response, record.ToEntity().Response)

		})

	}

}
