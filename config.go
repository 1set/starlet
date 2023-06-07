package starlet

import (
	"go.starlark.net/resolve"
)

// EnableRecursionSupport enables recursion support in Starlark.
func EnableRecursionSupport() {
	resolve.AllowRecursion = true
}

// DisableRecursionSupport disables recursion support in Starlark.
func DisableRecursionSupport() {
	resolve.AllowRecursion = false
}

// EnableGlobalReassign enables global reassignment in Starlark.
func EnableGlobalReassign() {
	resolve.AllowGlobalReassign = true
}

// DisableGlobalReassign disables global reassignment in Starlark.
func DisableGlobalReassign() {
	resolve.AllowGlobalReassign = false
}

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
