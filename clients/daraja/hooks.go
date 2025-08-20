package daraja

import (
	"fmt"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/SirWaithaka/payments-api/request"
)

// HTTPClient creates an instance of http.Client configured
// for daraja service.
func HTTPClient(client *http.Client) request.Hook {
	return request.Hook{
		Name: "daraja.HTTPClient",
		Fn: func(r *request.Request) {
			if client == nil {
				client = &http.Client{Timeout: 30 * time.Second}
			}

			r.Config.HTTPClient = client
		}}
}

type errResponse ErrorResponse

func (r errResponse) Error() string {
	return fmt.Sprintf("<%s> %s", r.ErrorCode, r.ErrorMessage)
}

// ResponseDecoder parse the http.Response body into the property
// request.Request.Data, if the status code is successful
// Otherwise for failed requests, it will parse the error response
// into the property request.Request.Error
var ResponseDecoder = request.Hook{
	Name: "daraja.ResponseDecoder",
	Fn: func(r *request.Request) {
		// response formats for non-200 status codes follow the same format
		if r.Response.StatusCode != http.StatusOK {
			response := &errResponse{}
			if err := jsoniter.NewDecoder(r.Response.Body).Decode(response); err != nil {
				r.Error = err
				return
			}
			r.Error = response
			return
		}

		if err := jsoniter.NewDecoder(r.Response.Body).Decode(r.Data); err != nil {
			r.Error = err
		}
	},
}

func Authenticate(reqFn AuthenticationRequestFunc) request.Hook {

	cache := NewCache[string]()

	return request.Hook{
		Name: "daraja.Authenticate",
		Fn: func(r *request.Request) {
			// make request to authenticate if token cache is empty
			if cache.Get() == "" {
				req, out := reqFn()
				req.WithContext(r.Context())
				req.Config.Logger = r.Config.Logger
				// make request
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
