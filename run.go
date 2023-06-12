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

	// for the first run
	var err error
	if m.thread == nil {
		// preset globals + preload modules -> predeclared
		if m.predeclared, err = convert.MakeStringDict(m.globals); err != nil {
			return nil, fmt.Errorf("starlet: convert: %w", err)
		}
		if err = m.preloadMods.LoadAll(m.predeclared); err != nil {
			return nil, err
		}

		// cache load + printFunc -> thread
		m.loadCache = &cache{
			cache:   make(map[string]*entry),
			loadMod: m.lazyloadMods.GetLazyLoader(),
			readFile: func(name string) ([]byte, error) {
				return readScriptFile(name, m.scriptFS)
			},
			globals: m.predeclared,
		}
		m.thread = &starlark.Thread{
			Print: m.printFunc,
			Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
				return m.loadCache.Load(module)
			},
		}
	} else {
		// for the following runs
		if m.lastResult != nil {
			// merge last result as globals
			for k, v := range m.lastResult {
				m.predeclared[k] = v
			}
			// set globals for cache
			m.loadCache.globals = m.predeclared
		}
		// set printFunc for thread anyway
	}

	// for each run commons
	m.thread.Print = m.printFunc
	m.thread.SetLocal("context", ctx)

	// TODO: run script with context and thread
	// run for various source code types
	m.runTimes++
	var (
		res starlark.StringDict
	)
	switch srcType {
	case sourceCodeTypeContent:
		res, err = starlark.ExecFile(m.thread, scriptName, m.scriptContent, m.predeclared)
	case sourceCodeTypeFSName:
		rd, e := m.scriptFS.Open(scriptName)
		if e != nil {
			return nil, fmt.Errorf("starlet: open: %w", e)
		}
		res, err = starlark.ExecFile(m.thread, scriptName, rd, m.predeclared)
	}

	// handle result and convert
	m.lastResult = res
	if err != nil {
		return nil, fmt.Errorf("starlet: exec: %w", err)
	}
	return convert.FromStringDict(res), nil
}

// Reset resets the machine to initial state before the first run.
func (m *Machine) Reset() {
	m.runTimes = 0
	m.lastResult = nil
	m.thread = nil
	m.loadCache = nil
	m.predeclared = nil
}
