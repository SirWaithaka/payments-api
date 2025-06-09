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

func main() {
	l := log.New(os.Stdout, "Daraja: ", log.LstdFlags|log.Llongfile)

	sCfg := ShortCodeConfig{
		ShortCode:         "000000",
		InitiatorName:     "testapi",
		InitiatorPassword: "Safaricom999!*!%",
		Passphrase:        "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72ada1ed2c919%",
		ConsumerKey:       "GW0TvN2gUTakps3b1AbAw48no1Yogu92oXI0N55fmlEVK40e",
		ConsumerSecret:    "CtpMOjvk47jm6A5hmCzaQjQTBOWAwK1LM95awGNLSTGawbGNPsy9f8Eabsr1Lg7P",
	}

	endpoint := "http://localhost:9002/daraja"
	darajaclient := daraja.New(daraja.Config{Endpoint: endpoint})
	darajaclient.Hooks.Build.PushBackHook(daraja.Authenticate(endpoint, sCfg.InitiatorName, sCfg.InitiatorPassword))

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
		CallBackURL:       "http://localhost:9002/daraja/c2b/callback",
		AccountReference:  "F0000020",
		TransactionDesc:   "Customer Deposit",
	}

	res, err := darajaclient.C2BExpress(context.Background(), req)
	if err != nil {
		l.Fatal(err)
	}
	l.Println(res)
}
