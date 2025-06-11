package daraja_test

import (
	"testing"

	"github.com/SirWaithaka/payments-api/clients/daraja"
)

func TestAuthenticationRequest(t *testing.T) {
	endpoint := "http://foo.bar"
	key := "fake_key"
	secret := "fake_secret"

	req, res := daraja.AuthenticationRequest(endpoint, key, secret)

	// request uri has grant type auth parameter
	e := endpoint + daraja.EndpointAuthentication + "?grant_type=client_credentials"
	if v := req.Request.URL.String(); v != e {
		t.Errorf("expected %s, got %s", e, v)
	}

	// check client is configured in request
	if req.Config.HTTPClient == nil {
		t.Errorf("expected http client to be non-nil")
	}

	// check that response is not nil
	if res == nil {
		t.Error("expected response, got nil")
	}
}
