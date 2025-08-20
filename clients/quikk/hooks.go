package quikk

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/SirWaithaka/payments-api/request"
)

// Sign is a build hook that generates a signature for the request
func Sign(key, secret string) request.Hook {
	return request.Hook{Name: "quikk.SignRequest", Fn: func(r *request.Request) {
		// get the current time and sign it
		now := time.Now().UTC().Format(time.RFC1123)
		signature := signer([]byte(now), []byte(secret))

		// build the authorization header
		authorization := fmt.Sprintf(`keyId=%q,algorithm="hmac-sha256",headers="date",signature=%q`, key, signature)

		r.Request.Header.Set("Date", now)
		r.Request.Header.Set("Authorization", authorization)

	}}
}

type errorResponse ErrorResponse

func (r errorResponse) Error() string {
	return fmt.Sprintf("<%s> %s", r.Errors[0].Status, r.Errors[0].Title)
}

// ResponseDecoder decodes the response body into the Data field of request.Request if the status code
// is 200. Otherwise, it decodes into the ErrorResponse model
var ResponseDecoder = request.Hook{
	Name: "quikk.ResponseDecoder",
	Fn: func(r *request.Request) {
		// response formats for non-200 status codes follow the same format
		if r.Response.StatusCode != http.StatusOK {
			statusError := fmt.Errorf("status code: %d", r.Response.StatusCode)

			response := &errorResponse{}
			if err := jsoniter.NewDecoder(r.Response.Body).Decode(response); err != nil {
				r.Error = errors.Join(statusError, errors.New("failed to decode response"), err)
				return
			}
			r.Error = errors.Join(statusError, response)
			return
		}

		// decode into Data field of r if status is 200
		if err := jsoniter.NewDecoder(r.Response.Body).Decode(r.Data); err != nil {
			r.Error = err
		}
	},
}
