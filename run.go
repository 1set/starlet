package starlet

import (
	"context"
	"errors"
	"fmt"

	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
)

// PrintFunc is a function that tells Starlark how to print messages.
// If nil, the default `fmt.Fprintln(os.Stderr, msg)` will be used instead.
type PrintFunc func(thread *starlark.Thread, msg string)

// LoadFunc is a function that tells Starlark how to find and load other scripts
// using the load() function. If you don't use load() in your scripts, you can pass in nil.
type LoadFunc func(thread *starlark.Thread, module string) (starlark.StringDict, error)

var (
	ErrNoFileToRun         = errors.New("no specific file")
	ErrNoScriptSourceToRun = errors.New("no script to execute")
	ErrModuleNotFound      = errors.New("module not found")
)

// Run runs the preset script with given globals and returns the result.
func (m *Machine) Run(ctx context.Context, extras map[string]interface{}) (out DataStore, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.internalRun(ctx, extras)
}

func (m *Machine) internalRun(ctx context.Context, extras map[string]interface{}) (out DataStore, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("starlet: panic: %v", r)
		}
	}()

	// either script content or name and FS must be set
	var (
		scriptName = m.scriptName
		source     interface{}
	)
	if m.scriptContent != nil {
		if scriptName == "" {
			// for default name
			scriptName = "eval.star"
		}
		source = m.scriptContent
	} else if m.scriptFS != nil {
		if scriptName == "" {
			// if no name, cannot load
			return nil, fmt.Errorf("starlet: run: %w", ErrNoFileToRun)
		}
		// load script from FS
		rd, e := m.scriptFS.Open(scriptName)
		if e != nil {
			return nil, fmt.Errorf("starlet: open: %w", e)
		}
		source = rd
	} else {
		return nil, fmt.Errorf("starlet: run: %w", ErrNoScriptSourceToRun)
	}

	// prepare thread
	if m.thread == nil {
		// -- for the first run
		// preset globals + preload modules + extras -> predeclared
		if m.predeclared, err = convert.MakeStringDict(m.globals); err != nil {
			return nil, fmt.Errorf("starlet: convert globals: %w", err)
		}
		if err = m.preloadMods.LoadAll(m.predeclared); err != nil {
			// TODO: wrap the errors
			return nil, err
		}
		esd, err := convert.MakeStringDict(extras)
		if err != nil {
			// TODO: test it
			return nil, fmt.Errorf("starlet: convert extras: %w", err)
		}
		for k, v := range esd {
			m.predeclared[k] = v
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

	// run with everything prepared
	res, err := starlark.ExecFile(m.thread, scriptName, source, m.predeclared)

	// handle result and convert
	m.lastResult = res
	out = convert.FromStringDict(res)
	if err != nil {
		// for exit code
		if err.Error() == `starlet runtime system exit` {
			var exitCode uint8
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
