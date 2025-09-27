package hooks

import (
	"time"

	"github.com/SirWaithaka/gorequest"

	pkgerrors "github.com/SirWaithaka/payments-api/pkg/errors"
	"github.com/SirWaithaka/payments-api/pkg/types"
	"github.com/SirWaithaka/payments-api/src/domains/requests"
)

func NewRequestRecorder(repository requests.Repository) RequestRecorder {
	return RequestRecorder{repository: repository}
}

type RequestRecorder struct {
	repository requests.Repository
}

// RecordRequest hooks saves all outgoing requests before the http request is sent to the
// external api.
func (recorder RequestRecorder) RecordRequest(paymentID, requestID string) gorequest.Hook {
	return gorequest.Hook{Name: "RequestRecorder.RecordRequest", Fn: func(r *gorequest.Request) {

		req := requests.Request{
			RequestID: requestID,
			PaymentID: paymentID,
			Partner:   r.Config.ServiceName,
			Status:    requests.StatusReceived,
		}

		// save request
		err := recorder.repository.Add(r.Context(), req)
		if err != nil {
			r.Error = err
			return
		}

	}}
}

// UpdateRequestResponse updates a request record after the http request is made and a response
// is/is not received.
func (recorder RequestRecorder) UpdateRequestResponse(requestID string) gorequest.Hook {
	return gorequest.Hook{Name: "RequestRecorder.UpdateRequestResponse", Fn: func(r *gorequest.Request) {
		opts := requests.OptionsUpdateRequest{}

		defer func() {
			err := recorder.repository.UpdateRequest(r.Context(), requestID, opts)
			if err != nil {
				r.Error = err
				return
			}
		}()

		resMap := make(map[string]any)

		// check if the request had an error
		if r.Error != nil {
			resMap["error"] = r.Error.Error()

			// check if it's a 4xx status code
			if r.Response.StatusCode >= 400 && r.Response.StatusCode < 500 {
				s := requests.StatusFailed
				opts.Status = &s
			} else // check if it is a timeout error
			if etimeout, ok := r.Error.(pkgerrors.Timeout); ok && etimeout.Timeout() {
				s := requests.StatusTimeout
				opts.Status = &s
			} else // check if it is a temporary error
			if etemp, ok := r.Error.(pkgerrors.Temporary); ok && etemp.Temporary() {
				s := requests.StatusError
				opts.Status = &s
			} else {
				s := requests.StatusError
				opts.Status = &s
			}

			opts.Response = resMap
			return
		}

		resMap["response"] = r.Data

		// check if it implements interface
		if in, ok := r.Data.(interface{ ExternalID() string }); ok {
			id := in.ExternalID()
			opts.ExternalID = &id
		}

		s := requests.StatusSucceeded
		opts.Status = &s
		opts.Latency = types.Pointer(time.Since(r.AttemptTime))
		opts.Response = resMap
	}}
}
