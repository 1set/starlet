// Package goidiomatic provides a Starlark module that defines Go idiomatic functions and values.
package goidiomatic

import (
	"context"
	"errors"
	"time"

	itn "github.com/1set/starlet/lib/internal"
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
		if c, ok := c.(context.Context); ok {
			ctx = c
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
