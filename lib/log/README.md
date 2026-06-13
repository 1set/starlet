# log

`log` writes log messages at five severity levels from Starlark, backed by a [`zap`](https://github.com/uber-go/zap) sugared logger on the host side. Capability profile: **Log** — it produces log output as a side effect but touches no filesystem, network, or process state directly.

By default the module logs through a `zap` development logger (console encoder, caller and stacktrace disabled). The host can swap the logger at runtime with `SetLog` (Go side), including a no-op logger that discards everything.

## Functions

| function | description |
|----------|-------------|
| `debug(msg, *misc, **kv) -> None` | Log `msg` at DEBUG level; returns `None`. |
| `info(msg, *misc, **kv) -> None` | Log `msg` at INFO level; returns `None`. |
| `warn(msg, *misc, **kv) -> None` | Log `msg` at WARN level; returns `None`. |
| `error(msg, *misc, **kv) -> None` | Log `msg` at ERROR level; returns `None` (does **not** halt the script). |
| `fatal(msg, *misc, **kv) -> error` | Log `msg` at ERROR level, then raise the message as an error that halts the script. |

All five functions share the same calling shape: a required string `msg`, optional extra positional arguments (`*misc`) appended to the message, and optional keyword arguments (`**kv`) attached as structured key-value fields.

This module exposes **no constants** and **no custom types** — its five members are all builtin callables on the `log` module struct.

## Details & examples

### Common argument handling

Every function takes the same arguments:

- `msg` (required, `string`) — the log message. It must be the first positional argument and must be a string.
- `*misc` (optional) — any further positional arguments. Each is rendered to text and appended to `msg`, separated by single spaces. Booleans render as `True`/`False`, numbers and strings render naturally.
- `**kv` (optional) — keyword arguments become structured fields. Keys are interpreted as strings; values are unmarshaled to Go types where possible (so dicts/lists/numbers serialize as JSON-like structures, `None` becomes `null`), falling back to the value's `String()` form for self-referential or non-marshalable values.

**Errors** (identical for all five functions):

- Calling with no arguments fails: `log.<name>: expected at least 1 argument, got 0`.
- A non-string first argument fails: `log.<name>: expected string as first argument, got <type>` (e.g. `got int`).

Note the error prefix uses the qualified builtin name, e.g. `log.debug`, `log.fatal`.

### `debug`

```python
load('log', 'debug')
debug('this is a debug message only')
# Output:
# DEBUG	this is a debug message only
```

Extra positional arguments are concatenated onto the message:

```python
load('log', 'debug')
debug('this is a broken message', "what", 123, True)
# Output:
# DEBUG	this is a broken message what 123 True
```

Keyword arguments are attached as structured fields:

```python
load('log', 'debug')
m = {"mm": "this is more"}
l = [2, "LIST", 3.14, True]
debug('this is a data message', map=m, list=l)
# Output:
# DEBUG	this is a data message	{"map": {"mm":"this is more"}, "list": [2,"LIST",3.14,true]}
```

### `info`

```python
load('log', 'info')
info('this is an info message', a1=2, hello="world")
# Output:
# INFO	this is an info message	{"a1": 2, "hello": "world"}
```

Self-referential values fall back to their string form rather than failing:

```python
load('log', 'info')
d = {"hello": "world"}
d["a"] = d
l = [1, 2, 3]
l.append(l)
s = set([4, 5, 6])
info('this is complex info message', self1=d, self2=l, self3=s)
# Output:
# INFO	this is complex info message	{"self1": "{\"hello\": \"world\", \"a\": {...}}", "self2": "[1, 2, 3, [...]]", "self3": [4,5,6]}
```

### `warn`

```python
load('log', 'warn')
warn('this is a warning message only')
# Output:
# WARN	this is a warning message only
```

### `error`

`error` logs at ERROR level and returns `None` — it does not stop execution. Use `fatal` (or Starlark's `fail`) to halt.

```python
load('log', 'error')
error('this is an error message only', dsat=None)
# Output:
# ERROR	this is an error message only	{"dsat": null}
```

### `fatal`

`fatal` logs the message at ERROR level and then raises it as an error, halting the script (the message becomes the error string). It does **not** call `os.Exit`; the host receives a normal Starlark error.

```python
load('log', 'fatal')
fatal('this is a fatal message only')
# Output:
# this is a fatal message only
```

(The line above is the error raised to the host; an ERROR-level log entry with the same message is also written before the error returns.)

## Notes / boundaries

- **Engine.** Logging is delegated to a `go.uber.org/zap` `SugaredLogger`. The exact output format (level token casing, field separators, JSON shape of structured values) depends on the configured zap encoder. The examples above show the console-encoder form used by the test suite; the default development logger emits a similar human-readable line. Timestamps and any caller/stacktrace fields are omitted here for brevity.
- **Levels.** Five levels are exposed: `debug`, `info`, `warn`, `error`, `fatal`. Internally `fatal` logs at zap's ERROR level (not zap's FatalLevel) and then returns an error — it never terminates the process.
- **`error` vs `fatal`.** `error` records and continues; `fatal` records and aborts the script. Both write at ERROR severity.
- **No-op mode.** When the host installs a nil/no-op logger via `SetLog`, all functions still validate arguments and return normally (and `fatal` still raises its error), but no output is produced.
- **Determinism.** Output content is deterministic for given inputs; structured-field ordering follows the order of keyword arguments. Whether a line is emitted at all depends on the configured logger's level threshold.
- All member names are snake_case-clean (single lowercase words); there are no irregular identifiers.
