// Package starlet provides powerful extensions and enriched wrappers for Starlark scripting.
//
// Its goal is to enhance the user's scripting experience by combining simplicity and functionality.
// It offers robust, thread-safe types such as Machine, which serves as a wrapper for Starlark runtime environments.
// With Starlet, users can easily manage global variables, load modules, and control the script execution flow.
package starlet

import (
	"fmt"
	"io/fs"
	"sync"

	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
)

// Machine is a thread-safe type that wraps Starlark runtime environments. Machine ensures thread safety by using a sync.RWMutex to control access to the environment's state.
//
// The Machine struct stores the state of the environment, including scripts, modules, and global variables. It provides methods for setting and getting these values, and for running the script. A Machine instance can be configured to preload modules and global variables before running a script or after resetting the environment. It can also lazyload modules right before running the script, the lazyload modules are defined in a list of module loaders and are invoked when the script is run.
//
// The global variables and preload modules can be set before the first run of the script or after resetting the environment. Additionally, extra variables can be set for each run of the script.
//
// Modules are divided into two types: preload and lazyload. Preload modules are loaded before the script is run, while lazyload modules are loaded as and when they are required during the script execution.
//
// The order of precedence for overriding is as follows: global variables, preload modules, and then extra variables before the run, while lazyload modules have the highest precedence during the run.
//
// Setting a print function allows the script to output text to the console or another output stream.
//
// The script to be run is defined by its name and content, and potentially a filesystem (fs.FS) if the script is to be loaded from a file.
//
// The result of each run is cached and written back to the environment, so that it can be used in the next run of the script.
//
// The environment can be reset, allowing the script to be run again with a fresh set of variables and modules.
//
type Machine struct {
	mu sync.RWMutex
	// set variables
	globals      StringAnyMap
	preloadMods  ModuleLoaderList
	lazyloadMods ModuleLoaderMap
	printFunc    PrintFunc
	// source code
	scriptName    string
	scriptContent []byte
	scriptFS      fs.FS
	// runtime core
	runTimes    uint
	loadCache   *cache
	thread      *starlark.Thread
	predeclared starlark.StringDict
}

func (m *Machine) String() string {
	steps := uint64(0)
	if m.thread != nil {
		steps = m.thread.Steps
	}
	return fmt.Sprintf("🌠Machine{run:%d,step:%d,script:%q,len:%d,fs:%v}",
		m.runTimes, steps, m.scriptName, len(m.scriptContent), m.scriptFS)
}

// PrintFunc is a function that tells Starlark how to print messages.
// If nil, the default `fmt.Fprintln(os.Stderr, msg)` will be used instead.
type PrintFunc func(thread *starlark.Thread, msg string)

// LoadFunc is a function that tells Starlark how to find and load other scripts
// using the load() function. If you don't use load() in your scripts, you can pass in nil.
type LoadFunc func(thread *starlark.Thread, module string) (starlark.StringDict, error)

// NewDefault creates a new Starlark runtime environment.
func NewDefault() *Machine {
	return &Machine{}
}

// NewWithGlobals creates a new Starlark runtime environment with given global variables.
func NewWithGlobals(globals StringAnyMap) *Machine {
	return &Machine{
		globals: globals,
	}
}

// NewWithLoaders creates a new Starlark runtime environment with given global variables and preload & lazyload module loaders.
func NewWithLoaders(globals StringAnyMap, preload ModuleLoaderList, lazyload ModuleLoaderMap) *Machine {
	return &Machine{
		globals:      globals,
		preloadMods:  preload,
		lazyloadMods: lazyload,
	}
}

// NewWithBuiltins creates a new Starlark runtime environment with given global variables and all preload & lazyload built-in modules.
func NewWithBuiltins(globals StringAnyMap, additionalPreload ModuleLoaderList, additionalLazyload ModuleLoaderMap) *Machine {
	pre := append(GetAllBuiltinModules(), additionalPreload...)
	lazy := GetBuiltinModuleMap()
	lazy.Merge(additionalLazyload)
	return &Machine{
		globals:      globals,
		preloadMods:  pre,
		lazyloadMods: lazy,
	}
}

// NewWithNames creates a new Starlark runtime environment with given global variables, preload and lazyload module names.
// The modules should be built-in modules, and it panics if any of the given modules fails to load.
func NewWithNames(globals StringAnyMap, preloads []string, lazyloads []string) *Machine {
	pre, err := MakeBuiltinModuleLoaderList(preloads...)
	if err != nil {
		panic(err)
	}
	lazy, err := MakeBuiltinModuleLoaderMap(lazyloads...)
	if err != nil {
		panic(err)
	}
	return &Machine{
		globals:      globals,
		preloadMods:  pre,
		lazyloadMods: lazy,
	}
}

// SetGlobals sets global variables in the Starlark runtime environment.
// These variables only take effect before the first run or after a reset.
func (m *Machine) SetGlobals(globals StringAnyMap) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.globals = globals
}

// AddGlobals adds the globals of the Starlark runtime environment.
// These variables only take effect before the first run or after a reset.
func (m *Machine) AddGlobals(globals StringAnyMap) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.globals == nil {
		m.globals = make(StringAnyMap)
	}
	for k, v := range globals {
		m.globals[k] = v
	}
}

// GetGlobals gets the globals of the Starlark runtime environment.
func (m *Machine) GetGlobals() StringAnyMap {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.globals.Clone()
}

// SetPreloadModules sets the preload modules of the Starlark runtime environment.
// These modules only take effect before the first run or after a reset.
func (m *Machine) SetPreloadModules(mods ModuleLoaderList) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.preloadMods = mods
}

// GetPreloadModules gets the preload modules of the Starlark runtime environment.
func (m *Machine) GetPreloadModules() ModuleLoaderList {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.preloadMods.Clone()
}

// AddPreloadModules adds the preload modules of the Starlark runtime environment.
// These modules only take effect before the first run or after a reset.
func (m *Machine) AddPreloadModules(mods ModuleLoaderList) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.preloadMods = append(m.preloadMods, mods...)
}

// SetLazyloadModules sets the modules allowed to be loaded later of the Starlark runtime environment.
func (m *Machine) SetLazyloadModules(mods ModuleLoaderMap) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lazyloadMods = mods
}

// GetLazyloadModules gets the modules allowed to be loaded later of the Starlark runtime environment.
func (m *Machine) GetLazyloadModules() ModuleLoaderMap {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.lazyloadMods.Clone()
}

// AddLazyloadModules adds the modules allowed to be loaded later of the Starlark runtime environment.
func (m *Machine) AddLazyloadModules(mods ModuleLoaderMap) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.lazyloadMods == nil {
		m.lazyloadMods = make(ModuleLoaderMap)
	}
	m.lazyloadMods.Merge(mods)
}

// SetPrintFunc sets the print function of the Starlark runtime environment.
func (m *Machine) SetPrintFunc(printFunc PrintFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.printFunc = printFunc
}

// SetScript sets the script related things of the Starlark runtime environment.
func (m *Machine) SetScript(name string, content []byte, fileSys fs.FS) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.scriptName = name
	m.scriptContent = content
	m.scriptFS = fileSys
}

// Export returns the current variables of the Starlark runtime environment.
func (m *Machine) Export() StringAnyMap {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return convert.FromStringDict(m.predeclared)
}

// StringAnyMap type is a map of string to interface{} and is used to store global variables like StringDict of Starlark, but not a Starlark type.
type StringAnyMap map[string]interface{}

// Clone returns a copy of the data store.
func (d StringAnyMap) Clone() StringAnyMap {
	clone := make(StringAnyMap)
	for k, v := range d {
		clone[k] = v
	}
	return clone
}

// Merge merges the given data store into the current data store. It does nothing if the current data store is nil.
func (d StringAnyMap) Merge(other StringAnyMap) {
	if d == nil {
		return
	}
	for k, v := range other {
		d[k] = v
	}
}

// MergeDict merges the given string dict into the current data store.
func (d StringAnyMap) MergeDict(other starlark.StringDict) {
	if d == nil {
		return
	}
	for k, v := range other {
		d[k] = v
	}
}
