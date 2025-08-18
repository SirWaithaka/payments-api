package daraja

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"

	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/request"
)

func TestAuthenticate(t *testing.T) {

	t.Run("test that authenticate hook sets the access token", func(t *testing.T) {
		key := "fake_key"
		secret := "fake_secret"
		acessToken := "fake_token"

		// create a test server
		mux := http.NewServeMux()
		mux.HandleFunc(EndpointAuthentication, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{"access_token":"%s","expires_in":"3600"}`, acessToken)))
		})
		mux.HandleFunc(EndpointQueryOrgInfo, func(w http.ResponseWriter, r *http.Request) {
			// check the access token is set
			assert.Equal(t, r.Header.Get("Authorization"), fmt.Sprintf("Bearer %s", acessToken))
			// check request is successful
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			// send fake response
			w.Write([]byte(`{"ResponseMessage":"Success","ResponseCode":"00"}`))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := New(Config{Endpoint: server.URL})
		// add authenticate hook to client build hooks
		client.Hooks.Build.PushFrontHook(Authenticate(client.AuthenticationRequest(key, secret)))

		// attempt request
		_, err := client.QueryOrgInfo(t.Context(), RequestOrgInfoQuery{})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("test that it reuses cached token for subsequent requests", func(t *testing.T) {
		key := "fake_key"
		secret := "fake_secret"
		acessToken := "fake_token"

		authCalls := 0

		// create a test server
		mux := http.NewServeMux()
		mux.HandleFunc(EndpointAuthentication, func(w http.ResponseWriter, r *http.Request) {
			authCalls++
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{"access_token":"%s","expires_in":"3600"}`, acessToken)))
		})
		mux.HandleFunc(EndpointQueryOrgInfo, func(w http.ResponseWriter, r *http.Request) {
			// check the access token is set
			assert.Equal(t, r.Header.Get("Authorization"), fmt.Sprintf("Bearer %s", acessToken))
			// check request is successful
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			// send fake response
			w.Write([]byte(`{"ResponseMessage":"Success","ResponseCode":"00"}`))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := New(Config{Endpoint: server.URL})
		// add authenticate hook to client build hooks
		client.Hooks.Build.PushFrontHook(Authenticate(client.AuthenticationRequest(key, secret)))

		// attempt request 1
		_, err := client.QueryOrgInfo(t.Context(), RequestOrgInfoQuery{})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// attempt request 2
		_, err = client.QueryOrgInfo(t.Context(), RequestOrgInfoQuery{})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		// assert that authentication endpoint was called only once
		assert.Equal(t, 1, authCalls)
	})
}

func TestResponseDecoder(t *testing.T) {

	requestID := ulid.Make().String()
	mux := http.NewServeMux()
	mux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		// mock successful request
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"ResponseMessage":"Success","ResponseCode":"0","MerchantRequestID":"%s","CheckoutRequestID":"%s"}`, requestID, requestID)))
	})
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		// mock request failure
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"requestId":"%s","errorCode":"23","errorMessage":"test failure"}`, requestID)))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	t.Run("test non-200 response", func(t *testing.T) {
		// build request
		cfg := request.Config{Endpoint: server.URL}
		op := &request.Operation{Name: "test", Path: "/error"}
		hooks := corehooks.DefaultHooks()
		hooks.Unmarshal.PushFrontHook(ResponseDecoder)
		req := request.New(cfg, hooks, nil, op, nil, nil)

		err := req.Send()
		if err == nil {
			t.Errorf("expected error, got nil")
		}

		// err should be convertible to client_daraja.ErrorResponse
		v := reflect.ValueOf(err)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		// targetType
		targetType := reflect.TypeOf(ErrorResponse{})
		// check can be converted to targetType
		if !v.Type().ConvertibleTo(targetType) {
			t.Errorf("expected %v to be convertible to %v", v.Type(), targetType)
		}
	})
}
