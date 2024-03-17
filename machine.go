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
type Machine struct {
	mu sync.RWMutex
	// set variables
	globals             StringAnyMap
	preloadMods         ModuleLoaderList
	lazyloadMods        ModuleLoaderMap
	printFunc           PrintFunc
	allowGlobalReassign bool
	allowRecursion      bool
	enableInConv        bool
	enableOutConv       bool
	customTag           string
	// source code
	scriptName    string
	scriptContent []byte
	scriptFS      fs.FS
	// runtime core
	progCache   ByteCache
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
	return &Machine{enableInConv: true, enableOutConv: true}
}

// NewWithGlobals creates a new Starlark runtime environment with given global variables.
func NewWithGlobals(globals StringAnyMap) *Machine {
	return &Machine{
		enableInConv:  true,
		enableOutConv: true,
		globals:       globals,
	}
}

// NewWithLoaders creates a new Starlark runtime environment with given global variables and preload & lazyload module loaders.
func NewWithLoaders(globals StringAnyMap, preload ModuleLoaderList, lazyload ModuleLoaderMap) *Machine {
	return &Machine{
		enableInConv:  true,
		enableOutConv: true,
		globals:       globals,
		preloadMods:   preload,
		lazyloadMods:  lazyload,
	}
}

// NewWithBuiltins creates a new Starlark runtime environment with given global variables and all preload & lazyload built-in modules.
func NewWithBuiltins(globals StringAnyMap, additionalPreload ModuleLoaderList, additionalLazyload ModuleLoaderMap) *Machine {
	pre := append(GetAllBuiltinModules(), additionalPreload...)
	lazy := GetBuiltinModuleMap()
	lazy.Merge(additionalLazyload)
	return &Machine{
		enableInConv:  true,
		enableOutConv: true,
		globals:       globals,
		preloadMods:   pre,
		lazyloadMods:  lazy,
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
		enableInConv:  true,
		enableOutConv: true,
		globals:       globals,
		preloadMods:   pre,
		lazyloadMods:  lazy,
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

var (
	// NoopPrintFunc is a no-op print function for the Starlark runtime environment, it does nothing.
	NoopPrintFunc PrintFunc = func(thread *starlark.Thread, msg string) {}
)

// SetPrintFunc sets the print function of the Starlark runtime environment.
// To disable printing, you can set it to NoopPrintFunc. Setting it to nil will invoke the default `fmt.Fprintln(os.Stderr, msg)` instead.
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

// SetScriptContent sets the script content of the Starlark runtime environment.
// It differs from SetScript in that it does not change the script name and filesystem.
func (m *Machine) SetScriptContent(content []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.scriptContent = content
}

// SetInputConversionEnabled controls the conversion of Starlark variables from input into Starlight wrappers.
func (m *Machine) SetInputConversionEnabled(enabled bool) {
	m.mu.Lock() // Locking to avoid concurrent access
	defer m.mu.Unlock()

	m.enableInConv = enabled
}

// SetOutputConversionEnabled controls the conversion of Starlark variables from output into Starlight wrappers.
func (m *Machine) SetOutputConversionEnabled(enabled bool) {
	m.mu.Lock() // Locking to avoid concurrent access
	defer m.mu.Unlock()

	m.enableOutConv = enabled
}

// SetScriptCache sets the cache for compiled Starlark programs.
func (m *Machine) SetScriptCache(cache ByteCache) {
	m.mu.Lock() // Locking to avoid concurrent access
	defer m.mu.Unlock()

	m.progCache = cache
}

// SetScriptCacheEnabled controls the cache for compiled Starlark programs with the default in-memory cache.
// If enabled is true, it creates the default in-memory cache instance, otherwise it uses no cache.
func (m *Machine) SetScriptCacheEnabled(enabled bool) {
	m.mu.Lock() // Locking to avoid concurrent access
	defer m.mu.Unlock()

	if enabled {
		m.progCache = NewMemoryCache()
	} else {
		m.progCache = nil
	}
}

// SetCustomTag sets the custom annotation tag of Go struct fields for Starlark.
func (m *Machine) SetCustomTag(tag string) {
	m.mu.Lock() // Locking to avoid concurrent access
	defer m.mu.Unlock()

	m.customTag = tag
}

// GetStarlarkPredeclared returns the Starlark predeclared names of the Starlark runtime environment.
// It's for advanced usage only, don't use it unless you know what you are doing.
func (m *Machine) GetStarlarkPredeclared() starlark.StringDict {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.predeclared
}

// GetStarlarkThread returns the Starlark thread of the Starlark runtime environment.
// It's for advanced usage only, don't use it unless you know what you are doing.
func (m *Machine) GetStarlarkThread() *starlark.Thread {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.thread
}

// GetThreadLocal returns the local value of the Starlark thread of the Starlark runtime environment.
// It returns nil if the thread is not set or the key is not found. Please ensure the machine already runs before calling this method.
func (m *Machine) GetThreadLocal(key string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.thread == nil {
		return nil
	}
	return m.thread.Local(key)
}

// Export returns the current variables of the Starlark runtime environment.
func (m *Machine) Export() StringAnyMap {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.convertOutput(m.predeclared)
}

// EnableRecursionSupport enables recursion support in all Starlark environments.
func (m *Machine) EnableRecursionSupport() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowRecursion = true
}

// DisableRecursionSupport disables recursion support in all Starlark environments.
func (m *Machine) DisableRecursionSupport() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowRecursion = false
}

// EnableGlobalReassign enables global reassignment in all Starlark environments.
func (m *Machine) EnableGlobalReassign() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowGlobalReassign = true
}

// DisableGlobalReassign disables global reassignment in all Starlark environments.
func (m *Machine) DisableGlobalReassign() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowGlobalReassign = false
}
