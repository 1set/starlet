package starlet

import (
	sjson "go.starlark.net/lib/json"
	smath "go.starlark.net/lib/math"
	stime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName represents a Starlark module name or collection of functions and values.
type ModuleName string

const (
	// ModuleGoIdiomatic is a collection of Go idiomatic functions and values. e.g. true, false, nil, exit, etc.
	ModuleGoIdiomatic = ModuleName("go_idiomatic")

	// ModuleStruct is the official Starlark struct and module.
	ModuleStruct = ModuleName("struct")

	// ModuleTime is the official Starlark time module.
	ModuleTime = ModuleName("time")

	// ModuleMath is the official Starlark math module.
	ModuleMath = ModuleName("math")

	// ModuleJSON is the official Starlark JSON module.
	ModuleJSON = ModuleName("json")
)

// ModuleNameList is a list of Starlark module names.
type ModuleNameList []ModuleName

// Clone returns a copy of the list.
func (l ModuleNameList) Clone() []ModuleName {
	return append([]ModuleName{}, l...)
}

// loadModuleByName loads a Starlark module with the given name.
// It returns an error as second return value if the module is found but fails to load.
// Otherwise, the first return value is nil if the module is not found.
func loadModuleByName(name ModuleName) (starlark.StringDict, error) {
	switch name {
	case ModuleGoIdiomatic:
		return starlark.StringDict{
			"true":  starlark.True,
			"false": starlark.False,
			"nil":   starlark.None,
			//"exit":  starlark.NewBuiltin("exit", exit),
		}, nil
	case ModuleStruct:
		return starlark.StringDict{
			"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
		}, nil
	case ModuleTime:
		return starlark.StringDict{
			"time": stime.Module,
		}, nil
	case ModuleMath:
		return starlark.StringDict{
			"math": smath.Module,
		}, nil
	case ModuleJSON:
		return starlark.StringDict{
			"json": sjson.Module,
		}, nil
	}
	return nil, nil
}

// Now my refactor begins 2023-06-08 19:10:38 CST

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
		}
		return ld()
	}
}

// CreateBuiltinModuleLoaderList creates a list of module loaders from a list of module names.
// It returns an error as second return value if any module is not found.
func CreateBuiltinModuleLoaderList(names []string) (ModuleLoaderList, error) {
	ld := make(ModuleLoaderList, len(names))
	for i, name := range names {
		ld[i] = allBuiltinModules[name]
		if ld[i] == nil {
			return ld, ErrModuleNotFound
		}
	}
	return ld, nil
}

// CreateBuiltinModuleLoaderMap creates a map of module loaders from a list of module names.
// It returns an error as second return value if any module is not found.
func CreateBuiltinModuleLoaderMap(names []string) (ModuleLoaderMap, error) {
	ld := make(ModuleLoaderMap, len(names))
	for _, name := range names {
		ld[name] = allBuiltinModules[name]
		if ld[name] == nil {
			return ld, ErrModuleNotFound
		}
	}
	return ld, nil
}
