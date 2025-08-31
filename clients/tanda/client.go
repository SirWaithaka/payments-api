package tanda

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/request"
)

func DefaultHooks() request.Hooks {
	// create default hooks
	hooks := corehooks.DefaultHooks()

	// create client with default timeout of 5 seconds
	client := &http.Client{Timeout: 5 * time.Second}
	hooks.Build.PushFront(request.WithHTTPClient(client))
	hooks.Build.PushBackHook(corehooks.EncodeRequestBody)
	hooks.Unmarshal.PushBackHook(ResponseDecoder)
	return hooks

}

type Config struct {
	Endpoint string
	Hooks    request.Hooks
	LogLevel request.LogLevel
}

func New(cfg Config) Client {
	if cfg.Hooks.IsEmpty() {
		cfg.Hooks = DefaultHooks()
	}

	// add log level to request config
	cfg.Hooks.Build.PushFront(request.WithLogLevel(cfg.LogLevel))

	return Client{endpoint: cfg.Endpoint, Hooks: cfg.Hooks}
}

type Client struct {
	endpoint string
	Hooks    request.Hooks
}

func (client Client) AuthenticationRequest(clientID, secret string) (*request.Request, *ResponseAuthentication) {
	op := &request.Operation{
		Name:   OperationAuthenticate,
		Method: http.MethodPost,
		Path:   EndpointAuthentication,
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", secret)

	// create a client with a 30-second timeout
	cl := &http.Client{Timeout: time.Second * 30}
	cfg := request.Config{HTTPClient: cl, Endpoint: client.endpoint}

	// default hooks
	hooks := corehooks.DefaultHooks()
	hooks.Build.PushFront(request.WithRequestHeader("Content-Type", "application/x-www-form-urlencoded"))
	hooks.Send.PushFrontHook(corehooks.LogHTTPRequest)
	hooks.Unmarshal.PushBackHook(ResponseDecoder)

	input := strings.NewReader(data.Encode())
	output := &ResponseAuthentication{}
	req := request.New(cfg, hooks, nil, op, input, output)

	return req, output
}

func (client Client) PaymentRequest(orgID string, payload RequestPayment, opts ...request.Option) (*request.Request, *ResponsePayment) {
	op := &request.Operation{
		Name:   OperationPayment,
		Method: http.MethodPost,
		Path:   fmt.Sprintf(EndpointPayments, orgID),
	}

	cfg := request.Config{Endpoint: client.endpoint}

	output := &ResponsePayment{}
	req := request.New(cfg, client.Hooks, nil, op, payload, output)
	req.ApplyOptions(opts...)

	return req, output

}

func (client Client) TransactionStatusRequest(orgID, trackingID, shortCode string, opts ...request.Option) (*request.Request, *ResponseTransactionStatus) {
	op := &request.Operation{
		Name:   OperationTransactionStatus,
		Method: http.MethodGet,
		Path:   fmt.Sprintf(EndpointTransactionStatus, orgID, trackingID),
	}

	// add short code as a query param to the request path
	uParams := url.Values{}
	uParams.Set("shortCode", shortCode)
	op.Path = op.Path + "?" + uParams.Encode()

	cfg := request.Config{Endpoint: client.endpoint}

	output := &ResponseTransactionStatus{}
	req := request.New(cfg, client.Hooks, nil, op, nil, output)
	req.ApplyOptions(opts...)

	return req, output
}
