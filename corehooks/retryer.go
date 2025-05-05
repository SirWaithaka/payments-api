package corehooks

import (
	"errors"

	"github.com/SirWaithaka/payments-api/request"
)

func Retryer(err error, retryable *request.RetryConfig) {
	// do nothing if err or retryable is nil
	if err == nil || retryable == nil {
		return
	}

	// check if we have exceeded max allowed retries
	if retryable.RetryCount >= retryable.MaxRetries {
		return
	}

	// check if the error is not temporary
	var te interface{ Temporary() bool }
	if errors.As(err, &te) && !te.Temporary() {
		return
	}

}
