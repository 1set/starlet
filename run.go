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
	ErrNoFileToRun         = errors.New("no specific file")
	ErrNoScriptSourceToRun = errors.New("no script to execute")
	ErrModuleNotFound      = errors.New("module not found")
)

type sourceCodeType uint8

const (
	sourceCodeTypeUnknown sourceCodeType = iota
	sourceCodeTypeContent
	sourceCodeTypeFSName
)

// Run runs the preset script with given globals and returns the result.
func (m *Machine) Run(ctx context.Context) (DataStore, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// TODO: handle panic, what about other defers?

	// either script content or name and FS must be set
	var (
		scriptName = m.scriptName
		srcType    sourceCodeType
	)
	if m.scriptContent != nil {
		srcType = sourceCodeTypeContent
		if scriptName == "" {
			scriptName = "eval.star"
		}
	} else if m.scriptFS != nil {
		srcType = sourceCodeTypeFSName
		if scriptName == "" {
			return nil, fmt.Errorf("starlet: run: %w", ErrNoFileToRun)
		}
	} else {
		return nil, fmt.Errorf("starlet: run: %w", ErrNoScriptSourceToRun)
	}

	// TODO: Assume: it's the first run -- for rerun, we need to reset the cache

	// preset globals + preload modules -> predeclared
	m.liveData = m.globals.Clone()
	if err := m.loadBuiltinModules(m.preloadMods...); err != nil {
		return nil, fmt.Errorf("starlet: preload: %w", err)
	}

	// convert into starlark.StringDict as predeclared
	predeclared, err := convert.MakeStringDict(m.liveData)
	if err != nil {
		return nil, fmt.Errorf("starlet: convert: %w", err)
	}

	// TODO: save or reuse thread
	// cache load + printFunc -> thread
	m.loadCache = &cache{
		cache:    make(map[string]*entry),
		loadMod:  m.loadAllowedModule,
		readFile: m.readScriptFile,
		globals:  predeclared,
	}
	thread := &starlark.Thread{
		Print: m.printFunc,
		Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
			return m.loadCache.Load(module)
		},
	}

	// TODO: run script with context and thread
	// run
	m.runTimes++
	var res starlark.StringDict
	switch srcType {
	case sourceCodeTypeContent:
		res, err = starlark.ExecFile(thread, scriptName, m.scriptContent, predeclared)
	case sourceCodeTypeFSName:
		rd, e := m.scriptFS.Open(scriptName)
		if e != nil {
			return nil, fmt.Errorf("starlet: open: %w", e)
		}
		res, err = starlark.ExecFile(thread, scriptName, rd, predeclared)
	}
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
			return fmt.Errorf("load module %q: %w", mod, err)
		} else if dict == nil {
			return fmt.Errorf("load module %q: %w", mod, ErrModuleNotFound)
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
	// TODO: check if module is already loaded as predeclared
	for _, mod := range m.allowMods {
		if mod == ModuleName(name) {
			// load module by name if it's allowed
			return loadModuleByName(mod)
		}
	}
	// module not found
	return nil, nil
}
