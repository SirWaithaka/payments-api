package hooks_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"

	"github.com/SirWaithaka/gorequest"
	"github.com/SirWaithaka/gorequest/corehooks"

	"github.com/SirWaithaka/payments-api/src/domains/requests"
	"github.com/SirWaithaka/payments-api/src/repositories/postgres"
	"github.com/SirWaithaka/payments-api/src/services/hooks"
	"github.com/SirWaithaka/payments-api/testdata"
)

func TestRequestRecorder_RecordRequest(t *testing.T) {
	defer testdata.ResetTables(inf)

	repository := postgres.NewRequestRepository(inf.Storage.PG)
	recorder := hooks.NewRequestRecorder(repository)

	// fake payment
	paymentID := ulid.Make().String()

	requestID := xid.New().String()
	reqHooks := gorequest.Hooks{}
	reqHooks.Send.PushFrontHook(recorder.RecordRequest(paymentID, requestID))

	// make request
	cfg := gorequest.Config{ServiceName: "test"}
	req := gorequest.New(cfg, gorequest.Operation{}, reqHooks, nil, nil, nil)
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
	recorder := hooks.NewRequestRecorder(repository)

	t.Run("test on a success response", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		// fake payment
		paymentID := ulid.Make().String()

		requestID := xid.New().String()
		// create request hooks for
		reqHooks := gorequest.Hooks{}
		reqHooks.Send.PushFrontHook(recorder.RecordRequest(paymentID, requestID))
		reqHooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

		// configure the request
		cfg := gorequest.Config{Endpoint: server.URL, DisableSSL: true, HTTPClient: http.DefaultClient, ServiceName: "test"}
		op := gorequest.Operation{Name: "FooBar", Path: "/"}
		reqHooks.Send.PushBackHook(corehooks.SendHook)

		// build request
		req := gorequest.New(cfg, op, reqHooks, nil, nil, nil)
		if err := req.Send(); err != nil {
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
		assert.Equal(t, requests.StatusSucceeded.String(), *rq.Status)

		if rq.Response == nil {
			t.Errorf("expected non-nil result, got nil")
		}

	})

	t.Run("test on an error response", func(t *testing.T) {
		defer testdata.ResetTables(inf)

		// fake payment
		paymentID := ulid.Make().String()

		requestID := xid.New().String()
		// create request hooks for
		reqHooks := gorequest.Hooks{}
		reqHooks.Send.PushFrontHook(recorder.RecordRequest(paymentID, requestID))
		reqHooks.Complete.PushFrontHook(recorder.UpdateRequestResponse(requestID))

		// configure the request
		cfg := gorequest.Config{Endpoint: server.URL, DisableSSL: true, HTTPClient: http.DefaultClient, ServiceName: "test", RequestID: xid.New().String()}
		op := gorequest.Operation{Name: "FooBar", Path: "/error"}
		reqHooks.Send.PushBackHook(corehooks.SendHook)
		reqHooks.Unmarshal.PushFront(func(r *gorequest.Request) {
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
		req := gorequest.New(cfg, op, reqHooks, nil, nil, &output)
		if err := req.Send(); err != nil {
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
		assert.Equal(t, requests.StatusSucceeded.String(), *rq.Status)

		if rq.Response == nil {
			t.Errorf("expected non-nil result, got nil")
		}

	})
}
