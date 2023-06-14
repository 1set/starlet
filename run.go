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
func (m *Machine) Run(ctx context.Context) (out DataStore, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("starlet: panic: %v", r)
		}
	}()

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

	// prepare thread
	if m.thread == nil {
		// -- for the first run
		// preset globals + preload modules -> predeclared
		if m.predeclared, err = convert.MakeStringDict(m.globals); err != nil {
			return nil, fmt.Errorf("starlet: convert: %w", err)
		}
		if err = m.preloadMods.LoadAll(m.predeclared); err != nil {
			// TODO: wrap the errors
			return nil, err
		}

		// cache load&read + printf -> thread
		m.loadCache = &cache{
			cache:    make(map[string]*entry),
			loadMod:  m.lazyloadMods.GetLazyLoader(),
			readFile: m.readFSFile,
			globals:  m.predeclared,
		}
		m.thread = &starlark.Thread{
			Print: m.printFunc,
			Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
				return m.loadCache.Load(module)
			},
		}
	} else if m.lastResult != nil {
		// -- for the second and following runs
		m.thread.Uncancel()
		// merge last result as globals
		for k, v := range m.lastResult {
			m.predeclared[k] = v
		}
		// set globals for cache
		m.loadCache.loadMod = m.lazyloadMods.GetLazyLoader()
		m.loadCache.globals = m.predeclared
	}

	// reset for each run
	m.thread.Print = m.printFunc
	m.thread.SetLocal("context", ctx)

	// cancel thread when context cancelled
	m.runTimes++
	if ctx != nil {
		go func() {
			<-ctx.Done()
			m.thread.Cancel("context cancelled")
		}()
	}

	// run for various source code types
	var res starlark.StringDict
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
	// TODO: merge back into code above

	// handle result and convert
	m.lastResult = res
	out = convert.FromStringDict(res)
	if err != nil {
		// for exit code
		if err.Error() == `starlet runtime system exit` {
			var exitCode uint8 = 0
			if c := m.thread.Local("exit_code"); c != nil {
				if co, ok := c.(uint8); ok {
					exitCode = co
				}
			}
			// exit code 0 means success
			if exitCode == 0 {
				err = nil
			} else {
				err = fmt.Errorf("starlet: exit code: %d", exitCode)
			}
		} else {
			// wrap other errors
			err = fmt.Errorf("starlet: exec: %w", err)
		}

		// TODO: call it convert error? maybe better error solutions
		return out, err
	}
	return out, nil
}

// Reset resets the machine to initial state before the first run.
func (m *Machine) Reset() {
	m.runTimes = 0
	m.lastResult = nil
	m.thread = nil
	m.loadCache = nil
	m.predeclared = nil
}

func (m *Machine) readFSFile(name string) ([]byte, error) {
	return readScriptFile(name, m.scriptFS)
}
