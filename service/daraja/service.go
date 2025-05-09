package daraja

import (
	"context"
	"net/http"
	"time"

	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/request"
)

const (
	OperationC2BExpress        = "express"
	OperationC2BQuery          = "stk_query"
	OperationReversal          = "reversal"
	OperationB2C               = "b2c"
	OperationB2B               = "b2b"
	OperationTransactionStatus = "search"
	OperationBalance           = "balance"
	OperationQueryOrgInfo      = "org_info_query"
)

func AuthenticationRequest(endpoint, key, secret string) (*request.Request, *ResponseAuthorization) {
	op := &request.Operation{
		Name:   "Authenticate",
		Method: http.MethodGet,
		Path:   EndpointAuthentication + "?grant_type=client_credentials",
	}

	cfg := request.Config{Endpoint: endpoint}

	// create a client with 40 second timeout
	client := &http.Client{Timeout: time.Second * 40}
	// default hooks
	hooks := corehooks.DefaultHooks()
	hooks.Build.PushBackHook(HTTPClient(client))
	hooks.Build.PushBackHook(corehooks.SetBasicAuth(key, secret))
	hooks.Build.PushBackHook(corehooks.EncodeRequestBody)
	hooks.Unmarshal.PushBackHook(DecodeResponse())

	output := &ResponseAuthorization{}
	req := request.New(cfg, hooks, op, nil, output)

	return req, output
}

// Daraja provides the API operation methods for making requests
// to MPESA Daraja service.
type Daraja struct {
	endpoint string
	hooks    request.Hooks
}

func New() Daraja {
	endpoint := "http://localhost:9002/daraja"

	// create default hooks
	hooks := corehooks.DefaultHooks()

	hooks.Build.PushBackHook(HTTPClient(nil))
	hooks.Build.PushBackHook(corehooks.EncodeRequestBody)
	hooks.Build.PushBackHook(Authenticate(endpoint, "fake_key", "fake_secret"))
	hooks.Unmarshal.PushBackHook(DecodeResponse())

	return Daraja{hooks: hooks, endpoint: endpoint}
}

func (daraja *Daraja) Hooks() request.Hooks {
	return daraja.hooks
}

func (daraja Daraja) C2BExpressRequest(input RequestC2BExpress, opts ...request.Option) (*request.Request, *ResponseC2BExpress) {
	op := &request.Operation{
		Name:   OperationC2BExpress,
		Method: http.MethodPost,
		Path:   EndpointC2bExpress,
	}

	cfg := request.Config{Endpoint: daraja.endpoint}

	output := &ResponseC2BExpress{}
	req := request.New(cfg, daraja.hooks, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (daraja Daraja) C2BExpress(ctx context.Context, payload RequestC2BExpress) (ResponseC2BExpress, error) {
	req, out := daraja.C2BExpressRequest(payload, request.WithRequestHeader("Content-Type", "application/json"))
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseC2BExpress{}, err
	}

	return *out, nil
}
