package starlet

import (
	"fmt"

	"go.starlark.net/starlark"
)

// ExecError is a custom error type for Starlet execution errors.
type ExecError struct {
	pkg   string // dependency source package or component name
	act   string // error happens when doing this action
	cause error  // the cause of the error
	hint  string // additional hint for the error
}

// Unwrap returns the cause of the error.
func (e ExecError) Unwrap() error {
	return e.cause
}

// Error returns the error message.
func (e ExecError) Error() string {
	s := fmt.Sprintf("%s: %s: %v", e.pkg, e.act, e.cause)
	if e.hint != "" {
		s += "\n" + e.hint
	}
	return s
}

// helper functions

// errorStarlarkPanic creates an ExecError from a recovered panic value.
func errorStarlarkPanic(action string, v interface{}) ExecError {
	return ExecError{
		pkg:   `starlark`,
		act:   action,
		cause: fmt.Errorf("panic: %v", v),
	}
}

// errorStarlarkError creates an ExecError from a Starlark error and an related action.
func errorStarlarkError(action string, err error) ExecError {
	// don't wrap if the error is already an ExecError
	if e, ok := err.(ExecError); ok {
		return e
	}
	// parse error from Starlark
	var hint string
	if se, ok := err.(*starlark.EvalError); ok {
		hint = se.Backtrace()
	}
	return ExecError{
		pkg:   `starlark`,
		act:   action,
		cause: err,
		hint:  hint,
	}
}

// errorStarletError creates an ExecError for starlet.
func errorStarletError(action string, err error) ExecError {
	// don't wrap if the error is already an ExecError
	if e, ok := err.(ExecError); ok {
		return e
	}
	return ExecError{
		pkg:   `starlet`,
		act:   action,
		cause: err,
	}
}

// errorStarletErrorf creates an ExecError for starlet with a formatted message.
func errorStarletErrorf(action string, format string, args ...interface{}) ExecError {
	return ExecError{
		pkg:   `starlet`,
		act:   action,
		cause: fmt.Errorf(format, args...),
	}
}

// errorStarlightConvert creates an ExecError for starlight data conversion.
func errorStarlightConvert(target string, err error) ExecError {
	// don't wrap if the error is already an ExecError
	if e, ok := err.(ExecError); ok {
		return e
	}
	return ExecError{
		pkg:   `starlight`,
		act:   fmt.Sprintf("convert %s", target),
		cause: err,
	}
}
