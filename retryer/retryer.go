package retryer

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/SirWaithaka/payments-api/request"
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
)

type (
	Operation func() error

	timer struct {
		timer *time.Timer
	}

	RetryConfig struct {
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
		// Number of retries attempted. Value should be incremented with each retry
		RetryCount int
		// Number of maximum allowed retries
		MaxRetries int
		// Maximum allowed time for retries
		MaxElapsedTime time.Duration

		// Additional API error codes that should be retried. IsErrorRetryable
		// will consider these codes in addition to its built-in cases.
		RetryErrorCodes []string

		retryable bool
	}
)

func (t *timer) C() <-chan time.Time {
	return t.timer.C
}

func (t *timer) Start(dur time.Duration) {
	if t.timer == nil {
		t.timer = time.NewTimer(dur)
	} else {
		t.timer.Reset(dur)
	}
}

// Stop is used to free resources when timer is no longer used
func (t *timer) Stop() {
	if t.timer != nil {
		t.timer.Stop()
	}
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

func New(cfg RetryConfig) DefaultRetryer {
	return DefaultRetryer{config: DefaultRetryConfig, timer: &timer{timer: time.NewTimer(cfg.InitialDelay)}}
}

type DefaultRetryer struct {
	config       RetryConfig
	timer        *timer
	currentDelay time.Duration
}

// Delay modifies the current delay with some jitter
func (r *DefaultRetryer) Delay() time.Duration {
	if r.currentDelay == 0 {
		r.currentDelay = r.config.InitialDelay
	}

	// calculate the next delay
	value := calculateRandomInterval(r.currentDelay, r.config.Jitter)
	return value
}

func (r *DefaultRetryer) nextDelay() time.Duration {
	// using the multiplier, calculate the next delay
	next := float64(r.currentDelay) * r.config.Multiplier
	// if the next calculated delay is greater than max delay, return max delay
	if next > float64(r.config.MaxDelay) {
		return r.config.MaxDelay
	}
	return time.Duration(next)

}

// Retryable performs validation checks on the retryer config to confirm if
// an operation is retry-able.
func (r *DefaultRetryer) Retryable(req *request.Request) bool {

	// check the number of max retries allowed
	if r.config.MaxRetries == 0 {
		// return false if the number of max retries is 0
		return false
	}

	// check if we have exceeded max allowed retries
	if r.config.RetryCount >= r.config.MaxRetries {
		return false
	}

	// total elapsed time plus the next delay duration should never be > than MaxElapsedTime
	next := r.Delay()
	if r.config.MaxElapsedTime > 0 && time.Since(req.AttemptTime)+next > r.config.MaxElapsedTime {
		return false
	}

	// check if the error is not temporary
	var te interface{ Temporary() bool }
	if errors.As(req.Error, &te) && !te.Temporary() {
		return false
	}

	return true

}

func (r *DefaultRetryer) Retry(ctx context.Context) error {
	// increment retry count
	r.config.RetryCount += 1

	if ctx != nil {
		// Stop retrying if context is canceled
		if cerr := context.Cause(ctx); cerr != nil {
			return cerr
		}
	}

	// start the timer and wait
	r.timer.Start(r.Delay())
	// wait for timer to complete or context Done signal
	select {
	case <-r.timer.C():
	case <-ctx.Done():
		return context.Cause(ctx)
	}

	// get the next delay duration
	next := r.nextDelay()
	r.currentDelay = next

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
