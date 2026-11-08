// Package starlet embeds Starlark scripts in Go with a stateful Machine,
// builtin libraries, and Go/Starlark value conversion.
package starlet

import (
	"fmt"
	"io/fs"
	"sync"

	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
)

// Machine configures and executes a stateful Starlark environment.
// Configure globals, preloads and dialect options before execution or after Reset.
// Preloads inject their exported values; lazy loaders supply load() members.
// Binding precedence is globals, preloads, execution extras, then explicit load
// bindings. Extras and script globals remain available to later executions.
// Reset discards runtime bindings and loaded modules, retaining configuration
// and the optional compiled-program cache; it does not undo host-side effects.
//
// Machine serializes execution and locks access to its state. A callback must
// not reenter methods on the same Machine, except the nonblocking String method. Externally
// shared Go data and raw handles returned by GetStarlarkThread or
// GetStarlarkPredeclared still require host ownership or synchronization.
type Machine struct {
	_  itn.DoNotCompare
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
	maxSteps            uint64
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

// String returns a state summary without blocking. If the Machine is locked,
// it returns a running marker, allowing safe formatting from a host callback.
func (m *Machine) String() string {
	if !m.mu.TryRLock() {
		return "🌠Machine{running}"
	}
	defer m.mu.RUnlock()

	steps := uint64(0)
	if t := m.thread; t != nil {
		steps = t.Steps
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

// NewDefault creates a Machine with standard Starlark builtins, no additional
// modules, and input/output conversion enabled.
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

// NewWithNames creates a Machine with globals and named builtin preloads/loaders.
// It panics on unknown names; module initialization happens during execution.
// For caller-supplied names, handle errors from MakeBuiltinModuleLoaderList and
// MakeBuiltinModuleLoaderMap, then pass the loaders to NewWithLoaders.
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

// SetInputConversionEnabled controls Go-to-Starlark conversion of globals and
// execution extras through Starlight. It is enabled by default; when disabled,
// those inputs must already implement starlark.Value. Call arguments always
// use the Starlight bridge, independently of this setting.
func (m *Machine) SetInputConversionEnabled(enabled bool) {
	m.mu.Lock() // Locking to avoid concurrent access
	defer m.mu.Unlock()

	m.enableInConv = enabled
}

// SetOutputConversionEnabled controls Starlark-to-Go conversion of execution,
// Call and Export results. It is enabled by default; when disabled, result
// entries retain their starlark.Value types.
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

// GetStarlarkPredeclared returns the live runtime dictionary, not a copy, or nil
// before initialization.
// Advanced callers must synchronize access with all Machine operations.
func (m *Machine) GetStarlarkPredeclared() starlark.StringDict {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.predeclared
}

// GetStarlarkThread returns the live runtime Thread, or nil before initialization.
// Advanced callers must synchronize access with all Machine operations.
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

// SetMaxExecutionSteps sets the per-execution budget of Starlark steps (an
// abstract measure of interpreter work); 0 (the default) means unlimited.
// When the budget is exhausted, the Run or Call fails with a
// MaxStepsExceededError, reachable host-side via errors.As. The step
// counter resets at the start of every Run/Call, so the budget applies per
// execution, not cumulatively.
func (m *Machine) SetMaxExecutionSteps(steps uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.maxSteps = steps
}

// EnableRecursionSupport enables recursion in this Machine's file options.
func (m *Machine) EnableRecursionSupport() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowRecursion = true
}

// DisableRecursionSupport disables recursion in this Machine's file options.
func (m *Machine) DisableRecursionSupport() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowRecursion = false
}

// EnableGlobalReassign enables global reassignment, top-level control flow and
// while loops in this Machine's file options.
func (m *Machine) EnableGlobalReassign() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowGlobalReassign = true
}

// DisableGlobalReassign disables global reassignment, top-level control flow
// and while loops in this Machine's file options.
func (m *Machine) DisableGlobalReassign() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowGlobalReassign = false
}
