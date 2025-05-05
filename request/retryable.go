package request

import (
	"time"
)

type RetryConfig struct {
	retryable bool
	// Number of retries attempted
	RetryCount int
	// Number of maximum allowed retries
	MaxRetries int
	// Duration to delay before making a retry
	RetryDelay time.Duration

	// Additional API error codes that should be retried. IsErrorRetryable
	// will consider these codes in addition to its built in cases.
	RetryErrorCodes []string
}

func (r *RetryConfig) IsRetryable() bool {
	if r.retryable {
		return true
	}

	if r.RetryCount >= r.MaxRetries {
		return false
	}

	// false by default
	return false
}
