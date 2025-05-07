package corehooks

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/SirWaithaka/payments-api/request"
)

var schemeRE = regexp.MustCompile("^([^:]+)://")

func DefaultHooks() request.Hooks {
	var hooks request.Hooks

	hooks.Build.PushBackHook(ResolveEndpoint)
	hooks.Send.PushBackHook(SendHook)

	return hooks
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

var reStatusCode = regexp.MustCompile(`^(\d{3})`)

var SendHook = request.Hook{Fn: func(r *request.Request) {
	sender := sendFollowRedirects
	if r.Config.DisableFollowRedirects {
		sender = sendWithoutFollowRedirects
	}

	if request.NoBody == r.Request.Body {
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
				Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
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
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		}
	}
	// Catch all request errors, and let the default retrier determine
	// if the error is retryable.
	r.Error = err

	// Override the error with a context canceled error if that was canceled.
	ctx := r.Context()
	select {
	case <-ctx.Done():
		// set r.Error to context error and set request retry to false
		r.Error = ctx.Err()
		r.RetryConfig.SetRetryable(false)
	default:
	}
}
