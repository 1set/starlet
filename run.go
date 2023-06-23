package starlet

import (
	"context"
	"io/fs"
	"sync"
	"time"

	"github.com/1set/starlet/lib/goidiomatic"
	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
)

// PrintFunc is a function that tells Starlark how to print messages.
// If nil, the default `fmt.Fprintln(os.Stderr, msg)` will be used instead.
type PrintFunc func(thread *starlark.Thread, msg string)

// LoadFunc is a function that tells Starlark how to find and load other scripts
// using the load() function. If you don't use load() in your scripts, you can pass in nil.
type LoadFunc func(thread *starlark.Thread, module string) (starlark.StringDict, error)

// Run executes a preset script and returns the output.
func (m *Machine) Run() (StringAnyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.runInternal(context.Background(), nil)
}

// RunScript executes a script with additional variables, which take precedence over global variables and modules, returns the result.
func (m *Machine) RunScript(content []byte, extras StringAnyMap) (StringAnyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.scriptName = "direct.star"
	m.scriptContent = content
	m.scriptFS = nil
	return m.runInternal(context.Background(), extras)
}

// RunFile executes a script from a file with additional variables, which take precedence over global variables and modules, returns the result.
func (m *Machine) RunFile(name string, fileSys fs.FS, extras StringAnyMap) (StringAnyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.scriptName = name
	m.scriptContent = nil
	m.scriptFS = fileSys
	return m.runInternal(context.Background(), extras)
}

// RunWithTimeout executes a preset script with a timeout and additional variables, which take precedence over global variables and modules, returns the result.
func (m *Machine) RunWithTimeout(timeout time.Duration, extras StringAnyMap) (StringAnyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return m.runInternal(ctx, extras)
}

// RunWithContext executes a preset script within a specified context and additional variables, which take precedence over global variables and modules, returns the result.
func (m *Machine) RunWithContext(ctx context.Context, extras StringAnyMap) (StringAnyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.runInternal(ctx, extras)
}

func (m *Machine) runInternal(ctx context.Context, extras StringAnyMap) (out StringAnyMap, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errorStarlarkPanic("exec", r)
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
			return nil, errorStarletErrorf("run", "no script name")
		}
		// load script from FS
		rd, e := m.scriptFS.Open(scriptName)
		if e != nil {
			return nil, errorStarletError("run", e)
		}
		source = rd
	} else {
		return nil, errorStarletErrorf("run", "no script to execute")
	}

	// prepare thread
	if err = m.prepareThread(extras); err != nil {
		return nil, err
	}

	// cancel thread when context cancelled
	if ctx == nil || ctx.Err() != nil {
		// for nil context, or context already cancelled, use a new one
		ctx = context.TODO()
	}
	m.thread.SetLocal("context", ctx)

	// wait for the routine to finish, or cancel it when context cancelled
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)

	// if context is not cancelled, cancel the routine when execution is done, or panic
	done := make(chan struct{}, 1)
	defer close(done)

	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			m.thread.Cancel("context cancelled")
		case <-done:
			// No action if the script has finished
		}
	}()

	// run with everything prepared
	m.runTimes++
	res, err := starlark.ExecFile(m.thread, scriptName, source, m.predeclared)
	done <- struct{}{}

	// merge result as predeclared for next run
	for k, v := range res {
		m.predeclared[k] = v
	}

	// handle result and convert
	out = convert.FromStringDict(res)
	if err != nil {
		// for exit code
		if err.Error() == goidiomatic.ErrSystemExit.Error() {
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
				err = errorStarletErrorf("run", "exit code: %d", exitCode)
			}
		} else {
			// wrap other errors
			err = errorStarlarkError("exec", err)
		}

		// TODO: call it convert error? maybe better error solutions
		return out, err
	}
	return out, nil
}

// prepareThread prepares the thread for execution, including preset globals, preload modules and extras.
func (m *Machine) prepareThread(extras StringAnyMap) (err error) {
	if m.thread == nil {
		// -- for the first run
		// preset globals + preload modules + extras -> predeclared
		if m.predeclared, err = convert.MakeStringDict(m.globals); err != nil {
			return errorStarlightConvert("globals", err)
		}
		if err = m.preloadMods.LoadAll(m.predeclared); err != nil {
			return errorStarletError("preload", err)
		}
		esd, err := convert.MakeStringDict(extras)
		if err != nil {
			return errorStarlightConvert("extras", err)
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
			Name:  "starlet",
			Print: m.printFunc,
			Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
				return m.loadCache.Load(module)
			},
		}
	} else {
		// -- for the second and following runs
		// set globals for cache
		m.loadCache.loadMod = m.lazyloadMods.GetLazyLoader()
		m.loadCache.globals = m.predeclared
		// reset for each run
		m.thread.Print = m.printFunc
		m.thread.Uncancel()
	}
	return nil
}

// Reset resets the machine to initial state before the first run.
func (m *Machine) Reset() {
	m.runTimes = 0
	m.thread = nil
	m.loadCache = nil
	m.predeclared = nil
}

func (m *Machine) readFSFile(name string) ([]byte, error) {
	return readScriptFile(name, m.scriptFS)
}
