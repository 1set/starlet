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
	globals      DataStore
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
func NewWithGlobals(globals DataStore) *Machine {
	return &Machine{
		globals: globals,
	}
}

// NewWithLoaders creates a new Starlark runtime environment with given global variables and preload module loaders.
func NewWithLoaders(globals DataStore, preload ModuleLoaderList, lazyload ModuleLoaderMap) *Machine {
	return &Machine{
		globals:      globals,
		preloadMods:  preload,
		lazyloadMods: lazyload,
	}
}

// NewWithNames creates a new Starlark runtime environment with given global variables, preload and lazyload module names.
// It panics if any of the given module fails to load.
func NewWithNames(globals DataStore, preloads []string, lazyloads []string) *Machine {
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

// SetGlobals sets the globals of the Starlark runtime environment.
// It only works before the first run.
func (m *Machine) SetGlobals(globals DataStore) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.globals = globals
}

// AddGlobals adds the globals of the Starlark runtime environment.
// It only works before the first run.
func (m *Machine) AddGlobals(globals DataStore) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.globals == nil {
		m.globals = make(DataStore)
	}
	for k, v := range globals {
		m.globals[k] = v
	}
}

// GetGlobals gets the globals of the Starlark runtime environment.
func (m *Machine) GetGlobals() DataStore {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.globals
}

// SetPreloadModules sets the preload modules of the Starlark runtime environment.
// It only works before the first run.
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
