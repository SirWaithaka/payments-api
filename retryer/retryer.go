package retryer

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

var (
	DefaultDelayConfig = DelayConfig{
		InitialDelay: 1 * time.Second,
		Multiplier:   1,
		MaxDelay:     10 * time.Second,
		Jitter:       0,
	}
)

type DelayConfig struct {
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
	// Valid values are between 0 and 1. A value of 0 disables jitter.
	Jitter float64
}

type RetryConfig struct {
	retryable bool
	// Number of retries attempted
	RetryCount int
	// Number of maximum allowed retries
	MaxRetries int

	// Additional API error codes that should be retried. IsErrorRetryable
	// will consider these codes in addition to its built-in cases.
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

func (r *RetryConfig) SetRetryable(retry bool) {
	r.retryable = retry
}

type DefaultRetryer struct {
	delay DelayConfig
	retry RetryConfig

	currentDelay time.Duration
}

func (r *DefaultRetryer) Delay() time.Duration {
	if r.currentDelay == 0 {
		r.currentDelay = r.delay.InitialDelay
	}

	// calculate the next delay
	value := calculateRandomInterval(r.currentDelay, r.delay.Jitter)
	r.incrementCurrentDelay()
	return value
}

func (r *DefaultRetryer) incrementCurrentDelay() {
	// validate that the next value for currentDelay is not greater than the max delay
	next := float64(r.currentDelay) * r.delay.Multiplier
	if next > float64(r.delay.MaxDelay) {
		r.currentDelay = r.delay.MaxDelay
	} else {
		r.currentDelay = time.Duration(next)
	}
}

func (r *DefaultRetryer) Retryable() bool {

	// check the number of max retries allowed
	if r.retry.MaxRetries == 0 {
		// return false if the number of max retries is 0
		return false
	}

	// check if we have exceeded max allowed retries
	if r.retry.RetryCount >= r.retry.MaxRetries {
		return false
	}

	return true

}

func (r *DefaultRetryer) Retry(ctx context.Context) error {

	for r.retry.RetryCount < r.retry.MaxRetries {
		// do something
	}

	return nil
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

func Retryer(err error, retryable *RetryConfig) {
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
