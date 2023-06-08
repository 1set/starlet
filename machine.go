package starlet

import (
	"io/fs"
	"sync"

	"go.starlark.net/starlark"
)

// PrintFunc is a function that tells Starlark how to print messages.
// If nil, the default `fmt.Fprintln(os.Stderr, msg)` will be used instead.
type PrintFunc func(thread *starlark.Thread, msg string)

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
	runTimes  uint
	thread    *starlark.Thread
	loadCache *cache
}

// NewEmptyMachine creates a new Starlark runtime environment.
func NewEmptyMachine() *Machine {
	return &Machine{}
}

// NewMachine creates a new Starlark runtime environment with given globals, preload and lazyload module names.
// TODO: Rename it as simple or easy. With Must in Name
func NewMachine(globals DataStore, preloads []string, lazyloads []string) *Machine {
	pre, err := CreateBuiltinModuleLoaderList(preloads)
	if err != nil {
		panic(err)
	}
	lazy, err := CreateBuiltinModuleLoaderMap(lazyloads)
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
func (m *Machine) SetGlobals(globals DataStore) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.globals = globals
}

// AddGlobals adds the globals of the Starlark runtime environment.
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
