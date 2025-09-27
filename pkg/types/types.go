package types

// Pointer returns a pointer to the given value
func Pointer[T any](v T) *T {
	return &v
}
