package main

import (
	"context"
	"os"

	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/clients/daraja"
)

func main() {
	l := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	zerolog.DefaultContextLogger = &l

	darajaclient := daraja.New()

	req := daraja.RequestC2BExpress{
		BusinessShortCode: "000000",
		Password:          "mnbvcxz", // daraja password should be encrypted using the public certificate
		Timestamp:         daraja.NewTimestamp(),
		TransactionType:   daraja.OperationC2BExpress,
		Amount:            "100",
		PartyA:            "100100",
		PartyB:            "000000",
		PhoneNumber:       "0720000000",
		CallBackURL:       "http://localhost:9002/daraja/c2b/callback",
		AccountReference:  "F0000020",
		TransactionDesc:   "Customer Deposit",
	}

	res, err := darajaclient.C2BExpress(context.Background(), req)
	if err != nil {
		l.Err(err).Send()
		return
	}
	l.Info().Any("response", res).Send()
}
