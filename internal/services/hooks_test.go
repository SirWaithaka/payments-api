package services_test

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
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	"github.com/SirWaithaka/payments-api/internal/repositories/postgres"
	"github.com/SirWaithaka/payments-api/internal/services"
	"github.com/SirWaithaka/payments-api/internal/testdata"
	"github.com/SirWaithaka/payments-api/request"
)

func TestRequestRecorder_RecordRequest(t *testing.T) {
	defer testdata.ResetTables(inf)

	repository := postgres.NewRequestRepository(inf.Storage.PG)
	paymentsRepo := postgres.NewPaymentsRepository(inf.Storage.PG)
	recorder := services.NewRequestRecorder(repository)

	// fake payment
	payment := payments.Payment{
		PaymentID:           ulid.Make().String(),
		ClientTransactionID: ulid.Make().String(),
		IdempotencyID:       ulid.Make().String(),
		PaymentReference:    ulid.Make().String(),
	}
	err := paymentsRepo.AddPayment(t.Context(), payment)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	hooks := request.Hooks{}
	hooks.Build.PushFront(func(r *request.Request) {
		r.Config.RequestID = xid.New().String()
	})
	hooks.Send.PushFrontHook(recorder.RecordRequest(payment.PaymentID))

	// make request
	cfg := request.Config{RequestID: xid.New().String(), ServiceName: "test"}
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
	assert.Equal(t, req.Config.RequestID, rq.RequestID)

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
	recorder := services.NewRequestRecorder(repository)

	// fake payment
	payment := payments.Payment{
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

		// create request hooks for
		hooks := request.Hooks{}
		hooks.Build.PushFront(func(r *request.Request) {
			r.Config.RequestID = xid.New().String()
		})
		hooks.Send.PushFrontHook(recorder.RecordRequest(payment.PaymentID))
		hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse())

		// configure the request
		cfg := request.Config{Endpoint: server.URL, DisableSSL: true, HTTPClient: http.DefaultClient, ServiceName: "test", RequestID: xid.New().String()}
		op := &request.Operation{Name: "FooBar", Path: "/"}
		hooks.Send.PushBackHook(corehooks.SendHook)

		// build request
		req := request.New(cfg, hooks, nil, op, nil, nil)
		if err = req.Send(); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// fetch record
		rq := postgres.RequestSchema{}
		result := inf.Storage.PG.Where(postgres.RequestSchema{RequestID: req.Config.RequestID}).First(&rq)
		if result.Error != nil {
			t.Errorf("expected nil error, got %v", result.Error)
		}

		// assert values
		assert.Equal(t, req.Config.RequestID, rq.RequestID)
		assert.Equal(t, "completed", rq.Status)

		if rq.Response == nil {
			t.Errorf("expected non-nil result, got nil")
		}

	})

	t.Run("test on an error response", func(t *testing.T) {

		// create request hooks for
		hooks := request.Hooks{}
		hooks.Build.PushFront(func(r *request.Request) {
			r.Config.RequestID = xid.New().String()
		})
		hooks.Send.PushFrontHook(recorder.RecordRequest(payment.PaymentID))
		hooks.Complete.PushFrontHook(recorder.UpdateRequestResponse())

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
		result := inf.Storage.PG.Where(postgres.RequestSchema{RequestID: req.Config.RequestID}).First(&rq)
		if result.Error != nil {
			t.Errorf("expected nil error, got %v", result.Error)
		}

		// assert values
		assert.Equal(t, req.Config.RequestID, rq.RequestID)
		assert.Equal(t, "completed", rq.Status)

		if rq.Response == nil {
			t.Errorf("expected non-nil result, got nil")
		}

	})
}
