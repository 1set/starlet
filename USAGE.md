# Machine usage

These examples use the checked-out root library. Confirm the version selected by your application with `go list -m github.com/1set/starlet`; a default pkg.go.dev page may describe a tagged release. From this checkout, `go doc Machine`, `go doc RunFile`, and `go doc ./dataconv` show local API comments. See [Machine configuration](machine.go), [execution](run.go), and [error types](error.go) for the full signatures.

## Files, modules and Reset

`RunFile(name, fileSys, extras)` opens the exact `name` in the supplied `fs.FS`; `nil` does not mean the OS filesystem. Use `os.DirFS("scripts")` for a disk-backed root, or an in-memory filesystem as below. File `load()` paths are relative to that filesystem root, not the importing file's directory; a missing `.star` suffix is added for loaded modules, but not for the main `RunFile` name.

```go
package main

import (
    "fmt"
    "log"
    "testing/fstest"

    "github.com/1set/starlet"
)

func main() {
    files := fstest.MapFS{
        "main.star":     &fstest.MapFile{Data: []byte("load(\"lib/data\", \"value\")\nanswer = value")},
        "lib/data.star": &fstest.MapFile{Data: []byte("value = 1")},
    }
    m := starlet.NewDefault()
    run := func() {
        result, err := m.RunFile("main.star", files, nil)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Println(result["answer"])
    }
    run()
    files["lib/data.star"] = &fstest.MapFile{Data: []byte("value = 2")}
    run() // Loaded modules are retained by this Machine.
    m.Reset()
    run() // Reset discards loaded-module state.
}

// Output:
// 1
// 1
// 2
```

`SetScript(name, content, fileSys)` + `Run()` supports supplied source with the same file-load root. `RunScript` clears that filesystem and bypasses the compiled-program cache. `Reset` retains configuration and the optional compiled-program cache; it does not undo host writes or detach shared values. [Custom source-loader constructors](module.go) use bare threads and do not inherit the Machine's execution guards.

## Script output

`print(...)` goes to stderr by default. Install a `PrintFunc` to route it, use `NoopPrintFunc` to discard it, or pass `nil` to restore the default. Host callbacks must not reenter the same Machine's locking methods.

```go
package main

import (
    "fmt"
    "log"

    "github.com/1set/starlet"
    "go.starlark.net/starlark"
)

func main() {
    m := starlet.NewDefault()
    m.SetPrintFunc(func(_ *starlark.Thread, msg string) {
        fmt.Println("script:", msg)
    })
    if _, err := m.RunScript([]byte(`print("hello")`), nil); err != nil {
        log.Fatal(err)
    }
    m.SetPrintFunc(starlet.NoopPrintFunc)
    if _, err := m.RunScript([]byte(`print("hidden")`), nil); err != nil {
        log.Fatal(err)
    }
}

// Output: script: hello
```

## Errors and cancellation

Machine execution returns Go errors. `ExecError` unwraps its cause; use `errors.As` to classify the exported **value** error types, rather than matching message text:

```go
package main

import (
    "errors"
    "fmt"

    "github.com/1set/starlet"
)

func describe(err error) {
    var missing starlet.ModuleNotFoundError
    var withheld starlet.ModuleWithheldError
    var steps starlet.MaxStepsExceededError
    switch {
    case errors.As(err, &missing):
        fmt.Println("missing:", missing.Name)
    case errors.As(err, &withheld):
        fmt.Println("withheld:", withheld.Name)
    case errors.As(err, &steps):
        fmt.Println("step limit:", steps.Limit)
    default:
        fmt.Println(err)
    }
}

func main() {
    m := starlet.NewDefault()
    _, err := m.RunScript([]byte(`load("json", "decode")`), nil)
    describe(err) // Known builtin, not enabled on this Machine.
    _, err = m.RunScript([]byte(`load("missing_module", "value")`), nil)
    describe(err)
    m.SetMaxExecutionSteps(1)
    _, err = m.RunScript([]byte(`answer = [i * i for i in range(10)]`), nil)
    describe(err)
}

// Output:
// withheld: json
// missing: missing_module
// step limit: 1
```

Use `errors.As(err, &evalErr)` with `var evalErr *starlark.EvalError` to obtain a Starlark backtrace. `SetMaxExecutionSteps(0)` removes the step limit. The budget is per Run/Call; file-loaded module initialization gets a separate allowance.

`RunWithContext` uses source previously set with `SetScript`; `CallWithContext` invokes an already-defined function. An already-cancelled context fails before main-script execution; module preparation happens first:

```go
package main

import (
    "context"
    "errors"
    "fmt"

    "github.com/1set/starlet"
)

func main() {
    m := starlet.NewDefault()
    m.SetScript("calc.star", []byte("answer = 1 + 2"), nil)
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    _, err := m.RunWithContext(ctx, nil)
    fmt.Println(errors.Is(err, context.Canceled))
}

// Output: true
```

During execution, cancellation is cooperative and may return a Starlark evaluation error; inspect `ctx.Err()` alongside the returned error. It cannot preempt arbitrary Go callbacks. These Go errors differ from [`try_eval`'s `(result, error_info)`](lib/goidiomatic/README.md#expression-evaluation), which is a script-level result with documented fields.

## Shared values

Default conversion is not a detached copy of host data. Likewise, `shared_dict.to_dict()` copies top-level entries but retains nested aliases:

```python
load("go_idiomatic", "make_shared_dict")
original = make_shared_dict("", {"items": [1]})
copy = original.to_dict()
copy["new"] = 3
copy["items"].append(2)
print("new" in original, original["items"])
# Output: False [1, 2]
```

Own or synchronize mutable nested values separately. Default Starlark
formatting/printing and the Go `String`/`ToJSON` methods still require external
synchronization with mapping writers.

From Go, use `text := shared.SnapshotString()` to format a snapshot of the
mapping's entries without racing with mapping updates. Formatting callbacks run
after the mapping lock is released; mutable nested values remain shared.
Call `SnapshotString` without already holding that dictionary's mutex, and
outside callbacks from operations that hold it, such as key `Hash`/`Equal` or
value `Freeze`. This Go method does not add a Starlark builtin or change default
printing. See [shared_dict methods and callback rules](lib/goidiomatic/README.md#shared_dict).

## Terminal REPL

Call `starlet.NewDefault().REPL()` from a terminal Go program for an embedded read-eval-print loop. Configure modules first if needed. It owns the Machine for the session and returns no Go error; errors are printed by the REPL. On `js` and `wasip1` targets it is a no-op; use `Run` / `RunScript` there. The examples above do not establish an interactive-session or legacy-command runtime result. For command-line use, see [StarCLI](https://github.com/1set/starcli) or the [separate legacy command guide](cmd/starlet/README.md).
