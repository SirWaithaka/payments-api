package main

import (
	"log"
	"net/http"

	"github.com/SirWaithaka/payments-api/clients/daraja"
	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/request"
)

// this example demonstrates an example of calling the authentication api on daraja

func main() {
	key := "fake_key"
	secret := "fake_secret"

	hooks := request.Hooks{}
	// needs a hook to set basic auth headers
	hooks.Build.PushFrontHook(corehooks.SetBasicAuth(key, secret))
	// needs a hook to unmarshal the response
	// the daraja auth endpoint returns a json response
	hooks.Unmarshal.PushBackHook(daraja.ResponseDecoder)

	// configure the request method and path
	op := &request.Operation{Method: http.MethodGet, Path: "/oauth/v1/generate?grant_type=client_credentials"}
	// model where the response will be unmarshalled to
	res := &daraja.ResponseAuthorization{}
	// create an instance of request
	req := request.New(request.Config{Endpoint: daraja.SandboxUrl}, hooks, nil, op, nil, res)
	// make request
	if err := req.Send(); err != nil {
		log.Println(err)
		return
	}
	// print response
	log.Println(res)
}
