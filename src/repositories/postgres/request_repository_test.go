package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"

	pkgerrors "github.com/SirWaithaka/payments-api/pkg/errors"
	"github.com/SirWaithaka/payments-api/pkg/types"
	"github.com/SirWaithaka/payments-api/src/domains/requests"
	"github.com/SirWaithaka/payments-api/src/repositories/postgres"
	"github.com/SirWaithaka/payments-api/testdata"
)

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

func TestRequestRepository_FindOne(t *testing.T) {
	ctx := context.Background()

	repo := postgres.NewRequestRepository(inf.Storage.PG)

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

		req, err := repo.FindOne(t.Context(), requests.OptionsFindRequest{ExternalID: &apiRequest.ExternalID})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		assert.Equal(t, apiRequest.RequestID, req.RequestID)
		assert.Equal(t, apiRequest.ExternalID, req.ExternalID)

	})

	t.Run("test that it finds by request id", func(t *testing.T) {
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

		req, err := repo.FindOne(t.Context(), requests.OptionsFindRequest{RequestID: &apiRequest.RequestID})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		assert.Equal(t, apiRequest.RequestID, req.RequestID)
		assert.Equal(t, apiRequest.ExternalID, req.ExternalID)

	})

	t.Run("test that it finds by payment id", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		apiRequest := requests.Request{
			RequestID:  ulid.Make().String(),
			ExternalID: ulid.Make().String(),
			PaymentID:  ulid.Make().String(),
			Partner:    "fake_partner",
			Status:     "received",
			Latency:    1000,
			Response:   nil,
		}
		err := repo.Add(ctx, apiRequest)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		req, err := repo.FindOne(t.Context(), requests.OptionsFindRequest{PaymentID: &apiRequest.PaymentID})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		assert.Equal(t, apiRequest.RequestID, req.RequestID)
		assert.Equal(t, apiRequest.ExternalID, req.ExternalID)

	})

	t.Run("test that it returns not found error if record does not exist", func(t *testing.T) {
		// fetch a non-existent record
		req, err := repo.FindOne(t.Context(), requests.OptionsFindRequest{RequestID: types.Pointer(ulid.Make().String())})
		if err == nil {
			t.Errorf("expected non-nil error")
		}

		// check err implements pkgerrors.NotFoundError
		e, ok := err.(pkgerrors.NotFounder)
		assert.True(t, ok)
		assert.True(t, e.NotFound())
		assert.Empty(t, req)
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
				ExternalID: types.Pointer(ulid.Make().String()),
				Status:     types.Pointer(requests.StatusSucceeded),
				Response:   map[string]any{"key": "value"},
			},
		},
		{
			name:  "test externalID provided",
			input: requests.OptionsUpdateRequest{ExternalID: types.Pointer(ulid.Make().String())},
		},
		{
			name:  "test status provided",
			input: requests.OptionsUpdateRequest{Status: types.Pointer(requests.StatusSucceeded)},
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
			if tc.input.Status == nil {
				assert.Nil(t, record.Status)
			} else {
				assert.Equal(t, tc.input.Status.String(), *record.Status)
			}
			assert.Equal(t, tc.input.ExternalID, record.ExternalID)
			assert.Equal(t, tc.input.Response, record.ToEntity().Response)

		})

	}

}
