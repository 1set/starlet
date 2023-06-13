package starlet_test

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/1set/starlet"
	"github.com/1set/starlight/convert"
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
		t.Logf("[⭐Log] %s", msg)
	}
}

// getErrorModuleLoader returns a module loader that always returns an error, and its name.
func getErrorModuleLoader() (string, starlet.ModuleLoader) {
	return "wrong", func() (starlark.StringDict, error) {
		return nil, errors.New("invalid module loader")
	}
}

// getAppleModuleLoader returns a module loader that always returns a simple module, and its name.
func getAppleModuleLoader() (string, starlet.ModuleLoader) {
	return "mock_apple", func() (starlark.StringDict, error) {
		const val = 10
		return starlark.StringDict{
			"apple":  starlark.String("🍎"),
			"number": starlark.MakeInt(val),
			"process": convert.MakeStarFn("process", func(x int) int {
				return x * val
			}),
		}, nil
	}
}

// getBlueberryModuleLoader returns a module loader that always returns a simple module, and its name.
func getBlueberryModuleLoader() (string, starlet.ModuleLoader) {
	return "mock_blueberry", func() (starlark.StringDict, error) {
		const val = 20
		return starlark.StringDict{
			"blueberry": starlark.String("🫐"),
			"number":    starlark.MakeInt(val),
			"process": convert.MakeStarFn("process", func(x int) int {
				return x * val
			}),
		}, nil
	}
}

// getCoconutModuleLoader returns a module loader that always returns a simple module, and its name.
func getCoconutModuleLoader() (string, starlet.ModuleLoader) {
	return "mock_coconut", func() (starlark.StringDict, error) {
		const val = 40
		return starlark.StringDict{
			"coconut": starlark.String("🥥"),
			"number":  starlark.MakeInt(val),
			"process": convert.MakeStarFn("process", func(x int) int {
				return x * val
			}),
		}, nil
	}
}

// expectSameDuration checks if the given duration is near to the expect duration.
func expectSameDuration(t *testing.T, act, exp time.Duration) bool {
	r := float64(act) / float64(exp)
	d := math.Abs(r - 1)
	same := d < 0.05
	if !same {
		t.Errorf("expected same duration, got diff: actual = %v, expected = %v", act, exp)
	}
	return same
}

// errorReader is a reader that reads that fails at the given point.
type errorReader struct {
	data   []byte
	count  int
	target int
}

func newErrorReader(data string, target int) *errorReader {
	return &errorReader{data: []byte(data), target: target}
}

// Read implements the io.Reader interface.
func (r *errorReader) Read(p []byte) (n int, err error) {
	r.count++
	if r.count == r.target {
		return 0, fmt.Errorf("desired error at %d", r.target)
	}
	copy(p, r.data)
	return len(r.data), nil
}
