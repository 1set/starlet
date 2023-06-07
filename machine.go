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
	globals     DataStore
	preloadMods ModuleNameList
	allowMods   ModuleNameList
	printFunc   PrintFunc
	// source code
	scriptName    string
	scriptContent []byte
	scriptFS      fs.FS
	// runtime core
	thread    *starlark.Thread
	coreCache *Cache
}

// NewEmptyMachine creates a new Starlark runtime environment.
func NewEmptyMachine() *Machine {
	return &Machine{}
}

// NewMachine creates a new Starlark runtime environment with given globals, preload modules and modules allowed to be loaded.
func NewMachine(globals DataStore, preloads ModuleNameList, allows ModuleNameList) *Machine {
	return &Machine{
		globals:     globals,
		preloadMods: preloads,
		allowMods:   allows,
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
func (m *Machine) SetPreloadModules(mods ModuleNameList) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.preloadMods = mods
}

// GetPreloadModules gets the preload modules of the Starlark runtime environment.
func (m *Machine) GetPreloadModules() ModuleNameList {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.preloadMods.Clone()
}

// SetAllowModules sets the modules allowed to be loaded of the Starlark runtime environment.
func (m *Machine) SetAllowModules(mods ModuleNameList) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.allowMods = mods
}

// GetAllowModules gets the modules allowed to be loaded of the Starlark runtime environment.
func (m *Machine) GetAllowModules() ModuleNameList {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.allowMods.Clone()
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
