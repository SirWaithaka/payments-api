package daraja_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/oklog/ulid/v2"

	"github.com/SirWaithaka/payments-api/clients/daraja"
	"github.com/SirWaithaka/payments-api/internal/pkg/types"
)

// TEST SUITE FOR REQUEST BUILDERS

func TestClient_AuthenticationRequest(t *testing.T) {

	t.Run("test that the request is build correctly", func(t *testing.T) {
		endpoint := "http://foo.bar"
		key := "fake_key"
		secret := "fake_secret"

		client := daraja.New(daraja.Config{Endpoint: endpoint})

		req, _ := client.AuthenticationRequest(key, secret)()

		// check request api url
		url := endpoint + daraja.EndpointAuthentication + "?grant_type=client_credentials"
		assert.Equal(t, req.Request.URL.String(), url)

	})

	t.Run("test that it parses successful response correctly", func(t *testing.T) {
		key := "fake_key"
		secret := "fake_secret"
		acessToken := "fake_token"

		// create a test server
		mux := http.NewServeMux()
		mux.HandleFunc(daraja.EndpointAuthentication, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{"access_token":"%s","expires_in":"3600"}`, acessToken)))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := daraja.New(daraja.Config{Endpoint: server.URL})
		req, res := client.AuthenticationRequest(key, secret)()
		req.WithContext(t.Context())

		// make request to mock server
		if err := req.Send(); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		assert.Equal(t, res.AccessToken, acessToken)

	})

	t.Run("test that it parses error response correctly", func(t *testing.T) {
		key := "fake_key"
		secret := "fake_secret"

		// create a test server
		mux := http.NewServeMux()
		mux.HandleFunc(daraja.EndpointAuthentication, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"requestId":"10101","errorCode":"4000","errorMessage":"Invalid credentials"}`))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := daraja.New(daraja.Config{Endpoint: server.URL})
		req, res := client.AuthenticationRequest(key, secret)()
		req.WithContext(t.Context())

		// make request to mock server
		if err := req.Send(); err == nil {
			t.Errorf("expected non-nil error, got nil")
		}
		t.Log(req.Error)

		// expect res to be empty
		assert.Equal(t, res, &daraja.ResponseAuthorization{})
	})
}

func TestClient_C2BExpressRequest(t *testing.T) {
	endpoint := "http://foo.bar"
	client := daraja.New(daraja.Config{Endpoint: endpoint})

	t.Run("test that the request is built correctly", func(t *testing.T) {
		payload := daraja.RequestC2BExpress{
			BusinessShortCode: "0000000",
			Password:          "fake_password",
			Timestamp:         daraja.NewTimestamp(),
			TransactionType:   "fake_type",
			Amount:            "10",
			PartyA:            "900000",
			PartyB:            "121212",
			PhoneNumber:       "254712345678",
			CallBackURL:       "http://foo.bar/callback",
			AccountReference:  "fake_ref",
			TransactionDesc:   "test payment",
		}
		req, _ := client.C2BExpressRequest(payload)

		// check payload is set in request
		assert.Equal(t, req.Params, payload)
		// check request api url
		url := endpoint + daraja.EndpointC2bExpress
		assert.Equal(t, req.Request.URL.String(), url)
		// check content-type
		assert.Equal(t, req.Request.Header.Get("Content-Type"), "application/json")
	})
}

func TestClient_B2CRequest(t *testing.T) {
	endpoint := "http://foo.bar"
	client := daraja.New(daraja.Config{Endpoint: endpoint})

	t.Run("test that the request is built correctly", func(t *testing.T) {
		payload := daraja.RequestB2C{
			OriginatorConversationID: ulid.Make().String(),
			InitiatorName:            "fake_name",
			SecurityCredential:       "fake_credential",
			CommandID:                "fake_command",
			Amount:                   "10",
			PartyA:                   "0000001",
			PartyB:                   "100002",
			Remarks:                  "test payment",
			QueueTimeOutURL:          "http://foo.bar/timeout",
			ResultURL:                "http://foo.bar/result",
			Occasion:                 "OK",
		}
		req, _ := client.B2CRequest(payload)

		// check payload is set in request
		assert.Equal(t, req.Params, payload)
		// check request api url
		url := endpoint + daraja.EndpointB2cPayment
		assert.Equal(t, req.Request.URL.String(), url)
		// check content-type
		assert.Equal(t, req.Request.Header.Get("Content-Type"), "application/json")
	})
}

func TestClient_B2BRequest(t *testing.T) {
	endpoint := "http://foo.bar"
	client := daraja.New(daraja.Config{Endpoint: endpoint})

	t.Run("test that the request is built correctly", func(t *testing.T) {
		payload := daraja.RequestB2B{
			Initiator:              "fake_initiator",
			SecurityCredential:     "fake_credential",
			CommandID:              "fake_command",
			SenderIdentifierType:   "fake_type",
			RecieverIdentifierType: "fake_type",
			Amount:                 "10",
			PartyA:                 "000100",
			PartyB:                 "100300",
			AccountReference:       "fake_ref",
			Remarks:                "test payment",
			QueueTimeOutURL:        "http://foo.bar/timeout",
			ResultURL:              "http://foo.bar/result",
		}
		req, _ := client.B2BRequest(payload)

		// check payload is set in request
		assert.Equal(t, req.Params, payload)
		// check request api url
		url := endpoint + daraja.EndpointB2bPayment
		assert.Equal(t, req.Request.URL.String(), url)
		// check content-type
		assert.Equal(t, req.Request.Header.Get("Content-Type"), "application/json")
	})
}

func TestClient_ReversalRequest(t *testing.T) {
	endpoint := "http://foo.bar"
	client := daraja.New(daraja.Config{Endpoint: endpoint})

	t.Run("test that the request is built correctly", func(t *testing.T) {
		payload := daraja.RequestReversal{
			Initiator:              "fake_initiator",
			SecurityCredential:     "fake_credential",
			CommandID:              "fake_command",
			TransactionID:          "fake_id",
			Amount:                 "10",
			ReceiverParty:          "000100",
			ReceiverIdentifierType: "fake_type",
			ResultURL:              "http://foo.bar/result",
			QueueTimeOutURL:        "http://foo.bar/timeout",
		}
		req, _ := client.ReversalRequest(payload)

		// check payload is set in request
		assert.Equal(t, req.Params, payload)
		// check request api url
		url := endpoint + daraja.EndpointReversal
		assert.Equal(t, req.Request.URL.String(), url)
		// check content-type
		assert.Equal(t, req.Request.Header.Get("Content-Type"), "application/json")
	})
}

func TestClient_TransactionStatusRequest(t *testing.T) {
	endpoint := "http://foo.bar"
	client := daraja.New(daraja.Config{Endpoint: endpoint})

	t.Run("test that the request is built correctly", func(t *testing.T) {
		payload := daraja.RequestTransactionStatus{
			Initiator:                "fake_initiator",
			SecurityCredential:       "fake_credential",
			CommandID:                "fake_command",
			TransactionID:            types.Pointer(ulid.Make().String()),
			OriginatorConversationID: types.Pointer(ulid.Make().String()),
			PartyA:                   "200200",
			IdentifierType:           "fake_type",
			ResultURL:                "http://foo.bar/result",
			QueueTimeOutURL:          "http://foo.bar/timeout",
			Remarks:                  "test payment",
			Occasion:                 "OK",
		}
		req, _ := client.TransactionStatusRequest(payload)

		// check payload is set in request
		assert.Equal(t, req.Params, payload)
		// check request api url
		url := endpoint + daraja.EndpointTransactionStatus
		assert.Equal(t, req.Request.URL.String(), url)
		// check content-type
		assert.Equal(t, req.Request.Header.Get("Content-Type"), "application/json")
	})
}

func TestClient_BalanceRequest(t *testing.T) {
	endpoint := "http://foo.bar"
	client := daraja.New(daraja.Config{Endpoint: endpoint})

	t.Run("test that the request is built correctly", func(t *testing.T) {
		payload := daraja.RequestBalance{
			Initiator:          "fake_initiator",
			SecurityCredential: "fake_credential",
			CommandID:          "fake_command",
			PartyA:             "200200",
			IdentifierType:     "fake_type",
			ResultURL:          "http://foo.bar/result",
			QueueTimeOutURL:    "http://foo.bar/timeout",
			Remarks:            "test payment",
		}
		req, _ := client.BalanceRequest(payload)

		// check payload is set in reequest
		assert.Equal(t, req.Params, payload)
		// check request api url
		url := endpoint + daraja.EndpointAccountBalance
		assert.Equal(t, req.Request.URL.String(), url)
		// check content-type
		assert.Equal(t, req.Request.Header.Get("Content-Type"), "application/json")
	})
}

func TestClient_QueryOrgInfoRequest(t *testing.T) {
	endpoint := "http://foo.bar"
	client := daraja.New(daraja.Config{Endpoint: endpoint})

	t.Run("test that the request is built correctly", func(t *testing.T) {
		payload := daraja.RequestOrgInfoQuery{
			IdentifierType: "fake_type",
			Identifier:     "fake_identifier",
		}
		req, _ := client.QueryOrgInfoRequest(payload)

		// check payload is set in reequest
		assert.Equal(t, req.Params, payload)
		// check request api url
		url := endpoint + daraja.EndpointQueryOrgInfo
		assert.Equal(t, req.Request.URL.String(), url)
		// check content-type
		assert.Equal(t, req.Request.Header.Get("Content-Type"), "application/json")
	})
}

// TEST SUITES FOR REQUEST EXECUTORS
