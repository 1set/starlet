// Package goidiomatic provides a Starlark module that defines Go idiomatic functions and values.
package goidiomatic

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	itn "github.com/1set/starlet/lib/internal"
	"go.starlark.net/starlark"
)

// ModuleName defines the expected name for this Module when used in Starlark's load() function, eg: load('go_idiomatic', 'nil')
const ModuleName = "go_idiomatic"

// LoadModule loads the Go idiomatic module.
func LoadModule() (starlark.StringDict, error) {
	return starlark.StringDict{
		"true":   starlark.True,
		"false":  starlark.False,
		"nil":    starlark.None,
		"length": starlark.NewBuiltin("length", length),
		"sum":    starlark.NewBuiltin("sum", sum),
		"oct":    starlark.NewBuiltin("oct", oct),
		"hex":    starlark.NewBuiltin("hex", hex),
		"sleep":  starlark.NewBuiltin("sleep", sleep),
		"exit":   starlark.NewBuiltin("exit", exit),
		"quit":   starlark.NewBuiltin("quit", exit), // alias for exit
	}, nil
}

// for convenience
var (
	none = starlark.None
)

// length returns the length of the given value.
func length(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if l := len(args); l != 1 {
		return none, fmt.Errorf(`length() takes exactly one argument (%d given)`, l)
	}

	switch r := args[0]; v := r.(type) {
	case starlark.String:
		return starlark.MakeInt(utf8.RuneCountInString(v.GoString())), nil
	case starlark.Bytes:
		return starlark.MakeInt(len(v)), nil
	default:
		if sv, ok := v.(starlark.Sequence); ok {
			return starlark.MakeInt(sv.Len()), nil
		}
		return none, fmt.Errorf(`object of type '%s' has no length()`, v.Type())
	}
}

// sum returns the sum of the given values.
func sum(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		lst   starlark.Iterable
		start starlark.Value
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "iterable", &lst, "start?", &start); err != nil {
		return none, err
	}

	// start with the given start value
	total := itn.NewNumericValue()
	if err := total.Add(start); err != nil {
		return none, err
	}

	// loop through the list
	iter := lst.Iterate()
	defer iter.Done()

	var x starlark.Value
	for iter.Next(&x) {
		if err := total.Add(x); err != nil {
			return none, err
		}
	}

	// return the result
	return total.Value(), nil
}

// hex returns the hexadecimal representation of the given value.
func hex(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Int
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return none, err
	}
	return convertStarlarkNumber(x, 16, "0x")
}

// oct returns the octal representation of the given value.
func oct(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Int
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &x); err != nil {
		return none, err
	}
	return convertStarlarkNumber(x, 8, "0o")
}

// convertStarlarkNumber converts the given starlark.Int number to the given base string.
func convertStarlarkNumber(x starlark.Int, base int, fmtPre string) (starlark.Value, error) {
	var (
		s    string
		n    = x.BigInt()
		sign = n.Sign()
	)
	if sign == 0 {
		s = fmtPre + "0"
	} else {
		signPre := ""
		if sign < 0 {
			signPre = "-"
			n.Neg(n)
		}
		s = signPre + fmtPre + n.Text(base)
	}
	return starlark.String(s), nil
}

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
