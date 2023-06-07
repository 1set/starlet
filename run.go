package starlet

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
)

var (
	ErrNoCodeToRun    = errors.New("no code to run")
	ErrModuleNotFound = errors.New("module not found")
)

// Run runs the preset script with given globals and returns the result.
func (m *Machine) Run(ctx context.Context) (DataStore, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// TODO: handle panic, what about other defers?

	// either script content or name and FS must be set
	if !((m.scriptContent != nil) || (m.scriptName != "" && m.scriptFS != nil)) {
		return nil, fmt.Errorf("starlet: run: %w", ErrNoCodeToRun)
	}

	// TODO: Assume: it's the first run -- for rerun, we need to reset the cache

	// preset globals + preload modules -> predeclared
	m.liveData = m.globals.Clone()
	if err := m.loadBuiltinModules(m.preloadMods...); err != nil {
		return nil, fmt.Errorf("starlet: load preload modules: %w", err)
	}

	// convert into starlark.StringDict as predeclared
	predeclared, err := convert.MakeStringDict(m.liveData)
	if err != nil {
		return nil, fmt.Errorf("starlet: convert predeclared: %w", err)
	}

	// TODO: save or reuse thread
	// cache load + printFunc -> thread
	m.loadCache = &cache{
		cache:      make(map[string]*entry),
		loadModule: m.loadAllowedModule,
		readFile:   m.readScriptFile,
		globals:    predeclared,
	}
	thread := &starlark.Thread{
		Print: m.printFunc,
		Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
			return m.loadCache.Load(module)
		},
	}

	// TODO: run script with context and thread
	// run
	scriptName := m.scriptName
	if scriptName == "" {
		scriptName = "eval.star"
	}
	m.runTimes++

	res, err := starlark.ExecFile(thread, scriptName, m.scriptContent, predeclared)
	if err != nil {
		return nil, fmt.Errorf("starlet: exec: %w", err)
	}

	// convert result to DataStore
	return convert.FromStringDict(res), nil
}

// TODO: Multiple FS for script and modules
// TODO: Reset machine
// TODO: run with existing threads (global and module preset)

func (m *Machine) loadBuiltinModules(modules ...ModuleName) error {
	if m.loadMod == nil {
		m.loadMod = make(map[ModuleName]struct{})
	}
	for _, mod := range modules {
		// skip if already loaded
		if _, ok := m.loadMod[mod]; ok {
			continue
		}
		// load module and merge into live data
		if dict, err := loadModuleByName(mod); err != nil {
			return fmt.Errorf("starlet: load module %q: %w", mod, err)
		} else if dict == nil {
			return fmt.Errorf("starlet: load module %q: %w", mod, ErrModuleNotFound)
		} else {
			m.liveData.MergeDict(dict)
		}
		// mark as loaded
		m.loadMod[mod] = struct{}{}
	}
	return nil
}

// readScriptFile reads the given filename from the given file system.
func (m *Machine) readScriptFile(filename string) ([]byte, error) {
	if m.scriptFS == nil {
		return nil, fmt.Errorf("no file system given")
	}
	rd, err := m.scriptFS.Open(filename)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(rd)
}

// loadAllowedModule loads a module by name if it's allowed.
func (m *Machine) loadAllowedModule(name string) (starlark.StringDict, error) {
	for _, mod := range m.allowMods {
		if mod == ModuleName(name) {
			// load module by name if it's allowed
			return loadModuleByName(mod)
		}
	}
	// module not found
	return nil, nil
}
