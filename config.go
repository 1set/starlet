package starlet

import (
	libb64 "github.com/1set/starlet/lib/base64"
	libfile "github.com/1set/starlet/lib/file"
	libgoid "github.com/1set/starlet/lib/goidiomatic"
	libhash "github.com/1set/starlet/lib/hashlib"
	libhttp "github.com/1set/starlet/lib/http"
	libjson "github.com/1set/starlet/lib/json"
	liblog "github.com/1set/starlet/lib/log"
	librand "github.com/1set/starlet/lib/random"
	libre "github.com/1set/starlet/lib/re"
	librt "github.com/1set/starlet/lib/runtime"
	libstr "github.com/1set/starlet/lib/string"
	stdmath "go.starlark.net/lib/math"
	stdtime "go.starlark.net/lib/time"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	stdstruct "go.starlark.net/starlarkstruct"
)

var allBuiltinModules = ModuleLoaderMap{
	libgoid.ModuleName: func() (starlark.StringDict, error) {
		return libgoid.LoadModule()
	},
	"math": func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"math": stdmath.Module,
		}, nil
	},
	"struct": func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"struct": starlark.NewBuiltin("struct", stdstruct.Make),
		}, nil
	},
	"time": func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"time": stdtime.Module,
		}, nil
	},
	// add third-party modules
	libhash.ModuleName: libhash.LoadModule,
	libhttp.ModuleName: libhttp.LoadModule,
	libre.ModuleName:   libre.LoadModule,
	libb64.ModuleName:  libb64.LoadModule,
	librand.ModuleName: librand.LoadModule,
	libjson.ModuleName: libjson.LoadModule,
	libstr.ModuleName:  libstr.LoadModule,
	librt.ModuleName:   librt.LoadModule,
	liblog.ModuleName:  liblog.LoadModule,
	libfile.ModuleName: libfile.LoadModule,
}

// GetAllBuiltinModuleNames returns a list of all builtin module names.
func GetAllBuiltinModuleNames() []string {
	return allBuiltinModules.Keys()
}

// GetAllBuiltinModules returns a list of all builtin modules.
func GetAllBuiltinModules() ModuleLoaderList {
	return allBuiltinModules.Values()
}

// GetBuiltinModuleMap returns a map of all builtin modules.
func GetBuiltinModuleMap() ModuleLoaderMap {
	return allBuiltinModules.Clone()
}

// GetBuiltinModule returns the builtin module with the given name.
func GetBuiltinModule(name string) ModuleLoader {
	return allBuiltinModules[name]
}

// EnableRecursionSupport enables recursion support in Starlark environments for loading modules.
func EnableRecursionSupport() {
	resolve.AllowRecursion = true
}

// DisableRecursionSupport disables recursion support in Starlark environments for loading modules.
func DisableRecursionSupport() {
	resolve.AllowRecursion = false
}

// EnableGlobalReassign enables global reassignment in Starlark environments for loading modules.
func EnableGlobalReassign() {
	resolve.AllowGlobalReassign = true
}

// DisableGlobalReassign disables global reassignment in Starlark environments for loading modules.
func DisableGlobalReassign() {
	resolve.AllowGlobalReassign = false
}
