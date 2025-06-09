package daraja

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"
)

var (
	ProductionCertificate = strings.TrimSpace(prodCertificate)
	SandboxCertificate    = strings.TrimSpace(sandboxCertificate)

	//go:embed certs/ProductionCertificate.cer
	prodCertificate string
	//go:embed certs/SandboxCertificate.cer
	sandboxCertificate string
)

func encrypt(data, certificate []byte) ([]byte, error) {
	// Decode the PEM encoded public key
	block, _ := pem.Decode(certificate)
	if block == nil {
		return nil, errors.New("failed to decode PEM public key")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.Join(errors.New("failed to parse public key"), err)
	}

	// Type assertion to get RSA public key
	rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}

	// Encrypt the password using RSA
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, data)
	if err != nil {
		return nil, errors.Join(errors.New("encryption error"), err)
	}

	// Convert the encrypted bytes to base64 encoded
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(encrypted)))
	base64.StdEncoding.Encode(buf, encrypted)
	return buf, nil
}

// OpenSSLEncrypt encrypts param data using param certificate
func OpenSSLEncrypt(data, certificate string) (string, error) {
	encrypted, err := encrypt([]byte(data), []byte(certificate))
	if err != nil {
		return "", err
	}
	return string(encrypted), nil
}
