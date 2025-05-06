package request

import (
	"errors"
	"log"
	"testing"
)

type MockHooks struct {
	str string
}

func (hooks *MockHooks) validate(r *Request) {
	hooks.str = hooks.str + "1"
}

func (hooks *MockHooks) build(r *Request) {
	hooks.str = hooks.str + "2"
}

func (hooks *MockHooks) send(r *Request) {
	hooks.str = hooks.str + "3"
}

func (hooks *MockHooks) unmarshal(r *Request) {
	hooks.str = hooks.str + "4"
	log.Println("unmarshal")
}

func (hooks *MockHooks) retry(r *Request) {
	hooks.str = hooks.str + "5"
}

func (hooks *MockHooks) complete(r *Request) {
	hooks.str = hooks.str + "6"
}

func TestRequest_Send(t *testing.T) {

	t.Run("test that calling order of hooks is correct", func(t *testing.T) {
		mockHooks := &MockHooks{}

		hooks := Hooks{
			Validate:  HookList{list: []Hook{{Fn: mockHooks.validate}}},
			Build:     HookList{list: []Hook{{Fn: mockHooks.build}}},
			Send:      HookList{list: []Hook{{Fn: mockHooks.send}}},
			Unmarshal: HookList{list: []Hook{{Fn: mockHooks.unmarshal}}},
			Retry:     HookList{list: []Hook{{Fn: mockHooks.retry}}},
			Complete:  HookList{list: []Hook{{Fn: mockHooks.complete}}},
		}

		// test that retry hooks are not called if no error occurs at send hooks
		t.Run("test order when no error occurs at send hooks", func(t *testing.T) {
			hooks := hooks.Copy()
			req := New(hooks, nil, nil, nil)

			err := req.Send()
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			if e, v := "12346", mockHooks.str; e != v {
				t.Errorf("expected %q, got %q", e, v)
			}
		})

		// test that retry hooks are called if error occurs at send hooks and request is retryable
		t.Run("test order when error occurs at send hooks and request is retryable", func(t *testing.T) {
			hooks := hooks.Copy()
			// mock an error at send hooks
			hooks.Send.PushBack(func(r *Request) {
				r.Error = errors.New("test error")

			})
			// mock a hook to update retry count
			hooks.Retry.PushBack(func(r *Request) {
				r.RetryConfig.RetryCount++
				t.Logf("retry count: %d", r.RetryConfig.RetryCount)
			})

			retryable := &RetryConfig{
				retryable:  true,
				MaxRetries: 5,
			}
			req := New(hooks, nil, nil, nil)
			req.RetryConfig = retryable

			err := req.Send()
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}

			if e, v := "123456", mockHooks.str; e != v {
				t.Errorf("expected %q, got %q", e, v)
			}
		})

	})
}
