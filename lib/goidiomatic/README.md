# go_idiomatic

`go_idiomatic` provides Go-flavored helpers, constants, and constructors for Starlark scripts — base conversions, `length`/`sum`/`distinct` utilities, struct and shared-dict factories, data-only expression evaluation, and `print`-style output. It is **not** a mirror of any single Python module; it borrows the spirit of several Python builtins (`hex`/`oct`/`bin`/`sum`/`bytes.hex`) while adding Star\* idioms.

Capability profile: **Log + Process**. `eprint`/`pprint` write to stderr or the thread's `Print` handler (Log); `sleep` blocks the thread and `exit`/`quit` halt the program (Process). Everything else is pure.

Load name for `load()`: `go_idiomatic`.

## Functions

| function | description |
|----------|-------------|
| `eval(source, vars={}) -> value` | Evaluate one Starlark expression against copied, read-only native data. |
| `try_eval(source, vars={}) -> (value, error_info)` | Return ordinary evaluation errors as frozen data; preserve caller cancellation and total step limits. |
| `length(obj) -> int` | Length of a string (in Unicode code points), bytes, or any sequence. |
| `sum(iterable, start=0) -> number` | Sum of `start` plus the items of `iterable`, left to right. |
| `distinct(iterable) -> iterable` | Iterable with duplicates removed; shape follows the input type. |
| `hex(x) -> str` | Lowercase hexadecimal string of an int, prefixed `0x`. |
| `oct(x) -> str` | Octal string of an int, prefixed `0o`. |
| `bin(x) -> str` | Binary string of an int, prefixed `0b`. |
| `bytes_hex(bytes, sep="", bytes_per_sep=1) -> str` | Hex string of each byte, with an optional grouping separator. |
| `is_nil(x) -> bool` | Whether `x` is `nil`/`None` or a Go wrapper holding a nil value. |
| `sleep(secs)` | Block the current thread for `secs` seconds (cancellable). |
| `exit(code=0)` / `quit(code=0)` | Halt the program with an exit code; `quit` is an alias of `exit`. |
| `module(name, **kv) -> module` | Build a `starlarkstruct` module named `name`. |
| `struct(**kv) -> struct` | Build an anonymous comparable struct. |
| `make_struct(name, **kv) -> struct` | Build a struct whose constructor name is `name`. |
| `shared_dict() -> shared_dict` | Create an empty thread-safe shared dictionary. |
| `make_shared_dict(name="", data=None) -> shared_dict` | Create a shared dictionary with an optional type name and initial data. |
| `to_dict(v) -> dict` | Convert a dict, `module`, `struct`, `GoStruct`, or `shared_dict` into a plain dict. |
| `eprint(*args, sep=" ")` | `print`-style output to stderr. |
| `pprint(*args, sep=" ")` | `print`-style output formatted as indented JSON. |

## Constants

| constant | meaning |
|----------|---------|
| `true` | Alias for the Starlark boolean `True`. |
| `false` | Alias for the Starlark boolean `False`. |
| `nil` | Alias for `None`. |

```python
load("go_idiomatic", "true", "false", "nil")
print(true, false, nil)
# Output: True False None
```

## Types

### `shared_dict`

A thread-safe, mutable dictionary returned by `shared_dict()` and `make_shared_dict()`. Every read and write is locked, so multiple Starlark threads may share and mutate one instance without races. Its `type()` is `shared_dict` by default, or the custom name passed to `make_shared_dict`. It supports indexing (`sd["k"] = v`, `sd["k"]`), membership (`in`), the `.len()` method, and `==`/`!=` comparison with another `shared_dict`. It is **not** iterable, and the builtin `len()` does not apply — use the `.len()` method.

Beyond the standard dict methods it inherits (`clear`, `get`, `items`, `keys`, `pop`, `popitem`, `setdefault`, `update`, `values`), it adds the methods below.

| method | description |
|--------|-------------|
| `len() -> int` | Number of items in the dictionary. |
| `perform(fn)` | Call `fn(self)` while holding the lock, for atomic compound updates. |
| `to_dict() -> dict` | Shallow clone into a plain dict; mutating the clone never affects the original. |
| `to_json() -> str` | Serialize the contents to a JSON string. |
| `from_json(json_str)` | Decode a JSON object string and merge its pairs into the dictionary. |

```python
load("go_idiomatic", "make_shared_dict")
sd = make_shared_dict("mydict", {"a": 1, "b": 2})
print(type(sd), sd.len())
# Output: mydict 2
```

`perform` runs the callback under the dict's lock so a read-modify-write stays atomic:

```python
load("go_idiomatic", "make_shared_dict")
sd = make_shared_dict()
def bump(d): d["cnt"] = d.get("cnt", 0) + 1
sd.perform(bump)
sd.perform(bump)
print(sd)
# Output: shared_dict({"cnt": 2})
```

`to_dict` clones; mutating the clone leaves the source empty:

```python
load("go_idiomatic", "make_shared_dict")
sd = make_shared_dict()
clone = sd.to_dict()
clone["k"] = "v"
print(sd, clone)
# Output: shared_dict({}) {"k": "v"}
```

`to_json` serializes and `from_json` merges:

```python
load("go_idiomatic", "make_shared_dict")
sd = make_shared_dict()
sd.from_json('{"new_key": "new_value"}')
print(sd.to_json())
# Output: {"new_key":"new_value"}
```

## Details & examples

### `length(obj) -> int`

For a string, returns the number of Unicode code points (unlike the builtin `len()`, which counts UTF-8 bytes); for bytes, the byte count; otherwise the `Len()` of any `starlark.Sequence` (list, tuple, dict, set, and `starlight` slices/maps). Errors with `length() takes exactly one argument` if not given exactly one positional argument, and `length() function isn't supported for '<type>' type object` for a non-sized type (e.g. `bool`).

```python
load("go_idiomatic", "length")
print(length("水光肌"), len("水光肌"))
# Output: 3 9
```

```python
load("go_idiomatic", "length")
print(length([1, 2, 3]), length(set(["a", "b"])), length({"a": 1, "b": 2}))
# Output: 3 2 2
```

### `sum(iterable, start=0) -> number`

Adds every item of `iterable` to `start`. Items and `start` must be numbers; `None` items are skipped (treated as zero). Mixing int and float yields an int when the result is whole. Errors with `unsupported type: <type>, expected float or int` on a non-numeric item, and `got <type>, want iterable` if `iterable` is not iterable.

```python
load("go_idiomatic", "sum")
print(sum([1, 2, 3]), sum([1, 2, 4], start=8), sum([1, 2, None]))
# Output: 6 15 3
```

### `distinct(iterable) -> iterable`

Removes duplicates, preserving first-seen order. Returns a **list** for a list or custom iterable, a **tuple** for a tuple, the result of `.keys()` (a list) for a dict, and the original **set** unchanged for a set. Errors with `unhashable type: <type>` if an element cannot be hashed, and `got <type>, want iterable` for a non-iterable argument.

```python
load("go_idiomatic", "distinct")
print(distinct([1, 2, 2, 3, 3, 3]), distinct((1, 2, 2, 3)))
# Output: [1, 2, 3] (1, 2, 3)
```

### `hex(x) -> str` / `oct(x) -> str` / `bin(x) -> str`

Convert an integer to a base-16 / base-8 / base-2 string with the `0x` / `0o` / `0b` prefix. A negative number keeps the sign before the prefix (`-0xf`); zero is `0x0` / `0o0` / `0b0`. Arbitrary-precision integers are supported. Each errors with `missing argument for x` if called with no argument.

```python
load("go_idiomatic", "hex", "oct", "bin")
print(hex(255), oct(255), bin(255))
# Output: 0xff 0o377 0b11111111
```

```python
load("go_idiomatic", "hex", "oct", "bin")
print(hex(-15), oct(-56), bin(-255))
# Output: -0xf -0o70 -0b11111111
```

### `bytes_hex(bytes, sep="", bytes_per_sep=1) -> str`

Two lowercase hex digits per byte. With a one-character `sep`, the separator is inserted between groups of `bytes_per_sep` bytes — a positive count groups from the right, a negative count from the left. Errors with `missing argument for bytes` if `bytes` is omitted.

```python
load("go_idiomatic", "bytes_hex")
print(bytes_hex(b"123456"))
# Output: 313233343536
```

```python
load("go_idiomatic", "bytes_hex")
print(bytes_hex(b"123456", "_", 4))
print(bytes_hex(b"123456", "_", -4))
# Output: 3132_33343536
# 31323334_3536
```

### `is_nil(x) -> bool`

`True` for `None`/`nil`, or for a `starlight` Go wrapper (`GoSlice`, `GoMap`, `GoStruct`, `GoInterface`) whose underlying Go value is nil. Errors with `unsupported type: <type>` for any other Starlark value (e.g. an int) — it is intentionally not a general truthiness test.

```python
load("go_idiomatic", "is_nil")
print(is_nil(None))
# Output: True
```

### `sleep(secs)`

Blocks the current thread for `secs` seconds (int or float). `secs` must be non-negative — otherwise `secs must be non-negative`. The sleep is cancelled if the thread's context is done, returning that context error. Errors with `missing argument for secs` if omitted, or `want float or int` for a non-number.

```python
load("go_idiomatic", "sleep")
sleep(0.01)
# Output:
```

### `exit(code=0)` / `quit(code=0)`

Stores `code` as the thread-local `exit_code` and returns the sentinel error `ErrSystemExit` to unwind the program: `starlet runtime system exit (Use Ctrl-D in REPL to exit)`. `code` must fit an unsigned 8-bit range (0–255) — `exit(-1)` errors with `out of range`. `quit` is an exact alias.

```python
load("go_idiomatic", "exit")
exit(1)
# Output:
```

### `module(name, **kv) -> module`

Builds a `starlarkstruct` module. Unlike `struct`, its string form hides the fields (`<module "name">`) and it is **not** comparable with `==`/`!=`. Takes exactly one positional argument (the name); extra positionals error with `got N arguments, want 1`.

```python
load("go_idiomatic", "module")
m = module("rose", a=100, b="hello")
print(m, m.a, m.b)
# Output: <module "rose"> 100 hello
```

### `struct(**kv) -> struct` / `make_struct(name, **kv) -> struct`

`struct` builds an anonymous, field-comparable struct (printed as `struct(field = value, ...)`, fields sorted). `make_struct` is the same but takes a leading positional `name` used as the constructor in its string form and in equality (two structs are equal only if constructor *and* fields match). `struct` rejects positional arguments (`unexpected positional arguments`); `make_struct` requires exactly one (`got N arguments, want 1`).

```python
load("go_idiomatic", "struct", "make_struct")
print(struct(rose="red", lily="white"))
print(make_struct("rose", color="red", price=100))
# Output: struct(lily = "white", rose = "red")
# rose(color = "red", price = 100)
```

### `shared_dict() -> shared_dict` / `make_shared_dict(name="", data=None) -> shared_dict`

Both create a `shared_dict` (see the Types section). `shared_dict()` takes no arguments and yields an empty dict named `shared_dict`. `make_shared_dict` accepts an optional type `name` and an optional `data` dict to seed it. `shared_dict(123)` errors with `got 1 arguments, want 0`; passing a non-string `name` errors with `want string`.

```python
load("go_idiomatic", "make_shared_dict")
print(make_shared_dict("manaʻo", {"abc": 123}))
# Output: manaʻo({"abc": 123})
```

### `to_dict(v) -> dict`

Converts a value into a plain Starlark `dict`. Accepts a native `dict` (cloned), a `module` or `struct` (members become entries), a `GoStruct` (marshalled to JSON via Go's encoder then decoded to a dict — Go field names are preserved, nil fields become `None`), and a `shared_dict` (clone of its contents). Any other type errors with `unsupported type: <type>`; a Go struct holding a JSON-unencodable field (e.g. a channel) surfaces the encoder error (`json: unsupported type: chan int`).

```python
load("go_idiomatic", "to_dict", "struct", "module")
print(to_dict(struct(name="Alice", age=30)))
print(to_dict(module("mod", foo="bar", num=42)))
# Output: {"age": 30, "name": "Alice"}
# {"foo": "bar", "num": 42}
```

### `eprint(*args, sep=" ")`

Like the builtin `print()` but always writes to **stderr**, bypassing the Go `Print` handler — useful for diagnostics kept out of normal output. `sep` (a string) joins the arguments. A non-string `sep` errors with `got <type>, want string`.

```python
load("go_idiomatic", "eprint")
eprint("Path", "/home/user/docs", sep=" -> ")
# Output:
```

(Output goes to stderr, so stdout shows nothing.)

### `pprint(*args, sep=" ")`

Like `print()` but renders each argument as **indented JSON** (4-space indent) before writing through the thread's `Print` handler (falling back to stderr). Values that cannot be JSON-encoded — including self-referential structures — fall back to their string form, so `pprint` never fails on cyclic input. A non-string `sep` errors with `got <type>, want string`.

```python
load("go_idiomatic", "pprint")
pprint({"key": "value", "list": [1, 2, 3]})
# Output: {
#     "key": "value",
#     "list": [
#         1,
#         2,
#         3
#     ]
# }
```

## Notes / boundaries

- **Not a Python module clone.** Names echo Python builtins, but this is a Star\* utility grab-bag, not a 1:1 API mirror — there is no module-level namespace; members load individually.
- **`length` vs `len`.** `length` counts Unicode code points for strings; the builtin `len` counts UTF-8 bytes. They agree for sequences.
- **`distinct` ordering.** First-seen order is preserved for list/tuple/custom inputs; dict-key order from `.keys()` is **not** guaranteed (sort it if you need stability). Sets are returned unchanged.
- **Process effects.** `sleep` blocks (and respects context cancellation); `exit`/`quit` do not stop the goroutine themselves — they return `ErrSystemExit` for the host runner to act on, and set the thread-local `exit_code`.
- **Determinism.** `struct`/`make_struct`/`module` print fields in sorted order; `to_dict` of a `GoStruct` keeps the Go field names verbatim.
- **`shared_dict` scope.** Locking protects a single instance across threads; it does not make the values it holds deeply immutable or persist anything beyond the process.

## Expression evaluation

`eval` and `try_eval` evaluate one **Starlark expression**, using only explicit
variables and approved computational builtins. They do not execute Go or
[Expr-language](https://github.com/expr-lang/expr) syntax, guess the language, or
retry failures with another engine. The `expr` module name remains available
for a separate future Expr-language adapter.

`source` must be a string. `vars` must be a native dict whose keys are valid
Starlark identifiers. Omitting `vars` supplies an empty environment. The module
does not inspect caller locals or copy a Machine's implicit globals.

```python
load("go_idiomatic", "eval", "try_eval")

items = [2, 3]
result = eval(
    '{"title": title.upper(), "values": [x * 2 for x in items]}',
    vars={"title": "example", "items": items},
)
# {"title": "EXAMPLE", "values": [4, 6]}

value, error = try_eval("items.append(9)", vars={"items": items})
# value == None; error["phase"] == "runtime"
items.append(4)  # The original input was neither changed nor frozen.

value, error = try_eval("None")
# (None, None): a valid None result is distinct from failure.
```

Preloading `go_idiomatic` makes `eval(...)` and `try_eval(...)` available as
bare functions, like its existing `sum` and `hex` helpers. A `NewDefault()`
Machine does not preload modules; configure this module explicitly. Neither a
`starlark` module nor a `go_idiomatic.eval` attribute wrapper is registered.
When importing `json.eval` or a future Expr-language evaluator too, use explicit
load aliases to distinguish the engines.

The existing Machine precedence also applies to these two new names: preload
functions override same-named host globals; per-run extras override preloads;
explicit `load` bindings take precedence in the script. Use lazy loading with
an alias, such as `load("go_idiomatic", evaluate="eval")`, when keeping a host's
existing `eval` name matters.

### Data and method boundaries

Inputs and results preserve native `None`, bool, int (including big integers),
float, string, bytes, list, dict, tuple and set values. Nested data is supported;
aliases are preserved within a copied graph. Reference cycles, nil values,
functions, bound methods, modules, iterators and custom or Go-backed values are
rejected with a data path. Convert a host object into an explicit native data
projection first; wrapping it with Starlight does not create a snapshot.

Only private input copies are frozen. Native read methods such as `upper`,
`lower`, `get` and `index`, comprehensions and approved computational builtins
remain usable. Mutating a frozen input fails. Expressions may create and modify
their own temporary containers; results are checked, copied and frozen before
publication. Even a frozen lambda or bound method remains callable, so callable
results are rejected recursively.

Caller functions and external modules are not inherited. Unapproved Universe
builtins, including `print`, are blocked without changing the global Universe or
calling the parent's Print handler. The standard computational builtin instances
are captured during package initialization; the embedding application's startup
must retain the standard definitions then. As upstream requires, do not modify
Universe concurrently with execution. Other `go_idiomatic` helpers retain their existing Log/Process capabilities;
they are not inherited by an expression. Random and clock sources are not
automatically exposed either.

The source graph must be exclusively owned/synchronized during evaluation, or
already deeply frozen native data. A lack of writers alone is insufficient:
iterating an unfrozen native set can change its iterator bookkeeping. The helper
does not acquire a Machine or application data lock for the caller.

### Errors and execution state

`error_info` is a frozen dict with these fields:

- `phase` and `kind`: the failing stage or policy category;
- `message`: a diagnostic string;
- `position`: `None` or `{file, line, column}`;
- `frames`: a tuple of `{file, line, column, name}` dicts;
- `truncated`: whether diagnostic text or frames were shortened.

Stages distinguish arguments, input, source admission, parsing, name resolution,
runtime, policy and output checks. Messages are limited to 4096 bytes, the most
recent 32 frames are retained, and frame file/name text is bounded.

Both functions execute on the caller's existing Thread. Its cancellation,
OnMaxSteps behavior and cumulative VM step count remain in force; no new step
allowance is installed or reset. Execution checkpoints also consume VM steps.
`try_eval` does **not** turn caller cancellation, exhausted total steps or a caller
panic into a successful tuple. It only handles errors occurring inside the call;
errors while evaluating the call's arguments still belong to the outer script.

There is no rollback of host callback effects. Such callbacks are excluded from
the data boundary. Errors discard the private evaluation data.

### Admission limits

| Limit | Value |
| --- | --- |
| Source bytes | 4096 |
| Non-whitespace source bytes | 512 |
| AST depth / nodes | 64 / 512 |
| Each copied data graph's depth / visited edges | 64 / 4096 |
| Each copied graph's scalar/name payload | 1 MiB |
| Integer magnitude | At most 4096 bits |

The pre-parser complexity count includes non-whitespace bytes in strings and
comments; only ASCII space, tab, CR and LF are excluded. Put large data in `vars`.
Input and output copying have separate graph quotas. Every visited edge counts,
including memo hits; payload accounting follows the copied graph rather than an
expanded JSON serialization.

These are admission and result limits, **not hard RSS or CPU limits**. The pinned
parser's general recursion issue is not repaired here. VM steps do not account for
every operation inside a native builtin; an intermediate allocation can occur
before output admission. Configure the caller's execution limits, and use a
separately validated process boundary when stronger isolation is required.

### Go usage

```go
import (
    "github.com/1set/starlet/lib/goidiomatic"
    "go.starlark.net/starlark"
)

thread := &starlark.Thread{Name: "expression"}
thread.SetMaxExecutionSteps(10000)
result, err := goidiomatic.Eval(thread, "qty * 2", starlark.StringDict{
    "qty": starlark.MakeInt(3),
})
```

The caller owns the Thread and source graph. Direct Go callers must wire
cancellation to `Thread.Cancel` when cancellation during VM execution is needed;
a context stored in Thread.Local is checked at helper checkpoints, not a new
cancellation goroutine. Machine calls retain their existing cancellation bridge.
The Go helper returns `starlark.Value` and an error preserving parser, resolver and
EvalError causes through `errors.As`/`errors.Is`. It does not implicitly convert
the result to Go or JSON.

The existing `go_idiomatic` capability classification remains Log + Process.
Adding these data-only helpers does not reduce that classification or expose the
module's other helpers inside an expression. Consumers that already allow or
preload `go_idiomatic` receive the new names with it; in particular, Starbox's
Safe and Network module sets already contain this module. Hosts needing a
separate explicit opt-in can instead pass `goidiomatic.Eval` through their own
loader or builtin policy. None of these policies establishes a hard sandbox.
