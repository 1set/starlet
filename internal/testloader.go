package internal

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
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
			if loader == nil {
				return nil, fmt.Errorf("nil module")
			}
			d, err := loader()
			if err != nil {
				// failed to load
				return nil, err
			}
			// Aligned with starlet/module.go: GetLazyLoader() function
			// extract all members of module from dict like `{name: module}` or `{name: struct}`
			if len(d) == 1 {
				m, found := d[moduleName]
				if found {
					if mm, ok := m.(*starlarkstruct.Module); ok && mm != nil {
						return mm.Members, nil
					} else if sm, ok := m.(*starlarkstruct.Struct); ok && sm != nil {
						sd := make(starlark.StringDict)
						sm.ToStringDict(sd)
						return sd, nil
					}
				}
			}
			return d, nil
		case "struct.star":
			return starlark.StringDict{
				"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
			}, nil
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
func ExecModuleWithErrorTest(t *testing.T, name string, loader ModuleLoadFunc, script string, wantErr string) (starlark.StringDict, error) {
	thread := &starlark.Thread{Load: NewAssertLoader(name, loader), Print: func(_ *starlark.Thread, msg string) { t.Log("※", msg) }}
	starlarktest.SetReporter(thread, t)
	header := `load('assert.star', 'assert')`
	out, err := starlark.ExecFile(thread, name+"_test.star", []byte(header+"\n"+script), nil)
	if err != nil {
		if wantErr == "" {
			if ee, ok := err.(*starlark.EvalError); ok {
				t.Errorf("got unexpected starlark error: '%v'", ee.Backtrace())
			} else {
				t.Errorf("got unexpected error: '%v'", err)
			}
		} else if wantErr != "" && !strings.Contains(err.Error(), wantErr) {
			t.Errorf("got mismatched error: '%v', want: '%v'", err, wantErr)
		}
	}
	return out, err
}
