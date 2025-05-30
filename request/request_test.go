package request

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

type FakeTemporaryError struct {
	error
	temporary bool
}

func (e FakeTemporaryError) Temporary() bool {
	return e.temporary
}

type MockHooks struct {
	str string
}

func (hooks *MockHooks) validate(r *Request) {
	hooks.str = hooks.str + "validate:"
}

func (hooks *MockHooks) build(r *Request) {
	hooks.str = hooks.str + "build:"
}

func (hooks *MockHooks) send(r *Request) {
	hooks.str = hooks.str + "send:"
}

func (hooks *MockHooks) unmarshal(r *Request) {
	hooks.str = hooks.str + "unmarshal:"
}

func (hooks *MockHooks) retry(r *Request) {
	hooks.str = hooks.str + "retry:"
}

func (hooks *MockHooks) complete(r *Request) {
	hooks.str = hooks.str + "complete:"
}

func TestRequest_New(t *testing.T) {

	t.Run("test retryer is not nil", func(t *testing.T) {
		req := New(Config{}, Hooks{}, nil, nil, nil, nil)
		if req.Retryer == nil {
			t.Errorf("expected non-nil retryer, got nil")
		}
	})

	t.Run("test http request url", func(t *testing.T) {
		tcs := map[string]struct {
			Endpoint      string
			Path          string
			ExpectedPath  string
			ExpectedQuery string
		}{
			"no http Path": {
				Endpoint:      "https://example.com",
				Path:          "/",
				ExpectedPath:  "/",
				ExpectedQuery: "",
			},
			"with path in endpoint": {
				Endpoint:      "https://example.com/foo",
				Path:          "",
				ExpectedPath:  "/foo",
				ExpectedQuery: "",
			},
			"with query in path": {
				Endpoint:      "https://example.com",
				Path:          "/foo?bar=baz",
				ExpectedPath:  "/foo",
				ExpectedQuery: "bar=baz",
			},
			"with path in endpoint and query": {
				Endpoint:      "https://example.com/foo?bar=baz",
				Path:          "/qux",
				ExpectedPath:  "/foo/qux",
				ExpectedQuery: "",
			},
			"with query in path and endpoint": {
				Endpoint:      "https://example.com/?bar=baz",
				Path:          "/?foo=qux",
				ExpectedPath:  "/",
				ExpectedQuery: "foo=qux",
			},
		}

		for name, tc := range tcs {
			t.Run(name, func(t *testing.T) {
				op := &Operation{Name: "FooBar", Path: tc.Path}
				req := New(Config{Endpoint: tc.Endpoint}, Hooks{}, nil, op, nil, nil)
				// assert results to expected values
				assertEquals(t, tc.ExpectedPath, req.Request.URL.Path)
				assertEquals(t, tc.ExpectedQuery, req.Request.URL.RawQuery)
				assertEquals(t, http.MethodPost, req.Request.Method)
			})
		}
	})
}

func TestRequest_Send(t *testing.T) {

	t.Run("test that calling order of hooks is correct", func(t *testing.T) {

		// test that retry hooks are not called if no error occurs at send hooks
		t.Run("test order when no error occurs at send hooks", func(t *testing.T) {
			mockHooks := MockHooks{}

			hooks := Hooks{
				Validate:  HookList{list: []Hook{{Fn: mockHooks.validate}}},
				Build:     HookList{list: []Hook{{Fn: mockHooks.build}}},
				Send:      HookList{list: []Hook{{Fn: mockHooks.send}}},
				Unmarshal: HookList{list: []Hook{{Fn: mockHooks.unmarshal}}},
				Retry:     HookList{list: []Hook{{Fn: mockHooks.retry}}},
				Complete:  HookList{list: []Hook{{Fn: mockHooks.complete}}},
			}

			req := New(Config{}, hooks.Copy(), nil, nil, nil, nil)

			err := req.Send()
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			expected := "validate:build:send:unmarshal:complete:"
			if e, v := expected, mockHooks.str; e != v {
				t.Errorf("expected %q, got %q", e, v)
			}
		})

		// test that retry hooks are called if error occurs at send hooks
		t.Run("test order when error occurs at send hooks", func(t *testing.T) {
			mockHooks := MockHooks{}

			hooks := Hooks{
				Validate:  HookList{list: []Hook{{Fn: mockHooks.validate}}},
				Build:     HookList{list: []Hook{{Fn: mockHooks.build}}},
				Send:      HookList{list: []Hook{{Fn: mockHooks.send}}},
				Unmarshal: HookList{list: []Hook{{Fn: mockHooks.unmarshal}}},
				Retry:     HookList{list: []Hook{{Fn: mockHooks.retry}}},
				Complete:  HookList{list: []Hook{{Fn: mockHooks.complete}}},
			}

			// mock an error at send hooks
			hooks.Send.PushBack(func(r *Request) {
				// create a temporary error
				tempErr := FakeTemporaryError{error: errors.New("fake error"), temporary: true}
				r.Error = tempErr
			})
			req := New(Config{}, hooks, nil, nil, nil, nil)

			err := req.Send()
			if err == nil {
				t.Errorf("expected error, got nil")
			}

			expected := "validate:build:send:complete:"
			if e, v := expected, mockHooks.str; e != v {
				t.Errorf("expected %q, got %q", e, v)
			}
		})

	})

	t.Run("test that retryable requests are retried", func(t *testing.T) {
		hooks := Hooks{}

		cfg := RetryConfig{
			MaxRetries:     1,
			InitialDelay:   100 * time.Millisecond,
			Jitter:         0.1,
			MaxElapsedTime: 1 * time.Second,
		}

		// mock an error using send hook
		hooks.Send.PushBack(func(r *Request) {
			// create a temporary error
			tempErr := FakeTemporaryError{error: errors.New("fake error"), temporary: true}
			r.Error = tempErr
		})
		hooks.Retry.PushBack(func(r *Request) {
			r.RetryConfig.RetryCount++
		})

		// create an instance of retryer
		ret := retryer{}
		req := New(Config{}, hooks, ret, nil, nil, nil)
		req.WithRetryConfig(cfg)

		if err := req.Send(); err == nil {
			t.Errorf("expected error, got nil")
		}

		// confirm that request was retried
		if e, v := cfg.MaxRetries, req.RetryConfig.RetryCount; e != v {
			t.Errorf("expected %v, got %v", e, v)
		}
	})
}
