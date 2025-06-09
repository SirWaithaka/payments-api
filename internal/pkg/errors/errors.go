package errors

// Payment describes an error that occurs during a payment operation.
// Can represent a payment failure
type Payment interface {
	PaymentError() bool
}

// NotFounder describes a not found exception during an operation
type NotFounder interface {
	NotFound() bool
}

// Timeout describes an error that occurs when an operation
// times out.
type Timeout interface {
	Timeout() bool
}

// Temporary describes an error that during an operation, but
// can be classified as temporary and perhaps the operation
// can be retried
type Temporary interface {
	Temporary() bool
}
