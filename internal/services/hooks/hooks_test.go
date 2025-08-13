package hooks_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/oklog/ulid/v2"
	"github.com/rs/xid"

	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/internal/domains/requests"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/services/hooks"
	"github.com/SirWaithaka/payments-api/internal/testdata"
	"github.com/SirWaithaka/payments-api/request"
)

func TestRequestRecorder_RecordRequest(t *testing.T) {
	defer testdata.ResetTables(inf)

	repository := postgres.NewRequestRepository(inf.Storage.PG)
	paymentsRepo := postgres.NewPaymentsRepository(inf.Storage.PG)
	recorder := hooks.NewRequestRecorder(repository)

	// fake payment
	payment := requests.Payment{
		PaymentID:           ulid.Make().String(),
		ClientTransactionID: ulid.Make().String(),
		IdempotencyID:       ulid.Make().String(),
		PaymentReference:    ulid.Make().String(),
	}
	err := paymentsRepo.AddPayment(t.Context(), payment)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	requestID := xid.New().String()
	hooks := request.Hooks{}
	hooks.Send.PushFrontHook(recorder.RecordRequest(payment.PaymentID, requestID))

	// make request
	cfg := request.Config{ServiceName: "test"}
	req := request.New(cfg, hooks, nil, nil, nil, nil)
	if err := req.Send(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	// fetch record
	rq := postgres.RequestSchema{}
	result := inf.Storage.PG.WithContext(t.Context()).First(&rq)
	if result.Error != nil {
		t.Errorf("expected nil error, got %v", result.Error)
	}

	// assert values
	assert.Equal(t, requestID, rq.RequestID)

}

func TestRequestRecorder_UpdateRequestResponse(t *testing.T) {
	defer testdata.ResetTables(inf)

	// create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("error"))
			return
		case "/timeout":
			time.Sleep(500 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}
	}))
	defer server.Close()

	// create an instance of recorder
	repository := postgres.NewRequestRepository(inf.Storage.PG)
	paymentsRepo := postgres.NewPaymentsRepository(inf.Storage.PG)
	recorder := hooks.NewRequestRecorder(repository)

	// fake payment
	payment := requests.Payment{
		PaymentID:           ulid.Make().String(),
		ClientTransactionID: ulid.Make().String(),
		IdempotencyID:       ulid.Make().String(),
		PaymentReference:    ulid.Make().String(),
	}
	err := paymentsRepo.AddPayment(context.Background(), payment)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	t.Run("test on a success response", func(t *testing.T) {

		requestID := xid.New().String()
		// create request hooks for
		hooks := request.Hooks{}
		hooks.Send.PushFrontHook(recorder.RecordRequest(payment.PaymentID, requestID))
		hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

		// configure the request
		cfg := request.Config{Endpoint: server.URL, DisableSSL: true, HTTPClient: http.DefaultClient, ServiceName: "test"}
		op := &request.Operation{Name: "FooBar", Path: "/"}
		hooks.Send.PushBackHook(corehooks.SendHook)

		// build request
		req := request.New(cfg, hooks, nil, op, nil, nil)
		if err = req.Send(); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// fetch record
		rq := postgres.RequestSchema{}
		result := inf.Storage.PG.Where(postgres.RequestSchema{RequestID: requestID}).First(&rq)
		if result.Error != nil {
			t.Errorf("expected nil error, got %v", result.Error)
		}

		// assert values
		assert.Equal(t, requestID, rq.RequestID)
		assert.Equal(t, "completed", rq.Status)

		if rq.Response == nil {
			t.Errorf("expected non-nil result, got nil")
		}

	})

	t.Run("test on an error response", func(t *testing.T) {

		requestID := xid.New().String()
		// create request hooks for
		hooks := request.Hooks{}
		hooks.Send.PushFrontHook(recorder.RecordRequest(payment.PaymentID, requestID))
		hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

		// configure the request
		cfg := request.Config{Endpoint: server.URL, DisableSSL: true, HTTPClient: http.DefaultClient, ServiceName: "test", RequestID: xid.New().String()}
		op := &request.Operation{Name: "FooBar", Path: "/error"}
		hooks.Send.PushBackHook(corehooks.SendHook)
		hooks.Unmarshal.PushFront(func(r *request.Request) {
			// the mock server returns non 200 status code and "error" in response body
			if r.Response.StatusCode != http.StatusOK {
				// read the response body and assert the value
				body, err := io.ReadAll(r.Response.Body)
				if err != nil {
					r.Error = err
					return
				}

				assert.Equal(t, "error", string(body))
			}
		})

		// build request
		var output string
		req := request.New(cfg, hooks, nil, op, nil, &output)
		if err = req.Send(); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// fetch record
		rq := postgres.RequestSchema{}
		result := inf.Storage.PG.Where(postgres.RequestSchema{RequestID: requestID}).First(&rq)
		if result.Error != nil {
			t.Errorf("expected nil error, got %v", result.Error)
		}

		// assert values
		assert.Equal(t, requestID, rq.RequestID)
		assert.Equal(t, "completed", rq.Status)

		if rq.Response == nil {
			t.Errorf("expected non-nil result, got nil")
		}

	})
}
