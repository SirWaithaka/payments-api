package daraja

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
