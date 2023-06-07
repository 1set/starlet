package starlet

import "sync"

// Machine is a wrapper of Starlark runtime environments.
type Machine struct {
	mu          sync.RWMutex
	globals     map[string]interface{}
	preloadMods ModuleNameList
	allowMods   ModuleNameList
}

// NewEmptyMachine creates a new Starlark runtime environment.
func NewEmptyMachine() *Machine {
	return &Machine{}
}

// NewMachine creates a new Starlark runtime environment with given globals, preload modules and modules allowed to be loaded.
func NewMachine(globals map[string]interface{}, preloads ModuleNameList, allows ModuleNameList) *Machine {
	return &Machine{
		globals:     globals,
		preloadMods: preloads,
		allowMods:   allows,
	}
}

// SetGlobals sets the globals of the Starlark runtime environment.
func (m *Machine) SetGlobals(globals map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.globals = globals
}

// GetGlobals gets the globals of the Starlark runtime environment.
func (m *Machine) GetGlobals() map[string]interface{} {
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
