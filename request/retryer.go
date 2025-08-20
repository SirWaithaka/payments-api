package request

import (
	"errors"
	"math/rand/v2"
	"time"
)

var (
	DefaultRetryConfig = RetryConfig{
		InitialDelay:   1 * time.Second,
		Multiplier:     1,
		MaxDelay:       10 * time.Second,
		Jitter:         0,
		MaxRetries:     1,
		MaxElapsedTime: 5 * time.Second,
	}

	DefaultRetryer = retryer{}
)

type RetryConfig struct {
	// InitialDelay before the first retry.
	InitialDelay time.Duration
	// Multiplier to apply to the delay interval between retries.
	// Values >1 increase the delay interval exponentially, values <1 decrease it.
	Multiplier float64
	// MaximumDelay n at which to cap the delay interval between retries.
	// As the delay interval increases, the highest interval will be capped
	// at this value.
	MaxDelay time.Duration
	// Jitter is the amount of random jitter to apply to the delay interval
	// between retries to prevent exact timing between retries.
	// Valid values are between 0 and 1. A value of 0 disables Jitter.
	Jitter float64
	// CurrentDelay tracks the current delay interval between retries.
	CurrentDelay time.Duration

	// Number of retries attempted. Value should be incremented with each retry
	RetryCount int
	// Number of maximum allowed retries
	MaxRetries int
	// Maximum allowed time for retries
	MaxElapsedTime time.Duration

	// Additional API error codes that should be retried. IsErrorRetryable
	// will consider these codes in addition to its built-in cases.
	RetryErrorCodes []string

	//retryable bool
}

type Retryer interface {
	// Delay returns the duration to wait before making another attempt for the
	// failed request.
	Delay(*Request) time.Duration

	// Retryable returns true if the request should be retried
	Retryable(*Request) bool
}

// noOpRetryer is the default retryer used when a request is created without
// a retryer.
type noOpRetryer struct{}

// Delay returns the duration to wait before making another attempt for the
// failed request.
// Since noOpRetryer does not retry, it always returns 0.
func (r noOpRetryer) Delay(*Request) time.Duration {
	return 0
}

// Retryable returns true if the request should be retried.
// Since noOpRetryer does not retry, it always returns false.
func (r noOpRetryer) Retryable(*Request) bool {
	return false
}

type retryer struct{}

// Delay modifies the current delay with some jitter
func (r retryer) Delay(req *Request) time.Duration {
	if req.RetryConfig.CurrentDelay == 0 {
		req.RetryConfig.CurrentDelay = req.RetryConfig.InitialDelay
	}

	// calculate the next delay
	value := calculateRandomInterval(req.RetryConfig.CurrentDelay, req.RetryConfig.Jitter)
	return value
}

// Retryable performs validation checks on the retryer config to confirm if
// an operation is retry-able.
func (r retryer) Retryable(req *Request) bool {

	// check the number of max retries allowed
	if req.RetryConfig.MaxRetries == 0 {
		// return false if the number of max retries is 0
		return false
	}

	// check if we have exceeded max allowed retries
	if req.RetryConfig.RetryCount >= req.RetryConfig.MaxRetries {
		return false
	}

	// total elapsed time plus the next delay duration should never be > than MaxElapsedTime
	next := r.Delay(req)
	if req.RetryConfig.MaxElapsedTime > 0 && time.Since(req.AttemptTime)+next > req.RetryConfig.MaxElapsedTime {
		return false
	}

	// check if the error is not temporary
	var te interface{ Temporary() bool }
	if e := req.Error; e == nil || !errors.As(e, &te) || !te.Temporary() {
		return false
	}

	return true

}

func calculateRandomInterval(currDelay time.Duration, jitter float64) time.Duration {
	if jitter <= 0 {
		// do not introduce randomness if jitter is less or equal to 0
		return currDelay
	}

	delta := float64(currDelay) * jitter
	minDelay := float64(currDelay) - delta
	maxDelay := float64(currDelay) + delta

	random := rand.Float64()

	// get a random value between minDelay and maxDelay
	return time.Duration(minDelay + (random * (maxDelay - minDelay)))
}
