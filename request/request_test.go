package request

import (
	"errors"
	"testing"
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
}
