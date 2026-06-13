package starlet

import (
	"fmt"
	"strings"

	libatom "github.com/1set/starlet/lib/atom"
	libb64 "github.com/1set/starlet/lib/base64"
	libcsv "github.com/1set/starlet/lib/csv"
	libfile "github.com/1set/starlet/lib/file"
	libgoid "github.com/1set/starlet/lib/goidiomatic"
	libhash "github.com/1set/starlet/lib/hashlib"
	libhttp "github.com/1set/starlet/lib/http"
	libjson "github.com/1set/starlet/lib/json"
	liblog "github.com/1set/starlet/lib/log"
	libnet "github.com/1set/starlet/lib/net"
	libpath "github.com/1set/starlet/lib/path"
	librand "github.com/1set/starlet/lib/random"
	libre "github.com/1set/starlet/lib/re"
	libregex "github.com/1set/starlet/lib/regex"
	librt "github.com/1set/starlet/lib/runtime"
	libserial "github.com/1set/starlet/lib/serial"
	libstat "github.com/1set/starlet/lib/stats"
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
	libatom.ModuleName:   libatom.LoadModule,
	libb64.ModuleName:    libb64.LoadModule,
	libcsv.ModuleName:    libcsv.LoadModule,
	libfile.ModuleName:   libfile.LoadModule,
	libhash.ModuleName:   libhash.LoadModule,
	libhttp.ModuleName:   libhttp.LoadModule,
	libnet.ModuleName:    libnet.LoadModule,
	libjson.ModuleName:   libjson.LoadModule,
	liblog.ModuleName:    liblog.LoadModule,
	libpath.ModuleName:   libpath.LoadModule,
	librand.ModuleName:   librand.LoadModule,
	libre.ModuleName:     libre.LoadModule,
	libregex.ModuleName:  libregex.LoadModule,
	librt.ModuleName:     librt.LoadModule,
	libserial.ModuleName: libserial.LoadModule,
	libstr.ModuleName:    libstr.LoadModule,
	libstat.ModuleName:   libstat.LoadModule,
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

// ModuleCapability is a bit set classifying what a builtin module can
// touch on the host. Policy layers use it to derive module sets instead of
// maintaining hand-written name lists that silently rot as modules are
// added.
//
// The bits model host *effects* (filesystem, network, process, log) only.
// They deliberately say nothing about determinism or reproducibility: e.g.
// random is CapPure (it has no host side effects) yet is non-deterministic
// by design. A consumer that needs a "replayable / pure-function" set must
// define it on its own side (excluding entropy/clock sources such as random
// and time) rather than reading it off these capability bits.
type ModuleCapability uint

const (
	// CapLog marks modules that write to the host's logging or console
	// facilities (including stderr helpers).
	CapLog ModuleCapability = 1 << iota
	// CapProcess marks modules that read or mutate process-level state:
	// working directory, host identity, runtime information.
	CapProcess
	// CapFileSystem marks modules that read or write the real filesystem.
	CapFileSystem
	// CapNetwork marks modules that open network connections.
	CapNetwork

	// CapPure is the empty set: pure computation with no host effects.
	CapPure ModuleCapability = 0
)

// Has reports whether c includes every capability in other.
func (c ModuleCapability) Has(other ModuleCapability) bool {
	return c&other == other
}

// Intersects reports whether c shares any capability with other.
func (c ModuleCapability) Intersects(other ModuleCapability) bool {
	return c&other != 0
}

// String returns a "+"-joined list of capability names, or "pure".
func (c ModuleCapability) String() string {
	if c == CapPure {
		return "pure"
	}
	var parts []string
	for _, e := range []struct {
		bit  ModuleCapability
		name string
	}{
		{CapLog, "log"},
		{CapProcess, "process"},
		{CapFileSystem, "filesystem"},
		{CapNetwork, "network"},
	} {
		if c.Has(e.bit) {
			parts = append(parts, e.name)
		}
	}
	return strings.Join(parts, "+")
}

// builtinModuleCapabilities declares the capability set of every builtin
// module. A test pins this map to the module registry, so adding a module
// without classifying it fails the build bar.
var builtinModuleCapabilities = map[string]ModuleCapability{
	"atom":         CapPure,
	"base64":       CapPure,
	"csv":          CapPure,
	"file":         CapFileSystem,
	"go_idiomatic": CapLog | CapProcess, // eprint/pprint write the host's stderr; sleep/exit touch run control
	"hashlib":      CapPure,
	"http":         CapNetwork,
	"json":         CapPure,
	"log":          CapLog,
	"math":         CapPure,
	"net":          CapNetwork,
	"path":         CapFileSystem | CapProcess, // listdir/mkdir touch the FS; chdir/getcwd the process CWD
	"random":       CapPure,
	"re":           CapPure,
	"regex":        CapPure,
	"runtime":      CapProcess, // exposes host identity, directories, process info
	"serial":       CapPure,
	"stats":        CapPure,
	"string":       CapPure,
	"struct":       CapPure,
	"time":         CapPure,
}

// GetBuiltinModuleCapability returns the capability set of a builtin
// module; ok is false for names that are not builtin modules.
func GetBuiltinModuleCapability(name string) (cap ModuleCapability, ok bool) {
	cap, ok = builtinModuleCapabilities[name]
	return
}

// GetBuiltinModuleNamesExcluding returns all builtin module names except
// the given ones. An unknown name in the exclusion list is an error -
// fail-closed, because a typo in a denylist must not silently include the
// module it meant to block.
func GetBuiltinModuleNamesExcluding(excludes ...string) ([]string, error) {
	ex := make(map[string]bool, len(excludes))
	for _, e := range excludes {
		if _, found := allBuiltinModules[e]; !found {
			return nil, fmt.Errorf("unknown builtin module to exclude: %q", e)
		}
		ex[e] = true
	}
	var names []string
	for _, n := range GetAllBuiltinModuleNames() {
		if !ex[n] {
			names = append(names, n)
		}
	}
	return names, nil
}

// GetBuiltinModuleNamesWithoutCapabilities returns the builtin module
// names whose capability sets share nothing with caps - e.g. passing
// CapNetwork|CapFileSystem yields the modules that can touch neither.
func GetBuiltinModuleNamesWithoutCapabilities(caps ModuleCapability) []string {
	var names []string
	for _, n := range GetAllBuiltinModuleNames() {
		if !builtinModuleCapabilities[n].Intersects(caps) {
			names = append(names, n)
		}
	}
	return names
}
