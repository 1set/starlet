// Package goidiomatic provides a Starlark module that defines Go idiomatic functions and values.
package goidiomatic

import (
	"context"
	"errors"
	"time"

	itn "github.com/1set/starlet/lib/internal"
	"go.starlark.net/starlark"
)

// ModuleName defines the expected name for this Module when used in Starlark's load() function, eg: load('go_idiomatic', 'nil')
const ModuleName = "go_idiomatic"

// LoadModule loads the Go idiomatic module.
func LoadModule() (starlark.StringDict, error) {
	return starlark.StringDict{
		"true":  starlark.True,
		"false": starlark.False,
		"nil":   starlark.None,
		"sleep": starlark.NewBuiltin("sleep", sleep),
		"exit":  starlark.NewBuiltin("exit", exit),
		"quit":  starlark.NewBuiltin("quit", exit), // alias for exit
	}, nil
}

// for convenience
var (
	none = starlark.None
)

// sleep sleeps for the given number of seconds.
func sleep(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// get the duration
	var sec itn.FloatOrInt
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "secs", &sec); err != nil {
		return none, err
	}
	if sec < 0 {
		return none, errors.New("secs must be non-negative")
	}
	dur := time.Duration(float64(sec) * float64(time.Second))
	// get the context
	ctx := context.TODO()
	if c := thread.Local("context"); c != nil {
		if co, ok := c.(context.Context); ok {
			ctx = co
		}
	}
	// sleep
	t := time.NewTimer(dur)
	defer t.Stop()
	select {
	case <-t.C:
		return none, nil
	case <-ctx.Done():
		return none, ctx.Err()
	}
}

// exit exits the program with the given exit code.
func exit(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var code uint8
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "code?", &code); err != nil {
		return none, err
	}
	thread.SetLocal("exit_code", code)
	return none, ErrSystemExit
}

var (
	// ErrSystemExit is returned by exit() to indicate the program should exit.
	ErrSystemExit = errors.New(`starlet runtime system exit (Use Ctrl-D in REPL to exit)`)
)
