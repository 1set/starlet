package starlet_test

import (
	"starlet"
	"strings"
	"testing"

	"go.starlark.net/starlark"
)

func expectErr(t *testing.T, err error, expected ...string) {
	// preconditions
	if err == nil {
		t.Errorf("unexpected nil error")
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
		t.Errorf(`expected error prefix %q, got: %v`, expected, err)
	}
	if suffix != "" && !strings.HasSuffix(act, suffix) {
		t.Errorf(`expected error suffix %q, got: %v`, suffix, err)
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
				t.Errorf("expected print %q, got %q", exp, act)
			}
		}
}
