package starlet

import (
	"context"
	"errors"
	"fmt"

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
	predeclared, err := convert.MakeStringDict(m.globals.Clone())
	if err != nil {
		return nil, fmt.Errorf("starlet: convert: %w", err)
	}
	if err = m.preloadMods.LoadAll(predeclared); err != nil {
		return nil, err
	}

	// TODO: save or reuse thread
	// cache load + printFunc -> thread
	m.loadCache = &cache{
		cache:    make(map[string]*entry),
		loadMod:  m.lazyloadMods.GetLazyLoader(),
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

// readScriptFile reads the given filename from the given file system.
func (m *Machine) readScriptFile(filename string) ([]byte, error) {
	return readScriptFile(filename, m.scriptFS)
}
