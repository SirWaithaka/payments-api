package quikk_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/rs/xid"

	"github.com/SirWaithaka/payments-api/clients/quikk"
)

// TEST SUITES FOR REQUEST BUILDERS

func TestClient_ChargeRequest(t *testing.T) {
	endpoint := "http://foo.bar"

	requestID := xid.New().String()
	client := quikk.New(quikk.Config{Endpoint: endpoint})
	req, _ := client.ChargeRequest(quikk.RequestCharge{}, requestID)

	// check endpoint
	assert.Equal(t, req.Request.URL.String(), endpoint+quikk.EndpointCharge)
}

func TestClient_PayoutRequest(t *testing.T) {
	endpoint := "http://foo.bar"

	requestID := xid.New().String()
	client := quikk.New(quikk.Config{Endpoint: endpoint})
	req, _ := client.PayoutRequest(quikk.RequestPayout{}, requestID)

	// check endpoint
	assert.Equal(t, req.Request.URL.String(), endpoint+quikk.EndpointPayout)
}

func TestClient_TransferRequest(t *testing.T) {
	endpoint := "http://foo.bar"

	requestID := xid.New().String()
	client := quikk.New(quikk.Config{Endpoint: endpoint})
	req, _ := client.TransferRequest(quikk.RequestTransfer{}, requestID)

	// check endpoint
	assert.Equal(t, req.Request.URL.String(), endpoint+quikk.EndpointTransfer)
}

func TestClient_TransactionSearchRequest(t *testing.T) {
	endpoint := "http://foo.bar"

	requestID := xid.New().String()
	client := quikk.New(quikk.Config{Endpoint: endpoint})
	req, _ := client.TransactionSearchRequest(quikk.RequestTransactionStatus{}, requestID)

	// check endpoint
	assert.Equal(t, req.Request.URL.String(), endpoint+quikk.EndpointTransactionSearch)
}

func TestClient_BalanceRequest(t *testing.T) {
	endpoint := "http://foo.bar"

	requestID := xid.New().String()
	client := quikk.New(quikk.Config{Endpoint: endpoint})
	req, _ := client.BalanceRequest(quikk.RequestAccountBalance{}, requestID)

	// check endpoint
	assert.Equal(t, req.Request.URL.String(), endpoint+quikk.EndpointBalance)
}

// TEST SUITES FOR REQUEST EXECUTORS

func TestClient_TransactionSearch(t *testing.T) {
	resourceID := xid.New().String()

	// create a mock test server
	mux := http.NewServeMux()
	mux.HandleFunc(quikk.EndpointTransactionSearch, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"data":{"id":"12345","type":"search","attributes":{"resource_id":"%s"}}}`, resourceID)))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := quikk.New(quikk.Config{Endpoint: server.URL})
	res, err := client.TransactionSearch(t.Context(), quikk.RequestTransactionStatus{}, xid.New().String())
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	assert.Equal(t, res.Data.Attributes.ResourceID, resourceID)
}
