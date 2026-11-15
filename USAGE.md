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

## Calling script functions

First execute the function definitions successfully, then use `Call(name, args...)` to pass positional Go arguments and receive `(result, error)`:

```go
package main

import (
    "fmt"
    "log"

    "github.com/1set/starlet"
)

func main() {
    m := starlet.NewDefault()
    if _, err := m.RunScript([]byte("def add(x, y):\n    return x + y\n"), nil); err != nil {
        log.Fatal(err)
    }
    result, err := m.Call("add", 20, 22)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result)
}

// Output: 42
```

`Call` takes positional arguments only, converting them through Starlight even when global/extra input conversion is disabled. Results use the configured output conversion. Calling before execution, using a missing name, or selecting a value other than a script function or builtin returns an error. Use `CallWithContext(ctx, name, args...)` or `CallWithTimeout(timeout, name, args...)` for cooperative cancellation; plain `Call` has no context or deadline. See [errors and cancellation](#errors-and-cancellation).

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

### Print byte budget

`m.SetMaxPrintBytes(64 * 1024)` limits bytes delivered through the Machine's print handler. The default `0` is unlimited. Each Run or Call gets a fresh allowance; a REPL session shares one allowance across its inputs. Machine-loaded source modules consume the caller's allowance. `Reset` preserves the configured limit.

The limit counts UTF-8 message bytes before host formatting or added newlines. A message that would exceed it is withheld in full, and Run/Call stop with the value error `PrintLimitExceededError`; use `errors.As` as shown below. An exceeded REPL allowance ends the session and reports the error.

The limit does not bound loggers, direct stderr writes such as `eprint()`, returned values, message construction, or trusted source-loader helpers using separate threads. Those channels need separate host limits. A nonzero budget adds allocation and synchronized accounting work; configuring it can allocate before the first execution. Costs depend on the compiler, machine and workload.

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
    var printLimit starlet.PrintLimitExceededError
    switch {
    case errors.As(err, &missing):
        fmt.Println("missing:", missing.Name)
    case errors.As(err, &withheld):
        fmt.Println("withheld:", withheld.Name)
    case errors.As(err, &steps):
        fmt.Println("step limit:", steps.Limit)
    case errors.As(err, &printLimit):
        fmt.Println("print byte limit:", printLimit.Limit)
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
    m.SetMaxExecutionSteps(0)
    m.SetMaxPrintBytes(2)
    _, err = m.RunScript([]byte(`print("too long")`), nil)
    describe(err)
}

// Output:
// withheld: json
// missing: missing_module
// step limit: 1
// print byte limit: 2
```

Use `errors.As(err, &evalErr)` with `var evalErr *starlark.EvalError` to obtain a Starlark backtrace. `SetMaxExecutionSteps(0)` removes the step limit. The step budget is per Run/Call; file-loaded module initialization gets a separate allowance.

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

`String()` (including Starlark `print`/`str`/`repr`) and `SnapshotString()` now capture mapping entries under an internal lock, then release that lock before formatting callbacks. Concurrent output operations are serialized without waiting: if an output or mapping lock is already held, they return the fixed marker `shared_dict(<busy>)`. This also applies to nested SharedDict values and reentrant calls. Normal output retains insertion order and configured type names. `SetTypeName` is safe to call concurrently.

Go `ToJSON() (string, error)` and Starlark `to_json()` use the same nonblocking serialization path. Go callers can check `errors.Is(err, dataconv.ErrSharedDictBusy)`; a failed call returns no JSON. Retry after leaving the current lock or callback, rather than retrying in a loop that still holds it.

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/1set/starlet/dataconv"
	"go.starlark.net/starlark"
)

func main() {
	shared := dataconv.NewSharedDict()
	if err := shared.SetKey(starlark.String("count"), starlark.MakeInt(1)); err != nil {
		log.Fatal(err)
	}
	text, err := shared.ToJSON()
	if errors.Is(err, dataconv.ErrSharedDictBusy) {
		log.Print("shared output is busy; retry outside the current critical section")
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(text)
	// Output: {"count":1}
}
```

SharedDict JSON accepts native JSON scalars and containers, native structs/modules, duration, and the interpreter's immutable range/string/bytes iterator sources. Dictionary keys must be strings. Unsupported values and native container cycles return errors. **Other Go-backed values, including `GoStruct`, `GoMap`, `GoSlice` and custom JSON/attribute implementations, are rejected by SharedDict JSON.** Convert such values to independently owned plain data before storing or encoding them. The general `EncodeStarlarkJSON` helper retains its existing behavior, including host callbacks, and requires the caller to own or synchronize its input.

Internal output locks do not take ownership of raw nested List/Dict/Set aliases or native data shared by independent SharedDict owners. Keep those values under one owner or a common application mutex for every access. Use that same boundary for whole Machine `Run`/`Call` operations and compound updates that need an atomic application snapshot. Formatting callbacks must not reenter an application mutex or wait for its owner while that boundary is held. `to_dict()`/`CloneDict()` are shallow copies; freezing the original changes that original rather than creating an isolated copy.

`SnapshotString` remains a Go-only name for the same formatter used by default output. Cycle detection follows list/dict/set references through tuples, nested SharedDict values and structs with unique field names. Formatting callbacks run after mapping locks are released; nested references reuse captured owner views, while reentrant output calls return a busy marker. Custom values and Go-created structs with duplicate field names retain their own `String` behavior. Go-created tuple-only cycles and general depth/output limits remain outside this formatter's guarantees. Other dictionary operations retain their ordinary mutex discipline.

Output locking, entry snapshots, cycle paths and JSON validation add work and temporary memory. Paired benchmarks observed higher costs for set formatting and JSON encoding; ordinary `Get`/`SetKey` added no allocations. Setting a nonempty type name also allocates. Results depend on data, compiler, machine and load; avoid unnecessary formatting in business hot loops. Run the included read and output benchmarks with:

```bash
go test ./dataconv -run '^$' -bench '^BenchmarkSharedDict(ReadPaths|OutputSafety)$' -benchmem
```

## Terminal REPL

Call `starlet.NewDefault().REPL()` from a terminal Go program for an embedded read-eval-print loop. Configure modules first if needed. It owns the Machine for the session and returns no Go error; errors are printed by the REPL. On `js` and `wasip1` targets it is a no-op; use `Run` / `RunScript` there. The examples above do not establish an interactive-session or legacy-command runtime result. For command-line use, see [StarCLI](https://github.com/1set/starcli) or the [separate legacy command guide](cmd/starlet/README.md).
