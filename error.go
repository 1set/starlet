package starlet

import "fmt"

// ExecError is a custom error type for Starlet execution errors.
type ExecError struct {
	pkg   string // dependency source package or component name
	act   string // error happens when doing this action
	cause error  // the cause of the error
}

// Unwrap returns the cause of the error.
func (e ExecError) Unwrap() error {
	return e.cause
}

// Error returns the error message.
func (e ExecError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.pkg, e.act, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.pkg, e.act)
}

// helper functions

// errorStarlarkPanic creates an ExecError from a recovered panic value.
func errorStarlarkPanic(v interface{}) ExecError {
	return ExecError{
		pkg: `starlark`,
		act: fmt.Sprintf("panic: %v", v),
	}
}

// errorStarlarkError creates an ExecError from a Starlark error and an related action.
func errorStarlarkError(action string, err error) ExecError {
	return ExecError{
		pkg:   `starlark`,
		act:   action,
		cause: err,
	}
}

// errorStarlarkErrorf creates an ExecError for starlet with a formatted message.
func errorStarletErrorf(format string, args ...interface{}) ExecError {
	return ExecError{
		pkg: `starlet`,
		act: fmt.Sprintf(format, args...),
	}
}

// errorStarlightConvert creates an ExecError for starlight data conversion.
func errorStarlightConvert(name string, err error) ExecError {
	return ExecError{
		pkg:   `starlight`,
		act:   fmt.Sprintf("convert %s", name),
		cause: err,
	}
}
