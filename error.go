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

// ModuleNotFoundError is returned by load() when the named module is not
// present in any source available to the machine: it is neither a builtin
// or custom loader configured for this machine, nor a script file on the
// configured filesystem. Hosts can detect it with errors.As through the
// Starlark error chain.
type ModuleNotFoundError struct {
	Name string
}

// Error returns the error message.
func (e ModuleNotFoundError) Error() string {
	return fmt.Sprintf("module %q not found in builtin modules, custom loaders, or the script filesystem", e.Name)
}

// ModuleWithheldError marks a module that exists but is deliberately not
// made available to the current machine: a known builtin that was not
// enabled, or a module blocked by a host-side policy layer (which can
// return this type from its own loaders). It lets hosts and script authors
// tell a misspelled module apart from a forbidden one; detect it with
// errors.As through the Starlark error chain.
type ModuleWithheldError struct {
	Name string
}

// Error returns the error message.
func (e ModuleWithheldError) Error() string {
	return fmt.Sprintf("module %q is withheld and not available to this machine", e.Name)
}

// MaxStepsExceededError marks an execution aborted because it exhausted the
// step budget configured with Machine.SetMaxExecutionSteps. Detect it with
// errors.As through the execution error chain.
type MaxStepsExceededError struct {
	Limit uint64
}

// Error returns the error message.
func (e MaxStepsExceededError) Error() string {
	return fmt.Sprintf("execution exceeded the step limit (%d)", e.Limit)
}
