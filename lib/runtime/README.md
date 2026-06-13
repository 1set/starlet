# runtime

`runtime` is a Starlark module that exposes Go and application runtime information — host, paths, OS/architecture, Go version, process IDs, app start time/uptime, and read/write access to environment variables.

Capability profile: **Process**. The module reports host/process facts captured at load time and can read and mutate the process's environment (`getenv` / `putenv` / `setenv` / `unsetenv`); it does not touch the filesystem or network.

## Functions

| function | description |
| --- | --- |
| `uptime() -> time.duration` | Time elapsed since the application started. |
| `getenv(key, default=None) -> string` | Value of environment variable `key`, or `default` if it is not set. |
| `putenv(key, value) -> None` | Set environment variable `key` to `value` (coerced to a string). `setenv` is an alias. |
| `setenv(key, value) -> None` | Alias of `putenv`. |
| `unsetenv(key) -> None` | Unset a single environment variable. |

## Constants

Constants are captured once when the module is first loaded.

| constant | meaning |
| --- | --- |
| `hostname` | `string` — the host name of the system (Go `os.Hostname()`). |
| `workdir` | `string` — the current working directory of the process (Go `os.Getwd()`). |
| `homedir` | `string` — the home directory of the user: `$HOME` on Unix/Linux, `%USERPROFILE%` on Windows (Go `os.UserHomeDir()`). |
| `tempdir` | `string` — the default directory for temporary files, like Python's `tempfile.gettempdir()` (Go `os.TempDir()`). |
| `os` | `string` — the operating system, from Go `runtime.GOOS` (e.g. `linux`, `darwin`, `windows`). |
| `arch` | `string` — the machine architecture, from Go `runtime.GOARCH` (e.g. `amd64`, `arm64`). |
| `gover` | `string` — the Go runtime version, from Go `runtime.Version()` (e.g. `go1.19`). |
| `pid` | `int` — the process ID of the current process. |
| `ppid` | `int` — the parent process ID. |
| `uid` | `int` — the user ID of the process owner. |
| `gid` | `int` — the group ID of the process owner. |
| `app_start` | `time.time` — the moment the application started; used to compute `uptime`. |

## Details & examples

### `uptime`

`uptime() -> time.duration`

Returns the elapsed time since the application started (relative to `app_start`) as a `time.duration`. Takes no arguments; passing any argument errors with `runtime.uptime: got 1 arguments, want 0`.

The exact value depends on how long the process has been running, so the output below is illustrative.

```python
load("runtime", "uptime")
print(uptime())
# Output: 883.583µs
```

### `getenv`

`getenv(key, default=None) -> string`

Returns the value of environment variable `key` as a `string` if it is set, otherwise returns `default` (which defaults to `None` and may be any value). `key` must be a string; a missing or non-string `key` errors (`runtime.getenv: missing argument for key`, `runtime.getenv: for parameter key: got int, want string`).

```python
load("runtime", "getenv")
x = getenv("very-long-long-non-existent")
print(x)
y = getenv("very-long-long-non-existent", 1000)
print(y)
# Output: None
# 1000
```

### `putenv` / `setenv`

`putenv(key, value) -> None` — `setenv(key, value) -> None` is an identical alias.

Sets environment variable `key` to `value`. `value` is coerced to a string before being stored, so non-string values become their string form (e.g. the int `123456` is stored as `"123456"`). Returns `None`. `key` must be a string and both arguments are required; otherwise it errors (`runtime.putenv: missing argument for key`, `runtime.putenv: missing argument for value`, `runtime.putenv: for parameter key: got int, want string`).

```python
load("runtime", "putenv", "getenv")
putenv("STARLET_TEST", 123456)
print(getenv("STARLET_TEST"))
# Output: 123456
```

`setenv` behaves the same way:

```python
load("runtime", "setenv", "getenv")
setenv("STARLET_TEST", 123456)
print(getenv("STARLET_TEST"))
# Output: 123456
```

### `unsetenv`

`unsetenv(key) -> None`

Unsets a single environment variable. Returns `None`. Unsetting a variable that does not exist is a no-op (not an error). `key` must be a string and is required; otherwise it errors (`runtime.unsetenv: missing argument for key`, `runtime.unsetenv: for parameter key: got int, want string`).

```python
load("runtime", "putenv", "unsetenv", "getenv")
putenv("STARLET_TEST", 123456)
unsetenv("STARLET_TEST")
print(getenv("STARLET_TEST"))
# Output: None
```

## Notes / boundaries

- **Capture timing.** The constants (`hostname`, `workdir`, `homedir`, `tempdir`, `os`, `arch`, `gover`, `pid`, `ppid`, `uid`, `gid`, `app_start`) are read when the module is first loaded and do not refresh afterwards. `uptime` is computed live at call time against `app_start`.
- **No custom types.** All members are native Starlark values: strings, ints, a `time.time` (`app_start`), and a `time.duration` returned by `uptime`. The `time.*` values come from `go.starlark.net/lib/time`.
- **Environment writes are global.** `putenv`/`setenv`/`unsetenv` mutate the host process environment, not a sandboxed copy; effects are visible to the rest of the process and to child processes.
- **Platform differences.** `os`, `arch`, `gover`, `homedir`, and the numeric IDs reflect the underlying platform; on Windows, `uid`/`gid` follow Go's `os.Getuid()`/`os.Getgid()` semantics (which may be `-1`).
- All exported names are snake_case.
