package request

import (
	"errors"
	"testing"
	"time"
)

func assertEquals(t *testing.T, expected, actual any) {
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestCalculateRandomInterval(t *testing.T) {

	t.Run("test that given jitter is 0 the returned value is equal to the delay", func(t *testing.T) {

		dur := calculateRandomInterval(500*time.Millisecond, 0)
		assertEquals(t, dur, 500*time.Millisecond)

	})

	t.Run("test that given jitter is not equal to 0 the returned value is within expected range", func(t *testing.T) {
		// given a delay and jitter, calculate the expected interval
		intervalF := func(delay time.Duration, jitter float64) (a float64, b float64) {
			a = float64(delay) - float64(delay)*jitter
			b = float64(delay) + float64(delay)*jitter
			return
		}

		// generate some test cases
		tcs := []struct {
			delay  time.Duration
			jitter float64
		}{
			{delay: 500 * time.Millisecond, jitter: 0.1},
			{delay: 500 * time.Millisecond, jitter: 0.2},
			{delay: 500 * time.Millisecond, jitter: 0.3},
			{delay: 500 * time.Millisecond, jitter: 0.4},

			{delay: 100 * time.Millisecond, jitter: 0.1},
			{delay: 200 * time.Millisecond, jitter: 0.2},
			{delay: 300 * time.Millisecond, jitter: 0.3},
			{delay: 400 * time.Millisecond, jitter: 0.4},
		}

		for _, tc := range tcs {
			// calculate the random interval and assert that it is within the expected range
			dur := calculateRandomInterval(tc.delay, tc.jitter)
			miN, maX := intervalF(tc.delay, tc.jitter)
			if dur < time.Duration(miN) || dur > time.Duration(maX) {
				t.Errorf("expected value to be in range %v-%v, got %v", miN, maX, dur)
			}
		}

	})
}

func TestDefaultRetryer_Delay(t *testing.T) {
	// test that the delay value returned has jitter within the expected range
	t.Run("test that the delay value returned has jitter within the expected range", func(t *testing.T) {
		mockHooks := MockHooks{}

		hooks := Hooks{
			Validate:  HookList{list: []Hook{{Fn: mockHooks.validate}}},
			Build:     HookList{list: []Hook{{Fn: mockHooks.build}}},
			Send:      HookList{list: []Hook{{Fn: mockHooks.send}}},
			Unmarshal: HookList{list: []Hook{{Fn: mockHooks.unmarshal}}},
			Retry:     HookList{list: []Hook{{Fn: mockHooks.retry}}},
			Complete:  HookList{list: []Hook{{Fn: mockHooks.complete}}},
		}

		cfg := RetryConfig{
			MaxRetries:     1,
			InitialDelay:   100 * time.Millisecond,
			Jitter:         0.1,
			MaxElapsedTime: 1 * time.Second,
		}

		req := New(Config{}, hooks, DefaultRetryer, nil, nil, nil)
		req.WithRetryConfig(cfg)

		//delay := DefaultRetryer.Delay(req)

		//// mock an error with a send hook
		//// mock an error at send hooks
		//hooks.Send.PushBack(func(r *Request) {
		//	// create a temporary error
		//	tempErr := FakeTemporaryError{error: errors.New("fake error"), temporary: true}
		//	r.Error = tempErr
		//})
		//
		//// send request
		//if err := req.Send(); err != nil {
		//	t.Errorf("expected nil error, got %v", err)
		//}
		//
		//t.Log(req.RetryConfig.RetryCount)

	})
}

func TestDefaultRetryer_Retryable(t *testing.T) {

	t.Run("test that the request is not retryable if no retry config is set", func(t *testing.T) {
		hooks := Hooks{}

		// create an instance of retryer
		ret := &retryer{}
		req := New(Config{}, hooks, ret, nil, nil, nil)
		isRetryable := ret.Retryable(req)

		if e, v := false, isRetryable; e != v {
			t.Errorf("expected %v, got %v", e, v)
		}

	})

	t.Run("test that the request is not retryable if RetryConfig.MaxRetries is 0", func(t *testing.T) {
		hooks := Hooks{}

		cfg := RetryConfig{
			MaxRetries:     0,
			InitialDelay:   100 * time.Millisecond,
			Jitter:         0.1,
			MaxElapsedTime: 1 * time.Second,
		}

		// create an instance of retryer
		ret := &retryer{}
		req := New(Config{}, hooks, ret, nil, nil, nil)
		req.WithRetryConfig(cfg)

		isRetryable := ret.Retryable(req)
		if e, v := false, isRetryable; e != v {
			t.Errorf("expected %v, got %v", e, v)
		}

	})

	t.Run("test that the request is not retryable if retry count equal max tries", func(t *testing.T) {
		hooks := Hooks{}

		cfg := RetryConfig{
			RetryCount:     5,
			MaxRetries:     5,
			InitialDelay:   100 * time.Millisecond,
			Jitter:         0.1,
			MaxElapsedTime: 1 * time.Second,
		}

		// create an instance of retryer
		ret := &retryer{}
		req := New(Config{}, hooks, ret, nil, nil, nil)
		req.WithRetryConfig(cfg)

		isRetryable := ret.Retryable(req)
		if e, v := false, isRetryable; e != v {
			t.Errorf("expected %v, got %v", e, v)
		}

	})

	t.Run("test that the request is not retryable if total retry duration is more than MaxElapsedTime", func(t *testing.T) {
		hooks := Hooks{}

		attemptTime := time.Now().Add(-1 * time.Second)

		cfg := RetryConfig{
			RetryCount:     1,
			MaxRetries:     5,
			InitialDelay:   100 * time.Millisecond,
			Jitter:         0.1,
			MaxElapsedTime: 1 * time.Second,
		}

		// create an instance of retryer
		ret := &retryer{}
		req := New(Config{}, hooks, ret, nil, nil, nil)
		req.WithRetryConfig(cfg)
		req.AttemptTime = attemptTime

		isRetryable := ret.Retryable(req)
		if e, v := false, isRetryable; e != v {
			t.Errorf("expected %v, got %v", e, v)
		}

	})

	t.Run("test that the request is not retryable if Request.Error is nil", func(t *testing.T) {
		hooks := Hooks{}

		cfg := RetryConfig{
			MaxRetries:     1,
			InitialDelay:   100 * time.Millisecond,
			Jitter:         0.1,
			MaxElapsedTime: 1 * time.Second,
		}

		// create an instance of retryer
		ret := &retryer{}
		req := New(Config{}, hooks, ret, nil, nil, nil)
		req.WithRetryConfig(cfg)

		isRetryable := ret.Retryable(req)
		if e, v := false, isRetryable; e != v {
			t.Errorf("expected %v, got %v", e, v)
		}

	})

	t.Run("test that the request is not retryable if Request.Error is not Temporary type", func(t *testing.T) {
		hooks := Hooks{}

		cfg := RetryConfig{
			MaxRetries:     1,
			InitialDelay:   100 * time.Millisecond,
			Jitter:         0.1,
			MaxElapsedTime: 1 * time.Second,
		}

		// create an instance of retryer
		ret := &retryer{}
		req := New(Config{}, hooks, ret, nil, nil, nil)
		req.WithRetryConfig(cfg)

		req.Error = errors.New("fake error")

		isRetryable := ret.Retryable(req)
		if e, v := false, isRetryable; e != v {
			t.Errorf("expected %v, got %v", e, v)
		}

	})

	t.Run("test that the request is not retryable if Request.Error is Temporary type but not temporary", func(t *testing.T) {
		hooks := Hooks{}

		cfg := RetryConfig{
			MaxRetries:     1,
			InitialDelay:   100 * time.Millisecond,
			Jitter:         0.1,
			MaxElapsedTime: 1 * time.Second,
		}

		// create an instance of retryer
		ret := &retryer{}
		req := New(Config{}, hooks, ret, nil, nil, nil)
		req.WithRetryConfig(cfg)

		req.Error = FakeTemporaryError{error: errors.New("fake error"), temporary: false}

		isRetryable := ret.Retryable(req)
		if e, v := false, isRetryable; e != v {
			t.Errorf("expected %v, got %v", e, v)
		}

	})

	t.Run("test that the request is retryable if Request.Error is Temporary", func(t *testing.T) {
		hooks := Hooks{}

		cfg := RetryConfig{
			MaxRetries:     1,
			InitialDelay:   100 * time.Millisecond,
			Jitter:         0.1,
			MaxElapsedTime: 1 * time.Second,
		}

		// create an instance of retryer
		ret := &retryer{}
		req := New(Config{}, hooks, ret, nil, nil, nil)
		req.WithRetryConfig(cfg)

		req.Error = FakeTemporaryError{error: errors.New("fake error"), temporary: true}

		isRetryable := ret.Retryable(req)
		if e, v := true, isRetryable; e != v {
			t.Errorf("expected %v, got %v", e, v)
		}

	})

}
