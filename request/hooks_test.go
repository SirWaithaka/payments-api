package request_test

import (
	"testing"

	"github.com/SirWaithaka/payments-api/request"
)

func TestHookList_Run(t *testing.T) {
	r := &request.Request{}
	h := request.HookList{}

	val := ""
	h.PushBack(func(r *request.Request) {
		val += "a"
		r.Params = val
	})
	h.Run(r)

	// assert
	if e, v := "a", val; e != v {
		t.Errorf("expected %q, got %q", e, v)
	}
	if e, v := "a", r.Params.(string); e != v {
		t.Errorf("expected %q, got %q", e, v)
	}
}

func TestHooksList_Remove(t *testing.T) {
	hooks := request.HookList{}
	hook := request.Hook{Name: "Foo", Fn: func(r *request.Request) {}}
	hook2 := request.Hook{Name: "Bar", Fn: func(r *request.Request) {}}
	// add 4 hooks
	hooks.PushFrontHook(hook)
	hooks.PushFrontHook(hook2)
	hooks.PushFrontHook(hook)
	hooks.PushFront(func(r *request.Request) {})

	// assert for 4 hooks
	if e, v := 4, hooks.Len(); e != v {
		t.Errorf("expected %d, got %d", e, v)
	}

	// remove hook
	hooks.RemoveHook(hook)
	if e, v := 2, hooks.Len(); e != v {
		t.Errorf("expected %d, got %d", e, v)
	}
}
