package starlet_test

import (
	"errors"
	"runtime"
	"starlet"
	"strings"
	"testing"

	"go.starlark.net/starlark"
)

var (
	isOnWindows = runtime.GOOS == `windows`
	isOnLinux   = runtime.GOOS == `linux`
	isOnDarwin  = runtime.GOOS == `darwin`
)

func expectErr(t *testing.T, err error, expected ...string) {
	// preconditions
	if err == nil {
		t.Errorf("unexpected nil error")
		return
	}
	if len(expected) == 0 {
		t.Errorf("no expected string is unexpected")
		return
	}

	// compare one or two strings
	var prefix, suffix string
	if len(expected) == 1 {
		prefix = expected[0]
	} else {
		prefix = expected[0]
		suffix = expected[1]
	}

	// compare
	act := err.Error()
	if prefix != "" && !strings.HasPrefix(act, prefix) {
		t.Errorf(`expected error prefix: %q, got: %q`, prefix, act)
		return
	}
	if suffix != "" && !strings.HasSuffix(act, suffix) {
		t.Errorf(`expected error suffix: %q, got: %q`, suffix, act)
		return
	}
}

// getPrintCompareFunc returns a print function and a compare function.
func getPrintCompareFunc(t *testing.T) (starlet.PrintFunc, func(s string)) {
	var sb strings.Builder
	return func(thread *starlark.Thread, msg string) {
			sb.WriteString(msg)
			sb.WriteString("\n")
		}, func(exp string) {
			act := sb.String()
			if act != exp {
				t.Errorf("expected print(): %q, got: %q", exp, act)
			}
		}
}

// getLogPrintFunc returns a print function that logs to testing.T.
func getLogPrintFunc(t *testing.T) starlet.PrintFunc {
	return func(thread *starlark.Thread, msg string) {
		t.Logf("[⭐ Log] %s", msg)
	}
}

// getErrorModuleLoader returns a module loader that always returns an error, and its name.
func getErrorModuleLoader() (string, starlet.ModuleLoader) {
	return "wrong", func() (starlark.StringDict, error) {
		return nil, errors.New("invalid module loader")
	}
}
