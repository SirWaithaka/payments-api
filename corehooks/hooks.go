package corehooks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/SirWaithaka/payments-api/request"
)

var schemeRE = regexp.MustCompile("^([^:]+)://")

func DefaultHooks() request.Hooks {
	var hooks request.Hooks

	hooks.Build.PushBackHook(ResolveEndpoint)
	hooks.Send.PushBackHook(SendHook)

	// new instance of retry hook
	retryHook := NewRetryer()
	hooks.Retry.PushBackHook(retryHook.Retry())
	hooks.Complete.PushBackHook(retryHook.Close())

	return hooks
}

// SetBasicAuth modifies the http.Request headers and adds basic auth credentials
func SetBasicAuth(username, password string) request.Hook {
	return request.Hook{Fn: func(r *request.Request) {
		r.Request.SetBasicAuth(username, password)
	}}
}

// AddScheme adds the HTTP or HTTPS schemes to an endpoint URL if there is no
// scheme. If disableSSL is true, HTTP will set HTTP instead of the default HTTPS.
//
// If disableSSL is set, it will only set the URL's scheme if the URL does not
// contain a scheme.
func AddScheme(endpoint string, disableSSL bool) string {
	if !schemeRE.MatchString(endpoint) {
		scheme := "https"
		if disableSSL {
			scheme = "http"
		}
		endpoint = fmt.Sprintf("%s://%s", scheme, endpoint)
	}

	return endpoint
}

var ResolveEndpoint = request.Hook{Fn: func(r *request.Request) {
	r.Config.Endpoint = AddScheme(r.Config.Endpoint, r.Config.DisableSSL)
}}

// EncodeRequestBody converts the value in r.Params into an io reader and adds it
// to the http.Request instance
var EncodeRequestBody = request.Hook{Fn: func(r *request.Request) {
	if r.Params == nil {
		return
	}

	buf := new(bytes.Buffer)
	if err := jsoniter.NewEncoder(buf).Encode(r.Params); err != nil {
		r.Error = err
		return
	}

	// add as body to request
	r.Request.Body = io.NopCloser(buf)

}}

var reStatusCode = regexp.MustCompile(`^(\d{3})`)

var SendHook = request.Hook{Fn: func(r *request.Request) {
	sender := sendFollowRedirects
	if r.Config.DisableFollowRedirects {
		sender = sendWithoutFollowRedirects
	}

	if r.Request.Body == request.NoBody {
		// Strip off the request body if the NoBody reader was used as a
		// placeholder for a request body. This prevents the SDK from
		// making requests with a request body when it would be invalid
		// to do so.
		//
		// Use a shallow copy of the http.Request to ensure the race condition
		// of transport on Body will not trigger
		reqOrig, reqCopy := r.Request, *r.Request
		reqCopy.Body = nil
		r.Request = &reqCopy
		defer func() {
			r.Request = reqOrig
		}()
	}

	var err error
	r.Response, err = sender(r)
	if err != nil {
		handleSendError(r, err)
	}
}}

func sendFollowRedirects(r *request.Request) (*http.Response, error) {
	return r.Config.HTTPClient.Do(r.Request)
}

func sendWithoutFollowRedirects(r *request.Request) (*http.Response, error) {
	transport := r.Config.HTTPClient.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	return transport.RoundTrip(r.Request)
}

func handleSendError(r *request.Request, err error) {
	// Prevent leaking if an HTTPResponse was returned. Clean up
	// the body.
	if r.Response != nil {
		r.Response.Body.Close()
	}

	// Capture the case where url.Error is returned for error processing
	// response. e.g., 301 without location header comes back as string
	// error and r.HTTPResponse is nil. Other URL redirect errors will
	// come back in a similar method.
	if e, ok := err.(*url.Error); ok && e.Err != nil {
		if s := reStatusCode.FindStringSubmatch(e.Err.Error()); s != nil {
			code, _ := strconv.ParseInt(s[1], 10, 64)
			r.Response = &http.Response{
				StatusCode: int(code),
				Status:     http.StatusText(int(code)),
				Body:       io.NopCloser(bytes.NewReader([]byte{})),
			}
			return
		}
	}
	if r.Response == nil {
		// Add a dummy request response object to ensure the http response
		// value is consistent.
		r.Response = &http.Response{
			StatusCode: int(0),
			Status:     http.StatusText(int(0)),
			Body:       io.NopCloser(bytes.NewReader([]byte{})),
		}
	}
	// Catch all request errors and let the default retrier determine
	// if the error is retryable.
	r.Error = err

	// Override the error with a context-canceled error if that was canceled.
	ctx := r.Context()
	select {
	case <-ctx.Done():
		// set r.Error to context error and set request retry to false
		r.Error = ctx.Err()
	default:
	}
}

type timer struct {
	timer *time.Timer
}

func (t *timer) C() <-chan time.Time {
	return t.timer.C
}

func (t *timer) Start(dur time.Duration) {
	if t.timer == nil {
		t.timer = time.NewTimer(dur)
	} else {
		t.timer.Reset(dur)
	}
}

// Stop is used to free resources when timer is no longer used
func (t *timer) Stop() {
	if t.timer != nil {
		t.timer.Stop()
	}
}

func NewRetryer() RetryHook {
	return RetryHook{timer: &timer{}}
}

type RetryHook struct {
	timer *timer
}

func (r RetryHook) nextDelay(cfg request.RetryConfig) time.Duration {
	// using the multiplier, calculate the next delay
	next := float64(cfg.CurrentDelay) * cfg.Multiplier
	// if the next calculated delay is greater than max delay, return max delay
	if next > float64(cfg.MaxDelay) {
		return cfg.MaxDelay
	}
	return time.Duration(next)
}

func (r *RetryHook) Retry() request.Hook {
	return request.Hook{Fn: func(req *request.Request) {
		// increment retry count
		req.RetryConfig.RetryCount += 1

		ctx := req.Context()
		// Stop retrying if context is canceled
		if err := context.Cause(ctx); err != nil {
			req.Error = err
			return
		}

		// start the timer and wait
		r.timer.Start(req.Delay(req))
		// wait for timer to complete or context Done signal
		select {
		case <-r.timer.C():
		case <-ctx.Done():
			req.Error = context.Cause(ctx)
			return
		}

		// get the next delay duration
		next := r.nextDelay(req.RetryConfig)
		req.RetryConfig.CurrentDelay = next

	}}
}

// Close stops the timer. Call close as a complete hook to ensure the timer is stopped.
func (r *RetryHook) Close() request.Hook {
	return request.Hook{Fn: func(_ *request.Request) {
		r.timer.Stop()
	}}
}

// LogHTTPRequest is a hook to log the HTTP request sent to a service. If log level
// matches request.LogDebugWithHTTPBody, the request body will be included.
var LogHTTPRequest = request.Hook{Fn: logRequest}

func logRequest(r *request.Request) {
	if !r.Config.LogLevel.AtLeast(request.LogDebug) || r.Config.Logger == nil {
		return
	}

	logBody := r.Config.LogLevel.Equals(request.LogDebugWithHTTPBody)
	b, err := httputil.DumpRequest(r.Request, logBody)
	if err != nil {
		r.Config.Logger.Log(fmt.Sprintf("DEBUG: %s failed, error %v",
			r.Operation.Name, err))
		return
	}

	r.Config.Logger.Log(fmt.Sprintf("DEBUG: %s, %s",
		r.Operation.Name, string(b)))

}
