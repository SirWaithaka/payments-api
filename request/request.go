package request

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

type (
	Operation struct {
		Name   string
		Method string
		Path   string
	}

	Request struct {
		Config Config

		//BaseUrl string

		// response body
		Body any
		// request payload
		Data any

		Hooks Hooks

		Error error
		ctx   context.Context

		operation *Operation
		Request   *http.Request
		response  *http.Response

		// a boolean to indicate with request is build
		built bool
	}
)

// New returns a new Request pointer for the api operation and parameters.
//
// Data is any value for the request payload.
func New(hooks Hooks, operation *Operation, data any) *Request {

	method := operation.Method
	if method == "" {
		method = http.MethodPost
	}

	httpReq, _ := http.NewRequest(method, "", nil)

	return &Request{
		Request:   httpReq,
		operation: operation,
		Hooks:     hooks.Copy(),
	}
}

// Build will build the request object to be sent. Build will also
// validate all the request's parameters.
//
// If any Validate or Build errors occur the build will stop and the error
// which occurred will be returned
func (r *Request) Build() error {
	if r.built {
		return r.Error
	}

	// run validate hooks
	r.Hooks.Validate.Run(r)
	if r.Error != nil {
		return r.Error
	}
	// run build hooks
	r.Hooks.Build.Run(r)
	if r.Error != nil {
		return r.Error
	}
	r.built = true

	return nil
}

func (r *Request) Send() error {
	defer func() {
		// Ensure a non-nil HTTPResponse parameter is set to ensure hooks
		// checking for HTTPResponse values, don't fail.
		if r.response == nil {
			r.response = &http.Response{
				Header: http.Header{},
				Body:   io.NopCloser(&bytes.Buffer{}),
			}
		}
		// Regardless of success or failure of the request trigger the Complete
		// request hooks.
		r.Hooks.Complete.Run(r)
	}()

	// build the request
	r.Build()
	if r.Error != nil {
		return r.Error
	}

	// run hooks that process sending the request
	r.Hooks.Send.Run(r)
	if r.Error != nil {
		// todo: log error
		return r.Error
	}

	// run any hooks that unmarshal/validate the response
	r.Hooks.Unmarshal.Run(r)
	if r.Error != nil {
		return r.Error
	}

	return nil
}

func (r *Request) SetContext(ctx context.Context) {
	if ctx == nil {
		return
	}

	r.ctx = ctx
	r.Request.WithContext(ctx)
}
