package daraja

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/request"
)

type AuthenticationRequestFunc func() (*request.Request, *ResponseAuthorization)

type Config struct {
	Endpoint string
	Hooks    request.Hooks
	LogLevel request.LogLevel
}

func DefaultHooks() request.Hooks {
	// create default hooks
	hooks := corehooks.DefaultHooks()

	// create client with default timeout of 5 seconds
	client := &http.Client{Timeout: 5 * time.Second}
	hooks.Build.PushBackHook(SetEndpoint(SandboxUrl))
	hooks.Build.PushBackHook(HTTPClient(client))
	hooks.Build.PushBackHook(corehooks.EncodeRequestBody)
	hooks.Unmarshal.PushBackHook(ResponseDecoder)
	return hooks

}

func PasswordEncode(shortcode, passphrase, timestamp string) string {
	return base64.StdEncoding.EncodeToString([]byte(shortcode + passphrase + timestamp))
}

// Client provides the API operation methods for making requests
// to MPESA daraja service.
type Client struct {
	endpoint string
	Hooks    request.Hooks
}

func New(cfg Config) Client {
	if cfg.Hooks.IsEmpty() {
		cfg.Hooks = DefaultHooks()
	}

	// add log level to request config
	cfg.Hooks.Build.PushFront(request.WithLogLevel(cfg.LogLevel))

	return Client{endpoint: cfg.Endpoint, Hooks: cfg.Hooks}
}

func (client Client) AuthenticationRequest(key, secret string) AuthenticationRequestFunc {
	return func() (*request.Request, *ResponseAuthorization) {
		op := &request.Operation{
			Name:   "Authenticate",
			Method: http.MethodGet,
			Path:   EndpointAuthentication + "?grant_type=client_credentials",
		}

		// create a client with a 40-second timeout
		cl := &http.Client{Timeout: time.Second * 40}
		cfg := request.Config{HTTPClient: cl, Endpoint: client.endpoint}

		// default hooks
		hooks := corehooks.DefaultHooks()
		hooks.Build.PushBackHook(corehooks.SetBasicAuth(key, secret))
		hooks.Send.PushFrontHook(corehooks.LogHTTPRequest)
		hooks.Unmarshal.PushBackHook(ResponseDecoder)

		output := &ResponseAuthorization{}
		req := request.New(cfg, hooks, nil, op, nil, output)

		return req, output
	}
}

func (client Client) C2BExpressRequest(input RequestC2BExpress, opts ...request.Option) (*request.Request, *ResponseC2BExpress) {
	op := &request.Operation{
		Name:   OperationC2BExpress,
		Method: http.MethodPost,
		Path:   EndpointC2bExpress,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// append to request options
	opts = append(opts, request.WithRequestHeader("Content-Type", "application/json"))

	output := ResponseC2BExpress{}
	req := request.New(cfg, client.Hooks, nil, op, input, &output)
	req.ApplyOptions(opts...)

	return req, &output
}

func (client Client) C2BExpress(ctx context.Context, payload RequestC2BExpress) (ResponseC2BExpress, error) {
	req, out := client.C2BExpressRequest(payload)
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseC2BExpress{}, err
	}

	return *out, nil
}

func (client Client) ReversalRequest(input RequestReversal, opts ...request.Option) (*request.Request, *ResponseReversal) {
	op := &request.Operation{
		Name:   OperationReversal,
		Method: http.MethodPost,
		Path:   EndpointReversal,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// append to request options
	opts = append(opts, request.WithRequestHeader("Content-Type", "application/json"))

	output := &ResponseReversal{}
	req := request.New(cfg, client.Hooks, nil, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (client Client) Reverse(ctx context.Context, payload RequestReversal) (ResponseReversal, error) {
	req, out := client.ReversalRequest(payload)
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseReversal{}, err
	}

	return *out, nil
}

func (client Client) B2CRequest(input RequestB2C, opts ...request.Option) (*request.Request, *ResponseB2C) {
	op := &request.Operation{
		Name:   OperationB2C,
		Method: http.MethodPost,
		Path:   EndpointB2cPayment,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// append to request options
	opts = append(opts, request.WithRequestHeader("Content-Type", "application/json"))

	output := &ResponseB2C{}
	req := request.New(cfg, client.Hooks, nil, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (client Client) B2C(ctx context.Context, payload RequestB2C) (ResponseB2C, error) {
	req, out := client.B2CRequest(payload)
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseB2C{}, err
	}

	return *out, nil
}

func (client Client) B2BRequest(input RequestB2B, opts ...request.Option) (*request.Request, *ResponseB2B) {
	op := &request.Operation{
		Name:   OperationB2B,
		Method: http.MethodPost,
		Path:   EndpointB2bPayment,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// append to request options
	opts = append(opts, request.WithRequestHeader("Content-Type", "application/json"))

	output := &ResponseB2B{}
	req := request.New(cfg, client.Hooks, nil, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (client Client) B2B(ctx context.Context, payload RequestB2B) (ResponseB2B, error) {
	req, out := client.B2BRequest(payload)
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseB2B{}, err
	}

	return *out, nil
}

func (client Client) TransactionStatusRequest(input RequestTransactionStatus, opts ...request.Option) (*request.Request, *ResponseTransactionStatus) {
	op := &request.Operation{
		Name:   OperationTransactionStatus,
		Method: http.MethodPost,
		Path:   EndpointTransactionStatus,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// append to request options
	opts = append(opts, request.WithRequestHeader("Content-Type", "application/json"))

	output := &ResponseTransactionStatus{}
	req := request.New(cfg, client.Hooks, nil, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (client Client) TransactionStatus(ctx context.Context, payload RequestTransactionStatus) (ResponseTransactionStatus, error) {
	req, out := client.TransactionStatusRequest(payload)
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseTransactionStatus{}, err
	}

	return *out, nil
}

func (client Client) BalanceRequest(input RequestBalance, opts ...request.Option) (*request.Request, *ResponseBalance) {
	op := &request.Operation{
		Name:   OperationBalance,
		Method: http.MethodPost,
		Path:   EndpointAccountBalance,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// append to request options
	opts = append(opts, request.WithRequestHeader("Content-Type", "application/json"))

	output := &ResponseBalance{}
	req := request.New(cfg, client.Hooks, nil, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (client Client) Balance(ctx context.Context, payload RequestBalance) (ResponseBalance, error) {
	req, out := client.BalanceRequest(payload)
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseBalance{}, err
	}

	return *out, nil
}

func (client Client) QueryOrgInfoRequest(input RequestOrgInfoQuery, opts ...request.Option) (*request.Request, *ResponseOrgInfoQuery) {
	op := &request.Operation{
		Name:   OperationQueryOrgInfo,
		Method: http.MethodPost,
		Path:   EndpointQueryOrgInfo,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// append to request options
	opts = append(opts, request.WithRequestHeader("Content-Type", "application/json"))

	output := &ResponseOrgInfoQuery{}
	req := request.New(cfg, client.Hooks, nil, op, input, output)
	req.ApplyOptions(opts...)

	return req, output
}

func (client Client) QueryOrgInfo(ctx context.Context, payload RequestOrgInfoQuery) (ResponseOrgInfoQuery, error) {
	req, out := client.QueryOrgInfoRequest(payload)
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseOrgInfoQuery{}, err
	}

	return *out, nil
}
