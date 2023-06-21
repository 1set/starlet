package starlet

import "fmt"

// ExecError is a custom error type for Starlet execution errors.
type ExecError struct {
	source  string // dependency source package or component name
	message string
	cause   error
}

// Unwrap returns the cause of the error.
func (e ExecError) Unwrap() error {
	return e.cause
}

// Error returns the error message.
func (e ExecError) Error() string {
	//if e.cause != nil {
	//	return fmt.Sprintf("%s: %s: %v", e.source, e.message, e.cause)
	//}
	return fmt.Sprintf("%s: %s", e.source, e.message)
}

func errorStarlarkPanic(v interface{}) ExecError {
	ie, _ := v.(error)
	return ExecError{
		source:  `starlark`,
		message: fmt.Sprintf("panic: %v", v),
		cause:   ie,
	}
}
