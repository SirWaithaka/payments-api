package daraja

import (
	"fmt"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/SirWaithaka/payments-api/request"
)

// SetBaseUrl checks the environment passed and uses the appropriate base url
// otherwise if the baseUrl is not empty, it sets the Request endpoint to it.
//
// Should be used as a build hook
func SetBaseUrl(baseUrl, environment string) request.Hook {
	// set default base url as the sandbox url
	url := sandboxUrl

	if environment == ENV_PRODUCTION {
		url = productionUrl
	}

	// lastly, if base url is not empty, use that instead
	if baseUrl != "" {
		url = baseUrl
	}

	return request.Hook{Fn: func(r *request.Request) {
		r.Config.Endpoint = url
	}}
}

// HTTPClient creates an instance of http.Client configured
// for daraja service.
func HTTPClient(client *http.Client) request.Hook {
	return request.Hook{Fn: func(r *request.Request) {
		if client == nil {
			client = &http.Client{Timeout: 30 * time.Second}
		}

		r.Config.HTTPClient = client
	}}
}

type errResponse struct {
	ErrorResponse
}

func (r errResponse) Error() string {
	return fmt.Sprintf("<%s> %s", r.ErrorCode, r.ErrorMessage)
}

// DecodeResponse parse the http.Response body into the property
// request.Request.Data, if the status code is successful
// Otherwise for
func DecodeResponse() request.Hook {

	return request.Hook{Fn: func(r *request.Request) {

		if r.Response.StatusCode != http.StatusOK {
			response := &errResponse{}
			if err := jsoniter.NewDecoder(r.Response.Body).Decode(response.ErrorResponse); err != nil {
				r.Error = err
				return
			}
			r.Error = response
			return
		}

		if err := jsoniter.NewDecoder(r.Response.Body).Decode(r.Data); err != nil {
			r.Error = err
		}
	}}
}

func Authenticate(endpoint, key, secret string) request.Hook {

	cache := NewCache[string]()

	return request.Hook{Fn: func(r *request.Request) {
		// make request to authenticate if token cache is empty
		if cache.Get() == "" {
			req, out := AuthenticationRequest(endpoint, key, secret)
			req.WithContext(r.Context())

			if err := req.Send(); err != nil {
				r.Error = err
				return
			}

			// if authentication request was successful, save token to cache
			cache.Set(out.AccessToken, time.Now().Add(12*time.Hour))
		}

		// add access token to request authorization header
		r.Request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cache.Get()))

	}}
}

func RecordRequest() request.Hook {
	return request.Hook{Fn: func(r *request.Request) {

	}}
}
