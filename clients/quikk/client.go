package quikk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/request"
)

// given a data and secret, signer generates a base64 encoded hmac signature
func signer(date, secret []byte) string {
	data := []byte(fmt.Sprintf("date: %s", date))
	h := hmac.New(sha256.New, secret)
	h.Write(data)
	b64 := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return url.QueryEscape(b64)
}

func DefaultHooks() request.Hooks {
	hooks := corehooks.DefaultHooks()

	hooks.Build.PushBackHook(corehooks.EncodeRequestBody)
	hooks.Unmarshal.PushBackHook(ResponseDecoder)
	return hooks
}

type Config struct {
	Endpoint string
	Hooks    request.Hooks
	LogLevel request.LogLevel
}

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
	cfg.Hooks.Build.PushFront(request.WithRequestHeader("accept", "application/vnd.api+json"))
	cfg.Hooks.Build.PushFront(request.WithRequestHeader("content-type", "application/vnd.api+json"))

	return Client{endpoint: cfg.Endpoint, Hooks: cfg.Hooks}
}

func (client Client) VerifyAuth(opts ...request.Option) (*request.Request, []byte) {
	op := &request.Operation{
		Name:   OperationAuthCheck,
		Method: http.MethodGet,
		Path:   EndpointAuthCheck,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	var output []byte
	req := request.New(cfg, client.Hooks, nil, op, nil, &output)
	req.ApplyOptions(opts...)

	return req, output
}

func (client Client) ChargeRequest(input RequestCharge, ref string, opts ...request.Option) (*request.Request, *ResponseDefault) {
	op := &request.Operation{
		Name:   OperationCharge,
		Method: http.MethodPost,
		Path:   EndpointCharge,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// append to request options
	//opts = append(opts, request.WithRequestHeader("Content-Type", "application/json"))

	// build actual payload
	payload := RequestDefault[RequestCharge]{
		Data: Data[RequestCharge]{
			ID:         ref,
			Type:       "charge",
			Attributes: input,
		},
	}

	output := ResponseDefault{}
	req := request.New(cfg, client.Hooks, nil, op, payload, &output)
	req.ApplyOptions(opts...)

	return req, &output
}

func (client Client) PayoutRequest(input RequestPayout, ref string, opts ...request.Option) (*request.Request, *ResponseDefault) {
	op := &request.Operation{
		Name:   OperationPayout,
		Method: http.MethodPost,
		Path:   EndpointPayout,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// build actual payload
	payload := RequestDefault[RequestPayout]{
		Data: Data[RequestPayout]{
			ID:         ref,
			Type:       "payout",
			Attributes: input,
		},
	}

	output := ResponseDefault{}
	req := request.New(cfg, client.Hooks, nil, op, payload, &output)
	req.ApplyOptions(opts...)

	return req, &output
}

func (client Client) TransferRequest(input RequestTransfer, ref string, opts ...request.Option) (*request.Request, *ResponseDefault) {
	op := &request.Operation{
		Name:   OperationTransfer,
		Method: http.MethodPost,
		Path:   EndpointTransfer,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// build actual payload
	payload := RequestDefault[RequestTransfer]{
		Data: Data[RequestTransfer]{
			ID:         ref,
			Type:       "transfer",
			Attributes: input,
		},
	}

	output := ResponseDefault{}
	req := request.New(cfg, client.Hooks, nil, op, payload, &output)
	req.ApplyOptions(opts...)

	return req, &output
}

func (client Client) BalanceRequest(input RequestAccountBalance, ref string, opts ...request.Option) (*request.Request, *ResponseDefault) {
	op := &request.Operation{
		Name:   OperationBalance,
		Method: http.MethodPost,
		Path:   EndpointBalance,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// build actual payload
	payload := RequestDefault[RequestAccountBalance]{
		Data: Data[RequestAccountBalance]{
			ID:         ref,
			Type:       "search",
			Attributes: input,
		},
	}

	output := ResponseDefault{}
	req := request.New(cfg, client.Hooks, nil, op, payload, &output)
	req.ApplyOptions(opts...)

	return req, &output
}

func (client Client) TransactionSearchRequest(input RequestTransactionStatus, ref string, opts ...request.Option) (*request.Request, *ResponseDefault) {
	op := &request.Operation{
		Name:   OperationTransactionSearch,
		Method: http.MethodPost,
		Path:   EndpointTransactionSearch,
	}

	cfg := request.Config{Endpoint: client.endpoint}

	// build actual payload
	payload := RequestDefault[RequestTransactionStatus]{
		Data: Data[RequestTransactionStatus]{
			ID:         ref,
			Type:       "search",
			Attributes: input,
		},
	}

	output := ResponseDefault{}
	req := request.New(cfg, client.Hooks, nil, op, payload, &output)
	req.ApplyOptions(opts...)

	return req, &output
}

func (client Client) TransactionSearch(ctx context.Context, input RequestTransactionStatus, ref string) (ResponseDefault, error) {
	req, out := client.TransactionSearchRequest(input, ref)
	req.WithContext(ctx)

	if err := req.Send(); err != nil {
		return ResponseDefault{}, err
	}

	return *out, nil
}
