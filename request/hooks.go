package request

type (
	Hook struct {
		//Name string
		Fn func(*Request)
	}

	HookList struct {
		list []Hook
	}

	Hooks struct {
		Validate  HookList
		Build     HookList
		Send      HookList
		Retry     HookList
		Unmarshal HookList
		Complete  HookList
	}
)

func (l *HookList) copy() HookList {
	n := HookList{}
	if len(l.list) == 0 {
		return n
	}

	n.list = append(make([]Hook, 0, len(l.list)), l.list...)
	return n
}

// Clear clears the hooks list
func (l *HookList) Clear() {
	l.list = l.list[0:0]
}

// Len returns the number of hooks in the list
func (l *HookList) Len() int {
	return len(l.list)
}

// PushBack pushes hook h to the back of the hooks list.
func (l *HookList) PushBack(f func(*Request)) {
	if cap(l.list) == 0 {
		l.list = make([]Hook, 0, 5)
	}
	l.list = append(l.list, Hook{f})
}

// PushFront pushes hook h to the front of the hook list
func (l *HookList) PushFront(f func(*Request)) {
	if cap(l.list) == len(l.list) {
		// allocating new list required
		l.list = append([]Hook{{f}}, l.list...)
	} else {
		// enough room to prepend into list
		l.list = append(l.list, Hook{})
		copy(l.list[1:], l.list)
		l.list[0] = Hook{f}
	}
}

// Run executes all handlers in the list with a given request object
func (l *HookList) Run(r *Request) {
	for _, h := range l.list {
		h.Fn(r)
	}
}

// Copy returns a copy of these hooks' lists
func (h *Hooks) Copy() Hooks {
	return Hooks{
		Validate:  h.Validate.copy(),
		Build:     h.Build.copy(),
		Send:      h.Send.copy(),
		Retry:     h.Retry.copy(),
		Unmarshal: h.Unmarshal.copy(),
		Complete:  h.Complete.copy(),
	}
}
