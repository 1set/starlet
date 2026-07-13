package starlet

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"math"
	"sync"
	"time"

	"github.com/1set/starlet/lib/goidiomatic"
	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// REPL is defined in run_repl.go (terminal targets) / run_repl_stub.go
// (non-terminal targets): the
// interactive REPL pulls go.starlark.net/repl -> chzyer/readline, a terminal
// library that does not compile for browser js/wasm or WASI. Isolating it
// behind a build tag keeps the library core (and every consumer, e.g. a WASM
// playground) free of that terminal dependency.

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

	return m.runInternal(context.Background(), nil, true)
}

// RunScript executes a script with additional variables, which take precedence over global variables and modules, returns the result.
func (m *Machine) RunScript(content []byte, extras StringAnyMap) (StringAnyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.scriptName = "direct.star"
	m.scriptContent = content
	m.scriptFS = nil
	return m.runInternal(context.Background(), extras, false)
}

// RunFile executes a script from a file with additional variables, which take precedence over global variables and modules, returns the result.
func (m *Machine) RunFile(name string, fileSys fs.FS, extras StringAnyMap) (StringAnyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.scriptName = name
	m.scriptContent = nil
	m.scriptFS = fileSys
	return m.runInternal(context.Background(), extras, true)
}

// RunWithTimeout executes a preset script with a timeout and additional variables, which take precedence over global variables and modules, returns the result.
func (m *Machine) RunWithTimeout(timeout time.Duration, extras StringAnyMap) (StringAnyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return m.runInternal(ctx, extras, true)
}

// RunWithContext executes a preset script within a specified context and additional variables, which take precedence over global variables and modules, returns the result.
func (m *Machine) RunWithContext(ctx context.Context, extras StringAnyMap) (StringAnyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.runInternal(ctx, extras, true)
}

// watchContextCancel cancels the machine's thread when ctx fires, until the
// returned stop function is called. stop is idempotent and waits for the
// watcher goroutine to exit, so callers can both defer it (panic safety)
// and invoke it right after execution finishes.
func (m *Machine) watchContextCancel(ctx context.Context) (stop func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			m.thread.Cancel("context cancelled")
		case <-done:
			// no action if execution has finished
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			close(done)
			wg.Wait()
		})
	}
}

func (m *Machine) runInternal(ctx context.Context, extras StringAnyMap, allowCache bool) (out StringAnyMap, err error) {
	defer func() {
		if r := recover(); r != nil {
			if me, ok := r.(MaxStepsExceededError); ok {
				err = errorStarlarkError("exec", me)
			} else {
				err = errorStarlarkPanic("exec", r)
			}
		}
	}()

	// either script content or name and FS must be set
	var (
		scriptName = m.scriptName
		source     interface{}
	)
	if m.scriptContent != nil {
		if scriptName == "" {
			// for default name, and disable cache to avoid conflict
			scriptName = "eval.star"
			allowCache = false
		}
		source = m.scriptContent
	} else if m.scriptFS != nil {
		if scriptName == "" {
			// if no name, cannot load
			return nil, errorStarletErrorf("run", "no script name")
		}
		// load the script content from FS, so that the program cache can
		// key on the content (passing the open reader through degraded the
		// cache key to the bare filename, letting different files with the
		// same name hit each other's compiled program) — and the file
		// handle was never closed
		rd, e := m.scriptFS.Open(scriptName)
		if e != nil {
			return nil, errorStarletError("run", e)
		}
		b, e := io.ReadAll(rd)
		_ = rd.Close()
		if e != nil {
			return nil, errorStarletError("run", e)
		}
		source = b
	} else {
		return nil, errorStarletErrorf("run", "no script to execute")
	}

	// prepare thread
	if err = m.prepareThread(extras); err != nil {
		return nil, err
	}

	// cancel thread when context cancelled
	if ctx == nil {
		// no context given: use an inert placeholder
		ctx = context.TODO()
	} else if e := ctx.Err(); e != nil {
		// an already-cancelled context used to be silently replaced with an
		// uncancellable one, running the script with no deadline at all —
		// fail fast instead
		return nil, errorStarletError("run", e)
	}
	m.thread.SetLocal("context", ctx)

	// cancel the thread when the context fires, until execution finishes
	stop := m.watchContextCancel(ctx)
	defer stop()

	// run with everything prepared
	m.runTimes++
	res, err := m.execStarlarkFile(scriptName, source, allowCache)
	stop()

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
	mergeExtra := func() error {
		// no extras
		if extras == nil {
			return nil
		}
		// convert extras if needed
		esd, err := m.convertInput(extras)
		if err != nil {
			return errorStarlightConvert("extras", err)
		}
		// merge extras
		for k, v := range esd {
			m.predeclared[k] = v
		}
		return nil
	}

	// initialize thread or reset for each run
	if m.thread == nil {
		// -- for the first run

		// preset globals + preload modules + extras -> predeclared
		if m.predeclared, err = m.convertInput(m.globals); err != nil {
			return errorStarlightConvert("globals", err)
		}
		if err = m.preloadMods.LoadAll(m.predeclared); err != nil {
			return errorStarletError("preload", err)
		}

		// merge extras into predeclared
		if err = mergeExtra(); err != nil {
			return err
		}

		// cache load&read + printf -> thread
		m.loadCache = &cache{
			cache:    make(map[string]*entry),
			execOpts: m.getFileOptions(),
			loadMod:  m.lazyloadMods.GetLazyLoader(),
			readFile: func(name string) ([]byte, error) {
				return readScriptFile(name, m.scriptFS)
			},
			globals:   m.predeclared,
			newThread: m.newLoadThread,
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

		// merge extras into predeclared
		if err = mergeExtra(); err != nil {
			return err
		}

		// set globals for cache
		m.loadCache.loadMod = m.lazyloadMods.GetLazyLoader()
		m.loadCache.globals = m.predeclared

		// reset for each run
		m.thread.Print = m.printFunc
		m.thread.Uncancel()
	}

	// arm the per-run step budget
	m.applyStepBudget()
	return nil
}

// newLoadThread builds the thread that runs a module executed by load(),
// mirroring the main thread's execution context: the same print func, an
// independent copy of the step budget (so a loaded module's work is bounded
// by the DoS guard instead of escaping it), and the current run's context
// local. The step budget is per-thread, not a shared aggregate counter, so a
// loaded module gets its own MaxSteps allowance — enough to stop a runaway
// loop, which is the DoS the bare thread let through.
func (m *Machine) newLoadThread(load func(*starlark.Thread, string) (starlark.StringDict, error)) *starlark.Thread {
	t := &starlark.Thread{
		Name:  "starlet:load",
		Print: m.printFunc,
		Load:  load,
	}
	limit := m.maxSteps
	if limit == 0 {
		limit = math.MaxUint64
	}
	t.SetMaxExecutionSteps(limit)
	if lim := m.maxSteps; lim > 0 {
		t.OnMaxSteps = func(*starlark.Thread) {
			// recovered by the run/call recover and mapped to a typed error
			panic(MaxStepsExceededError{Limit: lim})
		}
	}
	if m.thread != nil {
		if ctx := m.thread.Local("context"); ctx != nil {
			t.SetLocal("context", ctx)
		}
	}
	return t
}

// applyStepBudget arms the thread with the configured step budget; it must
// run before every execution. The Starlark runtime normalizes a zero limit
// to "unlimited" only on a thread's very first use, so the translation to
// MaxUint64 happens here for reused threads; the step counter resets so
// the budget applies per execution.
func (m *Machine) applyStepBudget() {
	limit := m.maxSteps
	if limit == 0 {
		limit = math.MaxUint64
	}
	m.thread.SetMaxExecutionSteps(limit)
	if lim := m.maxSteps; lim > 0 {
		m.thread.OnMaxSteps = func(*starlark.Thread) {
			// recovered by the run/call recover and mapped to a typed error
			panic(MaxStepsExceededError{Limit: lim})
		}
	} else {
		m.thread.OnMaxSteps = nil
	}
	m.thread.Steps = 0
}

// Reset resets the machine to initial state before the first run.
// Attention: It does not reset the compiled program cache.
//
// It takes the write lock like the other mutators: it clears the very fields
// (thread, runTimes, predeclared) that Run and the accessors read, so an
// unlocked Reset would race with them however carefully the readers lock.
func (m *Machine) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

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
