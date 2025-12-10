package sdk

import (
	"context"
	"net/http"

	"github.com/SirWaithaka/gorequest"
	"github.com/SirWaithaka/gorequest/corehooks"
)

func defaultHooks() gorequest.Hooks {
	hooks := corehooks.Default()

	hooks.Build.PushBackHook(corehooks.EncodeRequestBody)
	hooks.Send.PushBackHook(corehooks.ResponseStatusCode)
	return hooks
}

type Config struct {
	Endpoint string
	Hooks    gorequest.Hooks
	LogLevel gorequest.LogLevel
}

type Client struct {
	endpoint string
	Hooks    gorequest.Hooks
}

func New(cfg Config) Client {
	if cfg.Hooks.IsEmpty() {
		cfg.Hooks = defaultHooks()
	}
	// add log level to request config
	cfg.Hooks.Build.PushFront(gorequest.WithLogLevel(cfg.LogLevel))
	cfg.Hooks.Build.PushFront(gorequest.WithRequestHeader("content-type", "application/json"))

	return Client{endpoint: cfg.Endpoint, Hooks: cfg.Hooks}
}

func (client Client) AddShortCodeRequest(input RequestAddShortCode, opts ...gorequest.Option) *gorequest.Request {
	op := gorequest.Operation{
		Name:   OperationAddShortCode,
		Method: http.MethodPost,
		Path:   EndpointAddShortCode,
	}

	cfg := gorequest.Config{Endpoint: client.endpoint}
	req := gorequest.New(cfg, op, client.Hooks, nil, input, nil)
	req.ApplyOptions(opts...)

	return req
}

func (client Client) AddShortCode(ctx context.Context, input RequestAddShortCode) error {
	req := client.AddShortCodeRequest(input)
	req.WithContext(ctx)

	return req.Send()
}
