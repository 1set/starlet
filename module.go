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
// The second return value is nil if the module is not found, or an error if the module is found but fails to load.
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
