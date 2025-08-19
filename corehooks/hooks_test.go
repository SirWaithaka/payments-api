package corehooks_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"

	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/request"
)

func TestAddScheme(t *testing.T) {
	tcs := map[string]struct {
		Endpoint   string
		DisableSSL bool
		Expected   string
	}{
		"with no scheme": {
			Endpoint:   "example.com",
			DisableSSL: false,
			Expected:   "https://example.com",
		},
		"disable ssl": {
			Endpoint:   "example.com",
			DisableSSL: true,
			Expected:   "http://example.com",
		},
		"with ssl scheme": {
			Endpoint:   "https://example.com",
			DisableSSL: false,
			Expected:   "https://example.com",
		},
		"with no ssl scheme": {
			Endpoint:   "http://example.com",
			DisableSSL: false,
			Expected:   "http://example.com",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			endpoint := corehooks.AddScheme(tc.Endpoint, tc.DisableSSL)
			assert.Equal(t, tc.Expected, endpoint)
		})
	}
}

type testSendHandlerTransport struct{ timeout time.Duration }

func (t *testSendHandlerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.timeout > 0 {
		time.Sleep(t.timeout)
	}
	return nil, errors.New("mock error")
}

func TestSendHook(t *testing.T) {

	t.Run("test redirect", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/redirect":
				u := *r.URL
				u.Path = "/home"
				w.Header().Set("Location", u.String())
				w.WriteHeader(http.StatusTemporaryRedirect)

			case "/home":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			}
		}))
		defer server.Close()

		tcs := map[string]struct {
			Redirect       bool
			ExpectedStatus int
		}{
			"redirect": {
				Redirect:       true,
				ExpectedStatus: http.StatusOK,
			},
			"no redirect": {
				Redirect:       false,
				ExpectedStatus: http.StatusTemporaryRedirect,
			},
		}

		for name, tc := range tcs {
			t.Run(name, func(t *testing.T) {
				cfg := request.Config{Endpoint: server.URL, DisableSSL: true, HTTPClient: http.DefaultClient}
				op := &request.Operation{Name: "FooBar", Path: "/redirect"}

				hooks := request.Hooks{}
				hooks.Send.PushBackHook(corehooks.SendHook)

				cfg.DisableFollowRedirects = !tc.Redirect

				req := request.New(cfg, hooks, nil, op, nil, nil)
				if err := req.Send(); err != nil {
					t.Errorf("expected nil error, got %v", err)
				}

				// check response status
				assert.Equal(t, tc.ExpectedStatus, req.Response.StatusCode)

			})
		}
	})

	t.Run("test handle send error", func(t *testing.T) {

		t.Run("transport error", func(t *testing.T) {
			client := &http.Client{Transport: &testSendHandlerTransport{}}
			op := &request.Operation{Name: "Operation"}

			hooks := request.Hooks{}
			hooks.Send.PushBackHook(corehooks.SendHook)
			req := request.New(request.Config{HTTPClient: client}, hooks, nil, op, nil, nil)

			if err := req.Send(); err == nil {
				t.Errorf("expected error, got nil")
			}
			if req.Response == nil {
				t.Errorf("expected response, got nil")
			}
		})

		t.Run("url.Error timeout", func(t *testing.T) {
			client := &http.Client{
				Timeout:   100 * time.Millisecond,
				Transport: &testSendHandlerTransport{timeout: 500 * time.Millisecond},
			}
			op := &request.Operation{Name: "Operation"}

			hooks := request.Hooks{}
			hooks.Send.PushBackHook(corehooks.SendHook)
			req := request.New(request.Config{HTTPClient: client}, hooks, nil, op, nil, nil)

			if err := req.Send(); err == nil {
				t.Errorf("expected error, got nil")
			}
			if req.Response == nil {
				t.Errorf("expected response, got nil")
			}
			assert.Equal(t, 0, req.Response.StatusCode)
		})
	})
}

func TestSetRequestID(t *testing.T) {
	rid := xid.New().String()

	generator := func() string {
		return rid
	}

	// build request hooks
	hooks := request.Hooks{}
	hooks.Build.PushFrontHook(corehooks.SetRequestID(generator))
	hooks.Complete.PushFront(func(r *request.Request) {
		assert.Equal(t, rid, r.Config.RequestID)
	})

	req := request.New(request.Config{}, hooks, nil, nil, nil, nil)
	err := req.Send()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}
