// Package goidiomatic provides a Starlark module that defines Go idiomatic functions and values.
package goidiomatic

import (
	"context"
	"errors"
	"time"

	"go.starlark.net/starlark"
)

// ModuleName defines the expected name for this Module when used in starlark's load() function, eg: load('go_idiomatic', 'nil')
const ModuleName = "go_idiomatic"

// LoadModule loads the Go idiomatic module.
func LoadModule() (starlark.StringDict, error) {
	return starlark.StringDict{
		"true":  starlark.True,
		"false": starlark.False,
		"nil":   starlark.None,
		"sleep": starlark.NewBuiltin("sleep", sleep),
	}, nil
}

// sleep sleeps for the given number of seconds.
func sleep(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// get the duration
	var sec float64
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "secs", &sec); err != nil {
		return nil, err
	}
	if sec < 0 {
		return nil, errors.New("secs must be non-negative")
	}
	dur := time.Duration(sec) * time.Second
	// get the context
	ctx := context.TODO()
	if c := thread.Local("context"); c != nil {
		if c, ok := c.(context.Context); ok {
			ctx = c
		}
	}
	// sleep
	t := time.NewTimer(dur)
	defer t.Stop()
	select {
	case <-t.C:
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
