package request

import (
	"testing"
	"time"
)

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

func assertEquals(t *testing.T, expected, actual any) {
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
