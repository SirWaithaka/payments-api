package quikk

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/request"
)

func TestSign(t *testing.T) {
	// test that it adds the Authorization and Date headers

	// create a mock test server
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		assert.NotEqual(t, r.Header.Get("Authorization"), "")
		assert.NotEqual(t, r.Header.Get("Date"), "")

		// send fake response
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	key := "fake_key"
	secret := "fake_secret"

	// build a test request
	cfg := request.Config{Endpoint: server.URL}
	hooks := DefaultHooks()
	hooks.Build.PushFrontHook(Sign(key, secret)) // add sign hook
	hooks.Unmarshal.Clear()                      // since it's a test, remove any response decoder
	op := &request.Operation{Name: "test", Path: "/test"}
	req := request.New(cfg, hooks, nil, op, nil, nil)

	if err := req.Send(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

}

func TestResponseDecoder(t *testing.T) {

	type successResponse struct {
		Status string `json:"status"`
	}
	type errorResponse struct {
		Error string `json:"error"`
	}

	// create a mock test server
	mux := http.NewServeMux()
	mux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		// send fake response
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success"}`))
	})
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		// send fake response
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errors":[{"status":"failed"}]}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	t.Run("test that it decodes response in success status", func(t *testing.T) {
		// build a test request
		// add response decoder hook
		hooks := corehooks.DefaultHooks()
		hooks.Unmarshal.PushFrontHook(ResponseDecoder)
		// build request
		cfg := request.Config{Endpoint: server.URL}
		op := &request.Operation{Name: "test", Path: "/success"}
		data := &successResponse{}
		req := request.New(cfg, hooks, nil, op, nil, data)

		if err := req.Send(); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		assert.NotEqual(t, data.Status, "")

	})

	t.Run("test that it decodes response in error status ", func(t *testing.T) {
		// build a test request
		// add response decoder hook
		hooks := corehooks.DefaultHooks()
		hooks.Unmarshal.PushFrontHook(ResponseDecoder)
		// build request
		cfg := request.Config{Endpoint: server.URL}
		op := &request.Operation{Name: "test", Path: "/error"}
		req := request.New(cfg, hooks, nil, op, nil, nil)

		if err := req.Send(); err == nil {
			t.Errorf("expected error, got nil")
		}

	})
}
