package request

import (
	"time"
)

type Retryer interface {
	// Delay returns the duration to wait before making another attempt for the
	// failed request.
	Delay() time.Duration

	// Retryable returns true if the request should be retried
	Retryable(r *Request) bool
}

// noOpRetryer is the default retryer used when a request is created without
// a retryer.
type noOpRetryer struct{}

// Delay returns the duration to wait before making another attempt for the
// failed request.
// Since noOpRetryer does not retry, it always returns 0.
func (r noOpRetryer) Delay() time.Duration {
	return 0
}

// Retryable returns true if the request should be retried.
// Since noOpRetryer does not retry, it always returns false.
func (r noOpRetryer) Retryable(*Request) bool {
	return false
}
