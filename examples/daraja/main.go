package main

import (
	"context"
	"log"
	"os"

	"github.com/SirWaithaka/payments-api/clients/daraja"
)

type ShortCodeConfig struct {
	ShortCode         string // business pay bill or buy goods account
	InitiatorName     string // daraja api initiator name
	InitiatorPassword string // daraja api initiator password
	Passphrase        string // (optional) passphrase for c2b transfers
	ConsumerKey       string // daraja app consumer key
	ConsumerSecret    string // daraja app consumer secret
}

// example of making a c2b request
func makeC2bRequest(sCfg ShortCodeConfig) {
	l := log.New(os.Stdout, "Daraja: ", log.LstdFlags|log.Llongfile)

	// create an instance of daraja client
	client := daraja.New(daraja.Config{Endpoint: daraja.SandboxUrl})
	// configure authentication using request hooks
	client.Hooks.Build.PushBackHook(daraja.Authenticate(client.AuthenticationRequest(sCfg.ConsumerKey, sCfg.ConsumerSecret)))

	// encode the shortcode passphrase
	password := daraja.PasswordEncode(sCfg.ShortCode, sCfg.Passphrase, daraja.NewTimestamp().String())
	req := daraja.RequestC2BExpress{
		BusinessShortCode: sCfg.ShortCode,
		Password:          password, // encoded passphrase for c2b
		Timestamp:         daraja.NewTimestamp(),
		TransactionType:   daraja.OperationC2BExpress,
		Amount:            "100",
		PartyA:            "100100",
		PartyB:            sCfg.ShortCode,
		PhoneNumber:       "0720000000",
		CallBackURL:       "http://localhost:8000/daraja/c2b/callback",
		AccountReference:  "F0000020",
		TransactionDesc:   "Customer Deposit",
	}

	// make the stk push request through daraja
	res, err := client.C2BExpress(context.Background(), req)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(res)
}

func makeB2cRequest(sCfg ShortCodeConfig) {
	l := log.New(os.Stdout, "Daraja: ", log.LstdFlags|log.Llongfile)

	// create an instance of daraja client
	client := daraja.New(daraja.Config{Endpoint: daraja.SandboxUrl})
	// configure authentication using request hooks
	client.Hooks.Build.PushBackHook(daraja.Authenticate(client.AuthenticationRequest(sCfg.ConsumerKey, sCfg.ConsumerSecret)))

	// build security credential
	credential, err := daraja.OpenSSLEncrypt(sCfg.InitiatorPassword, daraja.SandboxCertificate)
	if err != nil {
		l.Fatal(err)
	}

	req := daraja.RequestB2C{
		OriginatorConversationID: "1234567890",
		InitiatorName:            "initiator_name",
		SecurityCredential:       credential,
		CommandID:                daraja.CommandBusinessPayment,
		Amount:                   "10",
		PartyA:                   sCfg.ShortCode,
		PartyB:                   "254712345678",
		Remarks:                  "test payment",
		QueueTimeOutURL:          "http://localhost:8000/daraja/b2c/callback",
		ResultURL:                "http://localhost:8000/daraja/b2c/callback",
		Occasion:                 "OK",
	}
	// make the b2c request through daraja
	res, err := client.B2C(context.Background(), req)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(res)
}

func main() {

	// daraja shortcode config
	sCfg := ShortCodeConfig{
		ShortCode:         "000000",
		InitiatorName:     "initiator_name",
		InitiatorPassword: "initiator_password",
		Passphrase:        "passphrase",
		ConsumerKey:       "consumer_key",
		ConsumerSecret:    "consumer_secret",
	}

	makeC2bRequest(sCfg)
	makeB2cRequest(sCfg)

}
