package corehooks_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SirWaithaka/payments-api/corehooks"
	"github.com/SirWaithaka/payments-api/request"
)

func assertEquals(t *testing.T, expected, actual any) {
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

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
			if e, v := tc.Expected, endpoint; e != v {
				assertEquals(t, e, v)
			}
		})
	}
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
				assertEquals(t, tc.ExpectedStatus, req.Response.StatusCode)

			})
		}

	})

}
