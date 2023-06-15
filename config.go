package starlet

import (
	"sort"

	"github.com/1set/starlet/lib/goidiomatic"
	sjson "go.starlark.net/lib/json"
	smath "go.starlark.net/lib/math"
	stime "go.starlark.net/lib/time"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var allBuiltinModules = ModuleLoaderMap{
	goidiomatic.ModuleName: func() (starlark.StringDict, error) {
		return goidiomatic.LoadModule()
	},
	"json": func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"json": sjson.Module,
		}, nil
	},
	"math": func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"math": smath.Module,
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

// EnableRecursionSupport enables recursion support in all Starlark environments.
func EnableRecursionSupport() {
	resolve.AllowRecursion = true
}

// DisableRecursionSupport disables recursion support in all Starlark environments.
func DisableRecursionSupport() {
	resolve.AllowRecursion = false
}

// EnableGlobalReassign enables global reassignment in all Starlark environments.
func EnableGlobalReassign() {
	resolve.AllowGlobalReassign = true
}

// DisableGlobalReassign disables global reassignment in all Starlark environments.
func DisableGlobalReassign() {
	resolve.AllowGlobalReassign = false
}
