package request

import (
	"strings"
)

type (
	Hook struct {
		Name string
		Fn   Option
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
	if l.Len() == 0 {
		return n
	}

	n.list = append(make([]Hook, 0, l.Len()), l.list...)
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
func (l *HookList) PushBack(f Option) {
	l.PushBackHook(Hook{Fn: f, Name: "__anon"})
}

// PushBackHook pushes hook h to the back of the hook list.
func (l *HookList) PushBackHook(h Hook) {
	if cap(l.list) == 0 {
		l.list = make([]Hook, 0, 5)
	}
	l.list = append(l.list, h)
}

func (l *HookList) PushFront(f Option) {
	l.PushFrontHook(Hook{Fn: f, Name: "__anon"})
}

// PushFrontHook pushes hook h to the front of the hook list
func (l *HookList) PushFrontHook(h Hook) {
	if cap(l.list) == l.Len() {
		// allocating a new list required
		l.list = append([]Hook{h}, l.list...)
	} else {
		// enough room to prepend into a list
		l.list = append(l.list, Hook{})
		copy(l.list[1:], l.list)
		l.list[0] = h
	}
}

// Remove removes a Hook by name
func (l *HookList) Remove(name string) {
	for i := 0; i < l.Len(); i++ {
		m := l.list[i]
		if m.Name == name {
			// shift slice elements in place
			copy(l.list[i:], l.list[i+1:])
			// zero last element
			l.list[l.Len()-1] = Hook{}
			// clear last element
			l.list = l.list[:l.Len()-1]

			i--
		}
	}
}

// RemoveHook removes Hook h
func (l *HookList) RemoveHook(h Hook) {
	l.Remove(h.Name)
}

func (l *HookList) Swap(name string, replace Hook) {
	for i := 0; i < l.Len(); i++ {
		if l.list[i].Name == name {
			l.list[i] = replace
		}
	}
}

// Run executes all handlers in the list with a given request object
func (l *HookList) Run(r *Request) {
	for _, h := range l.list {
		h.Fn(r)
	}
}

func (l *HookList) Debug() string {
	hooks := make([]string, l.Len())
	for _, h := range l.list {
		name := h.Name
		if name == "" {
			name = "__anonymous"
		}
		hooks = append(hooks, name)
	}
	return strings.Join(hooks, " ")
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

func (h *Hooks) Debug() map[string]string {
	hooks := make(map[string]string)
	hooks["validate"] = h.Validate.Debug()
	hooks["build"] = h.Build.Debug()
	hooks["send"] = h.Send.Debug()
	hooks["unmarshal"] = h.Unmarshal.Debug()
	hooks["retry"] = h.Retry.Debug()
	hooks["complete"] = h.Complete.Debug()
	return hooks
}
