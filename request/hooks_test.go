package request_test

import (
	"testing"

	"github.com/SirWaithaka/payments-api/request"
)

func TestHookList(t *testing.T) {
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
