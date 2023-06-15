package starlet

import (
	"fmt"
	"io/fs"
	"sync"

	"go.starlark.net/starlark"
)

// Machine is a wrapper of Starlark runtime environments.
type Machine struct {
	mu sync.RWMutex
	// set variables
	globals      StringAny
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
	lastResult  starlark.StringDict
}

func (m *Machine) String() string {
	steps := uint64(0)
	if m.thread != nil {
		steps = m.thread.Steps
	}
	return fmt.Sprintf("🌠Machine{run:%d,step:%d,script:%q,len:%d,fs:%v}",
		m.runTimes, steps, m.scriptName, len(m.scriptContent), m.scriptFS)
}

// NewDefault creates a new Starlark runtime environment.
func NewDefault() *Machine {
	return &Machine{}
}

// NewWithGlobals creates a new Starlark runtime environment with given global variables.
func NewWithGlobals(globals StringAny) *Machine {
	return &Machine{
		globals: globals,
	}
}

// NewWithLoaders creates a new Starlark runtime environment with given global variables and preload module loaders.
func NewWithLoaders(globals StringAny, preload ModuleLoaderList, lazyload ModuleLoaderMap) *Machine {
	return &Machine{
		globals:      globals,
		preloadMods:  preload,
		lazyloadMods: lazyload,
	}
}

// NewWithNames creates a new Starlark runtime environment with given global variables, preload and lazyload module names.
// The modules should be built-in modules, and it panics if any of the given modules fails to load.
func NewWithNames(globals StringAny, preloads []string, lazyloads []string) *Machine {
	pre, err := MakeBuiltinModuleLoaderList(preloads)
	if err != nil {
		panic(err)
	}
	lazy, err := MakeBuiltinModuleLoaderMap(lazyloads)
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
func (m *Machine) SetGlobals(globals StringAny) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.globals = globals
}

// AddGlobals adds the globals of the Starlark runtime environment.
// These variables only take effect before the first run or after a reset.
func (m *Machine) AddGlobals(globals StringAny) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.globals == nil {
		m.globals = make(StringAny)
	}
	for k, v := range globals {
		m.globals[k] = v
	}
}

// GetGlobals gets the globals of the Starlark runtime environment.
func (m *Machine) GetGlobals() StringAny {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.globals
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

// StringAny is a map of string to interface{} (i.e. any).
// It is used to store global variables like StringDict of Starlark, but not a Starlark type.
type StringAny map[string]interface{}

// Clone returns a copy of the data store.
func (d StringAny) Clone() StringAny {
	clone := make(StringAny)
	for k, v := range d {
		clone[k] = v
	}
	return clone
}

// Merge merges the given data store into the current data store.
func (d StringAny) Merge(other StringAny) {
	if d == nil {
		return
	}
	for k, v := range other {
		d[k] = v
	}
}

// MergeDict merges the given string dict into the current data store.
func (d StringAny) MergeDict(other starlark.StringDict) {
	if d == nil {
		return
	}
	for k, v := range other {
		d[k] = v
	}
}
