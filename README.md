# :sparkles: Starlet - Starlark scripts and libraries for Go

[![godoc](https://pkg.go.dev/badge/github.com/1set/starlet.svg)](https://pkg.go.dev/github.com/1set/starlet)
[![codecov](https://codecov.io/github/1set/starlet/branch/master/graph/badge.svg?token=M1tauam4Hw)](https://codecov.io/github/1set/starlet)
[![codacy](https://app.codacy.com/project/badge/Grade/4e9c3f67a9574e6caa1b0d4706535815)](https://app.codacy.com/gh/1set/starlet/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![codeclimate](https://api.codeclimate.com/v1/badges/290ec0cc3261d16c423f/maintainability)](https://codeclimate.com/github/1set/starlet/maintainability)
[![goreportcard](https://goreportcard.com/badge/github.com/1set/starlet)](https://goreportcard.com/report/github.com/1set/starlet)

Starlet embeds [Starlark in Go](https://github.com/google/starlark-go) through a stateful [`Machine`](https://pkg.go.dev/github.com/1set/starlet#Machine), builtin libraries, and Go/Starlark value conversion. It uses the [Starlight bridge](https://github.com/1set/starlight) by default and also offers explicit native-value helpers in [`dataconv`](https://pkg.go.dev/github.com/1set/starlet/dataconv).

## Installation

From a Go module:

```sh
go get github.com/1set/starlet
```

This README describes the checked-out source; available APIs depend on the version selected by your application (`go list -m github.com/1set/starlet`). GoDoc links may show the latest tagged release. For this checkout, use `go doc Machine`, [source API comments](machine.go), and the [usage guide](USAGE.md).

The root library supports Go 1.19 or newer. Use a supported, security-patched toolchain for production; see [compatibility and security](#compatibility-and-security).

## Quick start

This complete program enables JSON loading, passes a Go function to the script, and reads a script result back in Go:

```go
package main

import (
    "fmt"
    "log"

    "github.com/1set/starlet"
)

func main() {
    globals := starlet.StringAnyMap{
        "greet": func(name string) string {
            return fmt.Sprintf("Hello, %s!", name)
        },
    }
    machine := starlet.NewWithNames(globals, nil, []string{"json"})
    result, err := machine.RunScript([]byte(`
load("json", "decode")
data = decode('{"name":"Starlark"}')
message = greet(data["name"])
`), nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result["message"])
}
// Output: Hello, Starlark!
```

## Choosing modules

`NewDefault()` supplies standard Starlark builtins, with no additional Starlet modules. Select modules explicitly:

| Constructor | Script access |
| --- | --- |
| `NewWithNames(globals, []string{"json"}, nil)` | Preload: call `json.decode(...)` directly. |
| `NewWithNames(globals, nil, []string{"json"})` | Lazy load: `load("json", "decode")`, then call `decode(...)`. |
| `NewWithBuiltins(globals, nil, nil)` | All builtin preloads and loaders, including filesystem, network and process effects. |

Preloading and enabling `load()` are separate choices. Most preloaded modules expose a named object; [`go_idiomatic`](lib/goidiomatic) exposes flat helpers such as `eval`, and `struct` exposes the `struct(...)` constructor. Binding precedence is host globals < preloads < execution extras < explicit `load` bindings.

Unknown builtin names make `NewWithNames` panic. For names supplied by callers, use `MakeBuiltinModuleLoaderList` / `MakeBuiltinModuleLoaderMap` and `NewWithLoaders` to handle configuration errors explicitly. Errors while initializing a module are returned by execution.

## Execution and state

- Configure globals, preloads and dialect options before the first execution; use `Reset()` when changing initialization configuration. Extras and script globals are retained for later executions and `Call`.
- `Reset()` discards runtime bindings and loaded-module state, while retaining configuration and the optional compiled-program cache. It does not undo host-side effects.
- `RunScript` executes supplied bytes without the compiled-program cache. For reusable named scripts, use `SetScript` + `Run`, or `RunFile`; enable caching with `SetScriptCacheEnabled(true)` when needed.
- Route script `print(...)` with `SetPrintFunc`; `nil` uses stderr and `NoopPrintFunc` discards output. See [file execution, output, and errors](USAGE.md).
- After loading function definitions, use `Call`, `CallWithContext`, or `CallWithTimeout` to invoke them from Go. Context cancellation is cooperative; it cannot preempt arbitrary Go callbacks.
- `SetMaxExecutionSteps` bounds VM steps; `0` means unlimited. The counter resets for each Run/Call, and file-loaded module initialization receives a separate thread allowance. This is not a wall-clock or memory limit.

Execution on one Machine is serialized. Host callbacks must not reenter methods on the same Machine, except the nonblocking `String` method. The lock does not synchronize externally shared Go data or raw handles returned by `GetStarlarkThread` / `GetStarlarkPredeclared`; the host must own or synchronize those values.

## Value conversion

Global/extra input conversion and result conversion are enabled by default. Go values use Starlight wrappers where appropriate; results are converted back to Go values. This does not promise a detached copy of host data. With input conversion disabled, globals and extras must already implement `starlark.Value`; `Call` arguments still use the Starlight bridge. Disabling output conversion retains `starlark.Value` results.

Use [`dataconv.Marshal` / `Unmarshal`](https://pkg.go.dev/github.com/1set/starlet/dataconv) for explicit conversion of supported common Go/native Starlark types. These helpers are separate from the Machine's default bridge.

## Libraries

Each module README lists its functions, parameters, results and boundaries:

| Package                           | Go Doc                                                                                                                                       | Description                                                   |
|:----------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------|:--------------------------------------------------------------|
| [`atom`](/lib/atom)               | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/atom.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/atom)               | Atomic operations for integers, floats, and strings           |
| [`base64`](/lib/base64)           | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/base64.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/base64)           | Base64 encoding & decoding functions                          |
| [`csv`](/lib/csv)                 | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/csv.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/csv)                 | Parses and writes comma-separated values (csv) contents       |
| [`file`](/lib/file)               | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/file.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/file)               | Functions to interact with the file system                    |
| [`goidiomatic`](/lib/goidiomatic) | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/goidiomatic.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/goidiomatic) | Go-style helpers, values and constructors; read-only expression evaluation                |
| [`hashlib`](/lib/hashlib)         | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/hashlib.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/hashlib)         | Hash primitives for Starlark                                  |
| [`http`](/lib/http)               | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/http.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/http)               | HTTP client and server handler implementation for Starlark    |
| [`json`](/lib/json)               | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/json.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/json)               | Utilities for converting Starlark values to/from JSON strings |
| [`log`](/lib/log)                 | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/log.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/log)                 | Functionality for logging messages at various severity levels |
| [`net`](/lib/net)                 | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/net.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/net)                 | Network reachability checks: TCP/HTTP ping and DNS lookup     |
| [`path`](/lib/path)               | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/path.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/path)               | Functions to manipulate directories and file paths            |
| [`random`](/lib/random)           | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/random.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/random)           | Functions to generate random values for various distributions |
| [`re`](/lib/re)                   | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/re.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/re)                   | Regular expression functions for Starlark (legacy, frozen)    |
| [`regex`](/lib/regex)             | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/regex.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/regex)             | Python-style regex (RE2) with Match objects, flags, named groups |
| [`runtime`](/lib/runtime)         | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/runtime.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/runtime)         | Provides Go and app runtime information                       |
| [`serial`](/lib/serial)           | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/serial.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/serial)           | Lossless data-value serialization (bytes/set/tuple/bigint/time) |
| [`stats`](/lib/stats)             | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/stats.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/stats)             | Statistics: mean/median/stddev, correlation, percentiles, etc. |
| [`string`](/lib/string)           | [![godoc](https://pkg.go.dev/badge/github.com/1set/starlet/lib/string.svg)](https://pkg.go.dev/github.com/1set/starlet/lib/string)           | Constants and functions to manipulate strings                 |


`GetAllBuiltinModuleNames()` returns the complete sorted registry. The upstream [`math`](https://github.com/google/starlark-go/blob/ffb3f39dd27a/lib/math/math.go), [`struct`](https://github.com/google/starlark-go/blob/ffb3f39dd27a/starlarkstruct/struct.go), and [`time`](https://github.com/google/starlark-go/blob/ffb3f39dd27a/lib/time/time.go) APIs are also available; they have no `lib/` README here. [Capability metadata](https://pkg.go.dev/github.com/1set/starlet#ModuleCapability) describes builtin host effects for policy consumers; it is not process isolation or a determinism guarantee.

## Compatibility and security

The root library keeps the interpreter pin at `go.starlark.net v0.0.0-20260324133313-ffb3f39dd27a`. Its Go 1.19 compatibility floor is separate from the production compiler.

This interpreter predates the upstream parser recursion limit. Only parse host-selected, reviewed source and modules in the host process. Step limits, timeouts and module selection do not make arbitrary untrusted source safe. Read [SECURITY.md](SECURITY.md) for the deployment boundary.

## Command-line tools

For the dedicated command-line application, see [StarCLI](https://github.com/1set/starcli). The older [`cmd/starlet`](cmd/starlet/README.md) has separate dependency pins and [legacy install/REPL instructions](cmd/starlet/README.md); installing it does not automatically use the root library version described here. For an embedded terminal REPL, see [`Machine.REPL`](USAGE.md#terminal-repl).

## Contributing and license

Issues and pull requests are welcome. Please discuss significant changes before implementation to avoid duplicate work. Starlet is licensed under [MIT](LICENSE).

## Credits

Inspired by [Starlight](https://github.com/starlight-go/starlight) and [Starlib](https://github.com/qri-io/starlib).
