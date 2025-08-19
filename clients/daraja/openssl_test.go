package daraja_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SirWaithaka/payments-api/clients/daraja"
)

func TestOpenSSLEncrypt(t *testing.T) {

	data := "test string to encrypt"
	// encrypt with the sandbox certificate
	encrypted, err := daraja.OpenSSLEncrypt(data, daraja.SandboxCertificate)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	// encrypted should not be the same as the original data
	assert.NotEqual(t, data, encrypted)
	// encrypted should not be an empty string
	assert.NotEmpty(t, encrypted)

	// encrypt with the production certificate
	encrypted, err = daraja.OpenSSLEncrypt(data, daraja.ProductionCertificate)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	// encrypted should not be the same as the original data
	assert.NotEqual(t, data, encrypted)
	// encrypted should not be an empty string
	assert.NotEmpty(t, encrypted)

}
