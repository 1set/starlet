package starlet

import (
	"context"
	"errors"
)

var (
	ErrUnknownScriptSource = errors.New("starlet: unknown script source")
)

// Run runs the preset script with given globals and returns the result.
func (m *Machine) Run(ctx context.Context) (DataStore, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// either script content or name and FS must be set
	if !((m.scriptContent != nil) || (m.scriptName != "" && m.scriptFS != nil)) {
		return nil, ErrUnknownScriptSource
	}

	// Assume: it's the first run
	m.runTimes++

	// clone preset globals if it's the first run, otherwise merge if newer
	if m.liveData == nil {
		m.liveData = m.globals.Clone()
	} else {
		m.liveData.Merge(m.globals)
	}

	// load preload modules
	if err := m.loadBuiltinModules(m.preloadMods...); err != nil {
		return nil, err
	}

	// clone globals + preset modules -> predeclared
	// convert predeclared to starlark.StringDict
	// create cache with predeclared + module allowed + fs reader with deduped loader
	// thread = cache.Load + printFunc
	// saved thread
	// run script with context and thread
	// convert result to DataStore

	// TODO: implement
	return nil, nil
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
			return err
		} else {
			m.liveData.MergeDict(dict)
		}
		// mark as loaded
		m.loadMod[mod] = struct{}{}
	}
	return nil
}
