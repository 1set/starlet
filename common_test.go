package starlet_test

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"reflect"
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

func expectEqualStringAnyMap(t *testing.T, act map[string]interface{}, exp map[string]interface{}) bool {
	if len(act) != len(exp) {
		t.Errorf("expected map length: %d, got: %d", len(exp), len(act))
		return false
	}
	for k, v := range exp {
		actV, ok := act[k]
		if !ok {
			t.Errorf("expected key: %q, got: %v", k, actV)
			return false
		}
		if !reflect.DeepEqual(v, actV) {
			t.Errorf("expected value: %v, got: %v", v, actV)
			return false
		}
	}
	return true
}

func getFuncAddr(i interface{}) uintptr {
	return reflect.ValueOf(i).Pointer()
}

func expectEqualModuleList(t *testing.T, act starlet.ModuleLoaderList, exp starlet.ModuleLoaderList) bool {
	if len(act) != len(exp) {
		t.Errorf("expected module list length: %d, got: %d", len(exp), len(act))
		return false
	}
	for i := range exp {
		e := getFuncAddr(exp[i])
		a := getFuncAddr(act[i])
		if e != a {
			t.Errorf("expected module: %v, got: %v", e, a)
			return false
		}
	}
	return true
}

func expectEqualModuleMap(t *testing.T, act starlet.ModuleLoaderMap, exp starlet.ModuleLoaderMap) bool {
	if len(act) != len(exp) {
		t.Errorf("expected module map length: %d, got: %d", len(exp), len(act))
		return false
	}
	for k, v := range exp {
		actV, ok := act[k]
		if !ok {
			t.Errorf("expected key: %q, got: %p", k, actV)
			return false
		}
		e := getFuncAddr(v)
		a := getFuncAddr(actV)
		if e != a {
			t.Errorf("expected module: %v, got: %v", e, a)
			return false
		}
	}
	return true
}

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

// expectSameDuration checks if the actual duration is within 15% of the expected duration.
func expectSameDuration(t *testing.T, act, exp time.Duration) bool {
	r := float64(act) / float64(exp)
	d := math.Abs(r - 1)
	same := d <= 0.15
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

// MemFS is an in-memory filesystem.
type MemFS map[string]string

func (m MemFS) Open(name string) (fs.File, error) {
	if data, ok := m[name]; ok {
		return &MemFile{data: data, name: name}, nil
	}
	return nil, fs.ErrNotExist
}

// MemFile is an in-memory file.
type MemFile struct {
	data string
	name string
	pos  int
}

func (f *MemFile) Stat() (fs.FileInfo, error) {
	return &MemFileInfo{
		name: f.name,
		size: len(f.data),
	}, nil
}

func (f *MemFile) Read(p []byte) (n int, err error) {
	if f.pos >= len(f.data) {
		return 0, io.EOF // Indicate end of file
	}

	n = copy(p, f.data[f.pos:])
	f.pos += n
	return n, nil
}

func (f *MemFile) Close() error {
	return nil
}

// MemFileInfo is an in-memory file info.
type MemFileInfo struct {
	name string
	size int
}

func (fi *MemFileInfo) Name() string {
	return fi.name
}

func (fi *MemFileInfo) Size() int64 {
	return int64(fi.size)
}

func (fi *MemFileInfo) Mode() fs.FileMode {
	return 0444 // read-only
}

func (fi *MemFileInfo) ModTime() time.Time {
	return time.Time{} // zero time
}

func (fi *MemFileInfo) IsDir() bool {
	return false
}

func (fi *MemFileInfo) Sys() interface{} {
	return nil
}
