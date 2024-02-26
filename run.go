package starlet

import (
	"context"
	"fmt"
	"io/fs"
	"sync"
	"time"

	"github.com/1set/starlet/lib/goidiomatic"
	"github.com/1set/starlight/convert"
	"go.starlark.net/repl"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// REPL is a Read-Eval-Print-Loop for Starlark.
// It loads the predeclared symbols and modules into the global environment,
func (m *Machine) REPL() {
	if err := m.prepareThread(nil); err != nil {
		repl.PrintError(err)
		return
	}
	repl.REPLOptions(m.getFileOptions(), m.thread, m.predeclared)
}

// RunScript initiates a Machine, executes a script with extra variables, and returns the Machine and the execution result.
func RunScript(content []byte, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	m := NewDefault()
	res, err := m.RunScript(content, extras)
	return m, res, err
}

// RunFile initiates a Machine, executes a script from a file with extra variables, and returns the Machine and the execution result.
func RunFile(name string, fileSys fs.FS, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	m := NewDefault()
	res, err := m.RunFile(name, fileSys, extras)
	return m, res, err
}

// RunTrustedScript initiates a Machine, executes a script with all builtin modules loaded and extra variables, returns the Machine and the result.
// Use with caution as it allows script access to file system and network.
func RunTrustedScript(content []byte, globals, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	m := NewWithBuiltins(globals, nil, nil)
	res, err := m.RunScript(content, extras)
	return m, res, err
}

// RunTrustedFile initiates a Machine, executes a script from a file with all builtin modules loaded and extra variables, returns the Machine and the result.
// Use with caution as it allows script access to file system and network.
func RunTrustedFile(name string, fileSys fs.FS, globals, extras StringAnyMap) (*Machine, StringAnyMap, error) {
	m := NewWithBuiltins(globals, nil, nil)
	res, err := m.RunFile(name, fileSys, extras)
	return m, res, err
}

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
	//res, err := starlark.ExecFileOptions(m.getFileOptions(), m.thread, scriptName, source, m.predeclared)
	res, err := starlarkExecFile(m.getFileOptions(), m.thread, scriptName, source, m.predeclared)
	done <- struct{}{}

	// merge result as predeclared for next run
	for k, v := range res {
		m.predeclared[k] = v
	}

	// handle result and convert
	out = m.convertOutput(res)
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
			// wrap starlark errors
			err = errorStarlarkError("exec", err)
		}
		return out, err
	}
	return out, nil
}

// prepareThread prepares the thread for execution, including preset globals, preload modules and extras.
func (m *Machine) prepareThread(extras StringAnyMap) (err error) {
	if m.thread == nil {
		// -- for the first run
		// preset globals + preload modules + extras -> predeclared
		if m.predeclared, err = m.convertInput(m.globals); err != nil {
			return errorStarlightConvert("globals", err)
		}
		if err = m.preloadMods.LoadAll(m.predeclared); err != nil {
			return errorStarletError("preload", err)
		}

		// convert extras or not
		esd, err := m.convertInput(extras)
		if err != nil {
			return errorStarlightConvert("extras", err)
		}
		// merge extras
		for k, v := range esd {
			m.predeclared[k] = v
		}

		// cache load&read + printf -> thread
		m.loadCache = &cache{
			cache:   make(map[string]*entry),
			loadMod: m.lazyloadMods.GetLazyLoader(),
			readFile: func(name string) ([]byte, error) {
				return readScriptFile(name, m.scriptFS)
			},
			globals: m.predeclared,
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

// convertInput converts a StringAnyMap to a starlark.StringDict, usually for output variable.
func (m *Machine) convertInput(a StringAnyMap) (starlark.StringDict, error) {
	if m.enableInConv {
		return convert.MakeStringDictWithTag(a, m.customTag)
	}
	return castStringAnyMapToStringDict(a)
}

// convertOutput converts a starlark.StringDict to a StringAnyMap, usually for output variable.
func (m *Machine) convertOutput(d starlark.StringDict) StringAnyMap {
	if m.enableOutConv {
		return convert.FromStringDict(d)
	}
	return castStringDictToAnyMap(d)
}

// getFileOptions gets the exec options from the config.
func (m *Machine) getFileOptions() *syntax.FileOptions {
	opt := syntax.FileOptions{
		Set: true,
	}
	if m.allowRecursion {
		opt.Recursion = true
	}
	if m.allowGlobalReassign {
		opt.GlobalReassign = true
		opt.TopLevelControl = true
		opt.While = true
	}
	return &opt
}

// castStringDictToAnyMap converts a starlark.StringDict to a StringAnyMap without any Starlight conversion.
func castStringDictToAnyMap(m starlark.StringDict) StringAnyMap {
	ret := make(StringAnyMap, len(m))
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

// castStringAnyMapToStringDict converts a StringAnyMap to a starlark.StringDict without any Starlight conversion.
// It fails if any values are not starlark.Value.
func castStringAnyMapToStringDict(m StringAnyMap) (starlark.StringDict, error) {
	ret := make(starlark.StringDict, len(m))
	for k, v := range m {
		sv, ok := v.(starlark.Value)
		if !ok {
			return nil, fmt.Errorf("value of key %q is not a starlark.Value", k)
		}
		ret[k] = sv
	}
	return ret, nil
}
