package services

import (
	"github.com/SirWaithaka/payments-api/internal/domains/payments"
	pkgerrors "github.com/SirWaithaka/payments-api/internal/pkg/errors"
	"github.com/SirWaithaka/payments-api/request"
)

func NewRequestRecorder(repository payments.RequestRepository) RequestRecorder {
	return RequestRecorder{repository: repository}
}

type RequestRecorder struct {
	repository payments.RequestRepository
}

// RecordRequest hooks saves all outgoing requests before the http request is sent to the
// external api.
func (recorder RequestRecorder) RecordRequest(paymentID, requestID string) request.Hook {
	return request.Hook{Name: "RequestRecorder.RecordRequest", Fn: func(r *request.Request) {

		req := payments.Request{
			RequestID: requestID,
			PaymentID: paymentID,
			Partner:   r.Config.ServiceName,
			Status:    "received",
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
func (recorder RequestRecorder) UpdateRequestResponse(requestID string) request.Hook {
	return request.Hook{Name: "RequestRecorder.UpdateRequestResponse", Fn: func(r *request.Request) {
		opts := payments.OptionsUpdateRequest{}

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

			switch r.Error.(type) {
			case pkgerrors.Timeout:
				s := "timeout"
				opts.Status = &s
				break
			case pkgerrors.Temporary:
				s := "temporary_error"
				opts.Status = &s
				break
			default:
				s := "error"
				opts.Status = &s
				break
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

		s := "completed"
		opts.Status = &s
		opts.Response = resMap
	}}
}
