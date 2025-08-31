package request

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// NoBody is an http.NoBody reader instructing the Go HTTP client to not include
	// a body in the HTTP request.
	NoBody = http.NoBody
)

type (
	Operation struct {
		Name   string
		Method string
		Path   string
	}

	Request struct {
		Config Config

		// request payload
		Params any

		// response body
		Body any
		Data any

		Hooks Hooks

		Error error
		ctx   context.Context

		Retryer
		RetryConfig RetryConfig

		Operation *Operation
		Request   *http.Request
		Response  *http.Response

		AttemptTime time.Time

		// a boolean to indicate with request is build
		built bool
	}

	// An Option is a functional option that can augment or modify a request when
	// using a WithContext API operation method.
	Option func(*Request)
)

// WithRequestHeader builds a request Option which will add an http header to the
// request.
func WithRequestHeader(key, val string) Option {
	return func(r *Request) {
		r.Request.Header.Add(key, val)
	}
}

// WithLogLevel sets log level
func WithLogLevel(l LogLevel) Option {
	return func(r *Request) {
		r.Config.LogLevel = l
	}
}

// WithLogger sets the logger func used with the request
func WithLogger(logger Logger) Option {
	return func(r *Request) {
		r.Config.Logger = logger
	}
}

// WithServiceName sets the service name in config
func WithServiceName(name string) Option {
	return func(r *Request) {
		r.Config.ServiceName = name
	}
}

// WithRequestID sets the request id in config
func WithRequestID(id string) Option {
	return func(r *Request) {
		r.Config.RequestID = id
	}
}

// WithHTTPClient sets the http client used to send the request
func WithHTTPClient(client *http.Client) Option {
	return func(r *Request) {
		r.Config.HTTPClient = client
	}
}

// ApplyOptions will apply each option to the request calling them in the order
// they were provided.
func (r *Request) ApplyOptions(opts ...Option) {
	for _, opt := range opts {
		opt(r)
	}
}

// New returns a new Request pointer for the api operation and parameters.
//
// Params is any value for the request payload.
//
// Data is for the response payload
func New(cfg Config, hooks Hooks, retryer Retryer, operation *Operation, params, data any) *Request {
	// set a default http client if not provided
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}

	if retryer == nil {
		retryer = noOpRetryer{}
	}

	if operation == nil {
		operation = &Operation{Method: http.MethodPost}
	}

	method := operation.Method
	if method == "" {
		method = http.MethodPost
	}

	httpReq, _ := http.NewRequest(method, "", nil)

	var err error
	httpReq.URL, err = url.Parse(cfg.Endpoint)
	if err != nil {
		httpReq.URL = &url.URL{}
		err = errors.Join(errors.New("invalid endpoint url"), err)
	}

	// append path to request url
	if len(operation.Path) != 0 {
		opHTTPPath := operation.Path
		var opQueryString string
		if idx := strings.Index(opHTTPPath, "?"); idx >= 0 {
			opQueryString = opHTTPPath[idx+1:]
			opHTTPPath = opHTTPPath[:idx]
		}

		if strings.HasSuffix(httpReq.URL.Path, "/") && strings.HasPrefix(opHTTPPath, "/") {
			opHTTPPath = opHTTPPath[1:]
		}
		httpReq.URL.Path += opHTTPPath
		httpReq.URL.RawQuery = opQueryString
	}

	return &Request{
		Config:      cfg,
		Request:     httpReq,
		Operation:   operation,
		Hooks:       hooks.Copy(),
		Params:      params,
		Data:        data,
		Error:       err,
		Retryer:     retryer,
		RetryConfig: RetryConfig{}, // noOp retry config
	}
}

func debugLogReqError(r *Request, stage string, err error) {
	if !r.Config.LogLevel.AtLeast(LogError) {
		return
	}

	r.Config.Logger.Log(fmt.Sprintf("DEBUG: %s %s failed, error %v",
		stage, r.Operation.Name, err))
}

// Build will build the request object to be sent. Build will also
// validate all the request's parameters.
//
// If any Validate or Build errors occur, the build will stop and the error
// which occurred will be returned
func (r *Request) Build() error {
	if r.built {
		return r.Error
	}

	// run validate hooks
	r.Hooks.Validate.Run(r)
	if r.Error != nil {
		debugLogReqError(r, "Validate", r.Error)
		return r.Error
	}
	// run build hooks
	r.Hooks.Build.Run(r)
	if r.Error != nil {
		debugLogReqError(r, "Build", r.Error)
		return r.Error
	}
	r.built = true

	return nil
}

func (r *Request) Send() error {
	defer func() {
		// Ensure a non-nil HTTPResponse parameter is set to ensure hooks
		// checking for HTTPResponse values, don't fail.
		if r.Response == nil {
			r.Response = &http.Response{
				Header: http.Header{},
				Body:   io.NopCloser(&bytes.Buffer{}),
			}
		}
		// Regardless of success or failure of the request, trigger the Complete
		// request hooks.
		r.Hooks.Complete.Run(r)
	}()

	// build the request
	err := r.Build()
	if err != nil {
		return r.Error
	}

	r.AttemptTime = time.Now()
	for {
		r.Error = nil

		if err = r.sendRequest(); err == nil {
			// return immediately to break loop if we encounter no error
			return nil
		}

		// if an error occurred, return if Request is not retryable
		if r.Error != nil && !r.Retryer.Retryable(r) {
			return r.Error
		}

		// run hooks to retry the request
		r.Hooks.Retry.Run(r)
		if r.Error != nil {
			return r.Error
		}

		if err := r.prepareRetry(); err != nil {
			r.Error = err
			return err
		}
	}
}

func (r *Request) prepareRetry() error {
	if r.Config.LogLevel.Equals(LogDebugWithRequestRetries) && r.Config.Logger != nil {
		r.Config.Logger.Log(fmt.Sprintf("DEBUG: Retrying Request %s, attempt %d",
			r.Operation.Name, r.RetryConfig.RetryCount))
	}

	// The previous http.Request will have a reference to the r.Body
	// and the HTTP Client's Transport may still be reading from
	// the request's body even though the Client's Do returned.
	r.Request = copyHTTPRequest(r.Request, nil)

	// Closing response body to ensure that no response body is leaked
	// between retry attempts.
	if r.Response != nil && r.Response.Body != nil {
		r.Response.Body.Close()
	}

	return nil
}

func (r *Request) sendRequest() error {
	// run hooks that process sending the request
	r.Hooks.Send.Run(r)
	if r.Error != nil {
		debugLogReqError(r, "Send", r.Error)
		return r.Error
	}

	// run any hooks that unmarshal/validate the response
	r.Hooks.Unmarshal.Run(r)
	if r.Error != nil {
		debugLogReqError(r, "Unmarshal", r.Error)
		return r.Error
	}

	return nil
}

// Context will always return a non-nil context. If the Request does not have a
// context, context.Background will be returned.
func (r *Request) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

func (r *Request) WithContext(ctx context.Context) {
	if ctx == nil {
		return
	}

	r.ctx = ctx
	r.Request = r.Request.WithContext(ctx)
}

func (r *Request) WithRetryConfig(cfg RetryConfig) {
	r.RetryConfig = cfg
}
