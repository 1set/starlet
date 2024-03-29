package internal

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/starlarktest"
	"go.starlark.net/syntax"
)

// ModuleLoadFunc is a function that loads a Starlark module and returns the module's string dict.
type ModuleLoadFunc func() (starlark.StringDict, error)

// ThreadLoadFunc is a function that loads a Starlark module by name, usually used by the Starlark thread.
type ThreadLoadFunc func(thread *starlark.Thread, module string) (starlark.StringDict, error)

var initTestOnce sync.Once

// NewAssertLoader creates a Starlark thread loader that loads a module by name or asserts.star for testing.
func NewAssertLoader(moduleName string, loader ModuleLoadFunc) ThreadLoadFunc {
	initTestOnce.Do(func() {
		starlarktest.DataFile = func(pkgdir, filename string) string {
			_, currFileName, _, ok := runtime.Caller(1)
			if !ok {
				return ""
			}
			return filepath.Join(filepath.Dir(currFileName), filename)
		}
	})
	// for assert loader
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
		case "module.star":
			return starlark.StringDict{
				"module": starlark.NewBuiltin("module", starlarkstruct.MakeModule),
			}, nil
		case "assert.star":
			return starlarktest.LoadAssertModule()
		case "freeze.star":
			return starlark.StringDict{
				"freeze": starlark.NewBuiltin("freeze", freezeValue),
			}, nil
		}

		return nil, fmt.Errorf("invalid module")
	}
}

// ExecModuleWithErrorTest executes a Starlark script with a module loader and compares the error with the expected error.
func ExecModuleWithErrorTest(t *testing.T, name string, loader ModuleLoadFunc, script string, wantErr string, predecl starlark.StringDict) (starlark.StringDict, error) {
	thread := &starlark.Thread{Load: NewAssertLoader(name, loader), Print: func(_ *starlark.Thread, msg string) { t.Log("※", msg) }}
	starlarktest.SetReporter(thread, t)
	header := `load('assert.star', 'assert')`
	opts := syntax.FileOptions{
		Set: true,
	}
	out, err := starlark.ExecFileOptions(&opts, thread, name+"_test.star", []byte(header+"\n"+script), predecl)
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

func freezeValue(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var v starlark.Value
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "v", &v); err != nil {
		return nil, err
	}
	v.Freeze()
	return v, nil
}
