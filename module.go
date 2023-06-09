package starlet

import (
	"fmt"
	"sort"

	sjson "go.starlark.net/lib/json"
	smath "go.starlark.net/lib/math"
	stime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var allBuiltinModules = ModuleLoaderMap{
	"go_idiomatic": func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"true":  starlark.True,
			"false": starlark.False,
			"nil":   starlark.None,
			//"exit":  starlark.NewBuiltin("exit", exit),
		}, nil
	},
	"struct": func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
		}, nil
	},
	"time": func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"time": stime.Module,
		}, nil
	},
	"math": func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"math": smath.Module,
		}, nil
	},
	"json": func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"json": sjson.Module,
		}, nil
	},
}

// ListBuiltinModules returns a list of all builtin modules.
func ListBuiltinModules() []string {
	modules := make([]string, 0, len(allBuiltinModules))
	for k := range allBuiltinModules {
		modules = append(modules, k)
	}
	sort.Strings(modules)
	return modules
}

// GetBuiltinModule returns the builtin module with the given name.
func GetBuiltinModule(name string) ModuleLoader {
	return allBuiltinModules[name]
}

// ModuleLoader is a function that loads a Starlark module and returns the module's string dict.
type ModuleLoader func() (starlark.StringDict, error)

// NamedModuleLoader is a function that loads a Starlark module with the given name and returns the module's string dict.
// If the module is not found, it returns nil as the first and second return value.
type NamedModuleLoader func(string) (starlark.StringDict, error)

// ModuleLoaderList is a list of Starlark module loaders, usually used to load a list of modules in order.
type ModuleLoaderList []ModuleLoader

// Clone returns a copy of the list.
func (l ModuleLoaderList) Clone() []ModuleLoader {
	return append([]ModuleLoader{}, l...)
}

// LoadAll loads all modules in the list into the given StringDict.
// It returns an error as second return value if any module fails to load.
func (l ModuleLoaderList) LoadAll(d starlark.StringDict) error {
	if d == nil {
		return fmt.Errorf("starlet: cannot load modules into nil dict")
	}
	for _, ld := range l {
		if ld == nil {
			return fmt.Errorf("starlet: nil module loader")
		}
		m, err := ld()
		if err != nil {
			return fmt.Errorf("starlet: failed to load module: %w", err)
		}
		if m != nil {
			for k, v := range m {
				d[k] = v
			}
		}
	}
	return nil
}

// ModuleLoaderMap is a map of Starlark module loaders, usually used to load a map of modules by name.
type ModuleLoaderMap map[string]ModuleLoader

// Clone returns a copy of the map.
func (m ModuleLoaderMap) Clone() map[string]ModuleLoader {
	clone := make(map[string]ModuleLoader, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

// GetLazyLoader returns a lazy loader that loads the module with the given name.
// It returns an error as second return value if the module is found but fails to load.
// Otherwise, the first return value is nil if the module is not found.
func (m ModuleLoaderMap) GetLazyLoader() NamedModuleLoader {
	return func(s string) (starlark.StringDict, error) {
		if m == nil {
			return nil, nil
		}
		ld, ok := m[s]
		if !ok {
			return nil, nil
		} else if ld == nil {
			return nil, fmt.Errorf("starlet: nil module loader %q", s)
		}
		return ld()
	}
}

// MakeBuiltinModuleLoaderList creates a list of module loaders from a list of module names.
// It returns an error as second return value if any module is not found.
func MakeBuiltinModuleLoaderList(names []string) (ModuleLoaderList, error) {
	ld := make(ModuleLoaderList, len(names))
	for i, name := range names {
		ld[i] = allBuiltinModules[name]
		if ld[i] == nil {
			return ld, fmt.Errorf("starlet: module %q: %w", name, ErrModuleNotFound)
		}
	}
	return ld, nil
}

// MakeBuiltinModuleLoaderMap creates a map of module loaders from a list of module names.
// It returns an error as second return value if any module is not found.
func MakeBuiltinModuleLoaderMap(names []string) (ModuleLoaderMap, error) {
	ld := make(ModuleLoaderMap, len(names))
	for _, name := range names {
		ld[name] = allBuiltinModules[name]
		if ld[name] == nil {
			return ld, fmt.Errorf("starlet: module %q: %w", name, ErrModuleNotFound)
		}
	}
	return ld, nil
}
