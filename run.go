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
	ErrUnknownScriptSource = errors.New("unknown script source")
	ErrModuleNotFound      = errors.New("module not found")
)

// Run runs the preset script with given globals and returns the result.
func (m *Machine) Run(ctx context.Context) (DataStore, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// TODO: handle panic, what about other defers?

	// either script content or name and FS must be set
	if !((m.scriptContent != nil) || (m.scriptName != "" && m.scriptFS != nil)) {
		return nil, fmt.Errorf("starlet: run: %w", ErrUnknownScriptSource)
	}

	// TODO: Assume: it's the first run -- for rerun, we need to reset the cache

	// clone preset globals if it's the first run, otherwise merge if newer
	if m.liveData == nil {
		m.liveData = m.globals.Clone()
	} else {
		m.liveData.Merge(m.globals)
	}

	// load preload modules
	if err := m.loadBuiltinModules(m.preloadMods...); err != nil {
		return nil, fmt.Errorf("starlet: load preload modules: %w", err)
	}

	// convert into starlark.StringDict as predeclared
	predeclared, err := convert.MakeStringDict(m.liveData)
	if err != nil {
		return nil, fmt.Errorf("starlet: convert predeclared: %w", err)
	}

	// create cache
	if m.loadCache == nil {
		m.loadCache = &cache{
			cache:    make(map[string]*entry),
			readFile: m.readScriptFile,
			globals:  predeclared,
		}
	}

	// thread = cache.Load + printFunc
	// TODO: save or reuse thread
	thread := &starlark.Thread{
		Load:  m.cacheLoader,
		Print: m.printFunc,
	}

	// run
	// TODO: run script with context and thread
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

// cacheLoader is a starlark.Loader that loads modules from built-in modules and cache.
func (m *Machine) cacheLoader(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	// TODO: what if module is already loaded?
	for _, mod := range m.allowMods {
		// find module by name
		if string(mod) == module {
			if dict, err := loadModuleByName(mod); err != nil {
				return nil, fmt.Errorf("starlet: load module %q: %w", mod, err)
			} else {
				m.loadMod[mod] = struct{}{}
				return dict, nil
			}
		}
	}

	// built-in module not found
	// TODO: maybe script module can't use built-in module -- refine cache things
	return m.loadCache.Load(module)
}
