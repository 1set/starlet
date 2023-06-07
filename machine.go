package starlet

import "sync"

// Machine is a wrapper of Starlark runtime environments.
type Machine struct {
	mu          sync.RWMutex
	globals     map[string]interface{}
	preloadMods []ModuleName
	allowMods   []ModuleName
}

// NewMachine creates a new Starlark runtime environment with given globals, preload modules and modules allowed to be loaded.
func NewMachine(globals map[string]interface{}, preloads []ModuleName, allows []ModuleName) *Machine {
	return &Machine{
		globals:     globals,
		preloadMods: preloads,
		allowMods:   allows,
	}
}

// NewEmptyMachine creates a new Starlark runtime environment.
func NewEmptyMachine() *Machine {
	return &Machine{}
}
