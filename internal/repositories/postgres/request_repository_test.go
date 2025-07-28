package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/oklog/ulid/v2"

	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/testdata"
)

func StringPtr(s string) *string {
	return &s
}

func TestRequestRepository_AddRequest(t *testing.T) {

	t.Run("test that it saves record", func(t *testing.T) {
		ctx := context.Background()
		defer testdata.ResetTables(inf)

		successResponse := map[string]any{
			"MerchantRequestID":   "ed49-4afa-95de-fde98781ae6b37982271",
			"CheckoutRequestID":   "ws_CO_02062024212422901716772048",
			"ResponseCode":        "0",
			"ResponseDescription": "Success. Request accepted for processing",
			"CustomerMessage":     "Success. Request accepted for processing",
		}

		repo := postgres.NewRequestRepository(inf.Storage.PG)

		record := payments.Request{
			RequestID: ulid.Make().String(),
			Status:    "success",
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
		defer testdata.ResetTables(inf)

		response := map[string]any{
			"ResponseCode": "0",
		}

		repo := postgres.NewRequestRepository(inf.Storage.PG)

		record := payments.Request{
			RequestID: ulid.Make().String(),
			Response:  response,
		}
		err := repo.Add(ctx, record)
		if err != nil {
			t.Errorf("expected non-nil error")
		}

		// fetch record and check values
		r := postgres.RequestSchema{}
		result := inf.Storage.PG.First(&r)
		if result.Error != nil {
			t.Errorf("expected nil error, got %v", result.Error)
		}

		if r.PaymentID != nil {
			t.Errorf("expected nil value, got %v", r.PaymentID)
		}

		// assert values are nil and not empty strings
		AssertNil(t, r.PaymentID)
		AssertNil(t, r.ExternalID)
		AssertNil(t, r.Status)

	})
}

func TestRequestRepository_FindOneRequest(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testdata.ResetTables(inf) })

	paymentsRepo := postgres.NewPaymentsRepository(inf.Storage.PG)
	repo := postgres.NewRequestRepository(inf.Storage.PG)

	t.Run("test that it appends payment details to api request", func(t *testing.T) {

		record := payments.Payment{
			PaymentID:           ulid.Make().String(),
			PaymentReference:    ulid.Make().String(),
			ClientTransactionID: ulid.Make().String(),
			IdempotencyID:       ulid.Make().String(),
		}

		apiRequest := payments.Request{
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
		request, err := repo.FindOneRequest(ctx, payments.OptionsFindOneRequest{RequestID: &apiRequest.RequestID})
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
}

func TestRequestRepository_UpdateRequest(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testdata.ResetTables(inf) })

	repo := postgres.NewRequestRepository(inf.Storage.PG)

	testcases := []struct {
		name  string
		input payments.OptionsUpdateRequest
	}{
		{
			name: "test all values provided",
			input: payments.OptionsUpdateRequest{
				ExternalID: StringPtr(ulid.Make().String()),
				Status:     StringPtr("status"),
				Response:   map[string]any{"key": "value"},
			},
		},
		{
			name:  "test externalID provided",
			input: payments.OptionsUpdateRequest{ExternalID: StringPtr(ulid.Make().String())},
		},
		{
			name:  "test status provided",
			input: payments.OptionsUpdateRequest{Status: StringPtr("status")},
		},
		{
			name: "test response provided",
			input: payments.OptionsUpdateRequest{
				Response: map[string]any{"key": "value"},
			},
		},
		{
			name: "test response map with multiple keys",
			input: payments.OptionsUpdateRequest{
				Response: map[string]any{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
		},
		{
			name: "test response map with nested values",
			input: payments.OptionsUpdateRequest{
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
			req := payments.Request{
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
