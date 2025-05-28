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
		Unmarshal HookList
		Retry     HookList
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

// Clear clears the hook list
func (l *HookList) Clear() {
	l.list = l.list[0:0]
}

// Len returns the number of hooks in the list
func (l *HookList) Len() int {
	return len(l.list)
}

// PushBack pushes hook f to the back of the hook list.
func (l *HookList) PushBack(f func(*Request)) {
	l.PushBackHook(Hook{Fn: f})
}

// PushBackHook pushes hook h to the back of the hook list.
func (l *HookList) PushBackHook(h Hook) {
	if cap(l.list) == 0 {
		l.list = make([]Hook, 0, 5)
	}
	l.list = append(l.list, h)
}

func (l *HookList) PushFront(f func(*Request)) {
	l.PushFrontHook(Hook{Fn: f})
}

// PushFrontHook pushes hook h to the front of the hook list
func (l *HookList) PushFrontHook(h Hook) {
	if cap(l.list) == len(l.list) {
		// allocating a new list required
		l.list = append([]Hook{h}, l.list...)
	} else {
		// enough room to prepend into a list
		l.list = append(l.list, Hook{})
		copy(l.list[1:], l.list)
		l.list[0] = h
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
		Unmarshal: h.Unmarshal.copy(),
		Retry:     h.Retry.copy(),
		Complete:  h.Complete.copy(),
	}
}

// IsEmpty returns if there are no hooks in any of the hook lists.
func (h *Hooks) IsEmpty() bool {
	if h.Validate.Len() != 0 {
		return false
	}
	if h.Build.Len() != 0 {
		return false
	}
	if h.Send.Len() != 0 {
		return false
	}
	if h.Unmarshal.Len() != 0 {
		return false
	}
	if h.Retry.Len() != 0 {
		return false
	}
	if h.Complete.Len() != 0 {
		return false
	}
	return true
}
