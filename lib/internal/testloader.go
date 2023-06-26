package internal

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/starlarktest"
)

// ModuleLoadFunc is a function that loads a Starlark module and returns the module's string dict.
type ModuleLoadFunc func() (starlark.StringDict, error)

// ThreadLoadFunc is a function that loads a Starlark module by name, usually used by the Starlark thread.
type ThreadLoadFunc func(thread *starlark.Thread, module string) (starlark.StringDict, error)

// NewAssertLoader creates a Starlark thread loader that loads a module by name or asserts.star for testing.
func NewAssertLoader(moduleName string, loader ModuleLoadFunc) ThreadLoadFunc {
	return func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
		switch module {
		case moduleName:
			d, err := loader()
			if err != nil {
				// failed to load
				return nil, err
			}
			// extract all members of module from dict like `{name: module}`
			if len(d) == 1 {
				m, found := d[moduleName]
				if found {
					if md, ok := m.(*starlarkstruct.Module); ok && md != nil {
						return md.Members, nil
					}
				}
			}
			return d, nil
		case "assert.star":
			starlarktest.DataFile = func(pkgdir, filename string) string {
				_, currFileName, _, ok := runtime.Caller(1)
				if !ok {
					return ""
				}
				return filepath.Join(filepath.Dir(currFileName), filename)
			}
			return starlarktest.LoadAssertModule()
		}

		return nil, fmt.Errorf("invalid module")
	}
}

// ExecModuleWithErrorTest executes a Starlark script with a module loader and compares the error with the expected error.
func ExecModuleWithErrorTest(t *testing.T, name string, loader ModuleLoadFunc, script string, wantErr error) (starlark.StringDict, error) {
	thread := &starlark.Thread{Load: NewAssertLoader(name, loader)}
	starlarktest.SetReporter(thread, t)
	header := `load('assert.star', 'assert')`
	out, err := starlark.ExecFile(thread, name+"_test.star", []byte(header+"\n"+script), nil)
	if err != nil {
		if wantErr == nil {
			if ee, ok := err.(*starlark.EvalError); ok {
				t.Errorf("got unexpected starlark error: '%v'", ee.Backtrace())
			} else {
				t.Errorf("got unexpected error: '%v'", err)
			}
		}
		if wantErr != nil && err.Error() != wantErr.Error() {
			t.Errorf("got mismatched error: '%v', want: '%v'", err, wantErr)
		}
	}
	return out, err
}
