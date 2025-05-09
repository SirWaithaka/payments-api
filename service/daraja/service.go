package daraja

import (
	"context"
	"encoding/base64"
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

func PasswordEncode(shortcode, passphrase, timestamp string) string {
	return base64.StdEncoding.EncodeToString([]byte(shortcode + passphrase + timestamp))
}

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

func (daraja Daraja) ReversalRequest(input RequestReversal, opts ...request.Option) (*request.Request, *ResponseReversal) {
	op := &request.Operation{
		Name:   OperationReversal,
		Method: http.MethodPost,
		Path:   EndpointReversal,
	}

	cfg := request.Config{Endpoint: daraja.endpoint}
	output := &ResponseReversal{}

	req := request.New(cfg, daraja.hooks, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (daraja Daraja) Reverse(ctx context.Context, payload RequestReversal) (ResponseReversal, error) {
	req, out := daraja.ReversalRequest(payload, request.WithRequestHeader("Content-Type", "application/json"))
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseReversal{}, err
	}

	return *out, nil
}

func (daraja Daraja) B2CRequest(input RequestB2C, opts ...request.Option) (*request.Request, *ResponseB2C) {
	op := &request.Operation{
		Name:   OperationB2C,
		Method: http.MethodPost,
		Path:   EndpointB2cPayment,
	}

	cfg := request.Config{Endpoint: daraja.endpoint}
	output := &ResponseB2C{}

	req := request.New(cfg, daraja.hooks, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (daraja Daraja) B2C(ctx context.Context, payload RequestB2C) (ResponseB2C, error) {
	req, out := daraja.B2CRequest(payload, request.WithRequestHeader("Content-Type", "application/json"))
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseB2C{}, err
	}

	return *out, nil
}

func (daraja Daraja) B2BRequest(input RequestB2B, opts ...request.Option) (*request.Request, *ResponseB2B) {
	op := &request.Operation{
		Name:   OperationB2B,
		Method: http.MethodPost,
		Path:   EndpointB2bPayment,
	}

	cfg := request.Config{Endpoint: daraja.endpoint}
	output := &ResponseB2B{}

	req := request.New(cfg, daraja.hooks, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (daraja Daraja) B2B(ctx context.Context, payload RequestB2B) (ResponseB2B, error) {
	req, out := daraja.B2BRequest(payload, request.WithRequestHeader("Content-Type", "application/json"))
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseB2B{}, err
	}

	return *out, nil
}

func (daraja Daraja) TransactionStatusRequest(input RequestTransactionStatus, opts ...request.Option) (*request.Request, *ResponseTransactionStatus) {
	op := &request.Operation{
		Name:   OperationTransactionStatus,
		Method: http.MethodPost,
		Path:   EndpointTransactionStatus,
	}

	cfg := request.Config{Endpoint: daraja.endpoint}
	output := &ResponseTransactionStatus{}

	req := request.New(cfg, daraja.hooks, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (daraja Daraja) TransactionStatus(ctx context.Context, payload RequestTransactionStatus) (ResponseTransactionStatus, error) {
	req, out := daraja.TransactionStatusRequest(payload, request.WithRequestHeader("Content-Type", "application/json"))
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseTransactionStatus{}, err
	}

	return *out, nil
}

func (daraja Daraja) BalanceRequest(input RequestBalance, opts ...request.Option) (*request.Request, *ResponseBalance) {
	op := &request.Operation{
		Name:   OperationBalance,
		Method: http.MethodPost,
		Path:   EndpointAccountBalance,
	}

	cfg := request.Config{Endpoint: daraja.endpoint}
	output := &ResponseBalance{}

	req := request.New(cfg, daraja.hooks, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (daraja Daraja) Balance(ctx context.Context, payload RequestBalance) (ResponseBalance, error) {
	req, out := daraja.BalanceRequest(payload, request.WithRequestHeader("Content-Type", "application/json"))
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseBalance{}, err
	}

	return *out, nil
}

func (daraja Daraja) QueryOrgInfoRequest(input RequestOrgInfoQuery, opts ...request.Option) (*request.Request, *ResponseOrgInfoQuery) {
	op := &request.Operation{
		Name:   OperationQueryOrgInfo,
		Method: http.MethodPost,
		Path:   EndpointQueryOrgInfo,
	}

	cfg := request.Config{Endpoint: daraja.endpoint}
	output := &ResponseOrgInfoQuery{}

	req := request.New(cfg, daraja.hooks, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (daraja Daraja) QueryOrgInfo(ctx context.Context, payload RequestOrgInfoQuery) (ResponseOrgInfoQuery, error) {
	req, out := daraja.QueryOrgInfoRequest(payload, request.WithRequestHeader("Content-Type", "application/json"))
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseOrgInfoQuery{}, err
	}

	return *out, nil
}
