# go_idiomatic

`go_idiomatic` provides Go-flavored helpers, values and constructors for Starlark, including `true`/`false`/`nil`, structs and shared dictionaries. It also adds collection, output and read-only expression-evaluation helpers.

Capability profile: **Log + Process**. `eprint`/`pprint` write to stderr or the thread's `Print` handler (Log); `sleep` blocks the thread and `exit`/`quit` halt the program (Process). Everything else is pure.

Load name for `load()`: `go_idiomatic`.

## Functions

| function | description |
|----------|-------------|
| `eval(source, vars={}) -> value` | Evaluate a Starlark expression. [Parameters and examples](#expression-evaluation). |
| `try_eval(source, vars={}) -> (value, error_info)` | Go-style value/error results. [Return values and errors](#returns-and-errors). |
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

## Expression evaluation

Evaluate one **Starlark expression** with `eval`. Use `try_eval` for Go-style
`(value, error)` results: `error_info == None` means success.

```python
load("go_idiomatic", "eval", "try_eval")

total = eval("price * qty", vars={"price": 12, "qty": 3})
print(total)  # 36

value, error = try_eval("price")  # No implicit caller variables.
print(value, error["phase"])  # None resolve
```

### Parameters

```text
eval(source, vars={}) -> value
try_eval(source, vars={}) -> (value, error_info)
```

Both parameters accept positional or keyword arguments.

| Parameter | Type | Required / default | Meaning |
| --- | --- | --- | --- |
| `source` | Native `string` | Required | A single expression, such as `"price * qty"`. Statements are not accepted. |
| `vars` | Native `dict` with string keys | Optional; `{}` | Variables visible to the expression. Keys must be valid Starlark identifiers, at most 512 bytes each. `None` is not accepted. |

`vars` supports nested native `None`, bool, int, float, string, bytes, list,
tuple, dict and set values. Functions, methods, modules, iterators, cycles and
Go-backed/custom values (including `shared_dict`) are rejected.

Input snapshots and returned containers are frozen; originals are untouched.
Native reads such as `.upper()`, `.get()` and `.index()` work; input writes fail.
New temporary containers inside an expression may be mutated.

```python
items = [1, 2]
value, error = try_eval("items.append(9)", vars={"items": items})
print(error["phase"])  # runtime
items.append(3)         # Original list remains writable: [1, 2, 3].
```

Expressions see only `vars` and these standard computation names:

```text
None True False abs any all bool bytes chr dict dir enumerate fail float
getattr hasattr hash int len list max min ord range repr reversed set
sorted str tuple type zip
```

Caller locals, other helpers (`sum`, `sleep`, `exit`) and modules such as
`runtime` are not inherited. `print()` is disabled.

### Returns and errors

| Outcome | `eval` | `try_eval` |
| --- | --- | --- |
| Success | Frozen native value | `(value, None)` |
| Successful expression `None` | `None` | `(None, None)` |
| Argument, input or evaluation failure | Raises an error | `(None, error_info)` |
| Caller cancellation or total step limit reached | Raises an execution error | Raises an execution error |

Test `error != None` for failure; valid results include `None`, `False`, `0` and
empty containers. Cancellation and cumulative steps remain in force.
`try_eval` does not catch caller panics or errors computing its arguments:
`try_eval("1 // 0")` catches division by zero; `try_eval(1 // 0)` does not.

### `error_info` structure

`error_info` is a **frozen dict** with six fields. Use `error["phase"]` to read
fields; nested dicts and frames are frozen too.

| Field | Type | Meaning |
| --- | --- | --- |
| `phase` | `string` | Failing stage, listed below. |
| `kind` | `string` | Same as `phase`, except `policy` uses `"capability"`. |
| `message` | `string` | Human-readable diagnostic, at most 4096 bytes. Do not match its exact wording in program logic. |
| `position` | `None` or `dict` | Best available source location: `{"file": string, "line": int, "column": int}`. Lines and Unicode character (rune) columns are 1-based; 0 means unknown. |
| `frames` | `tuple` of `dict` | Frames from outermost to innermost, retaining the most recent 32. Each has the location fields above plus `name: string`. Builtin frames can use `"<builtin>"`, 0, 0. |
| `truncated` | `bool` | Whether message/location/frame text or the frame count was truncated. File names are capped at 512 bytes, frame names at 256. |

| `phase` | Typical cause |
| --- | --- |
| `arguments` | Missing/extra arguments, duplicate or unknown keywords. |
| `input` | Wrong parameter type, invalid variable name or unsupported/oversized input data. |
| `source` | Source length or AST complexity limit exceeded. |
| `parse` | Invalid expression syntax. |
| `resolve` | Undefined name, such as an omitted variable or caller-only helper. |
| `policy` | Denied builtin, such as `print`; `kind` is `"capability"`. |
| `runtime` | Division by zero, invalid operation or attempted input mutation. |
| `output` | Result contains a callable/unsupported value or exceeds data limits. |

For `try_eval("price")` with no variables:

```python
{
    "phase": "resolve",
    "kind": "resolve",
    "message": "eval: resolve: <eval>:1:1: undefined: price",
    "position": {"file": "<eval>", "line": 1, "column": 1},
    "frames": (),
    "truncated": False,
}
```

Argument/input/source errors usually have `position=None` and `frames=()`.
Runtime/policy errors can include caller frames; `position` uses the innermost
available non-builtin location. Cancellation/budget errors are raised, never
returned with `phase="execution"`.

### Loading and limits

Preload `go_idiomatic` for bare calls, or enable lazy loading for `load(...)`.
`NewDefault()` enables neither. Name precedence is host globals < preloads <
per-run extras < explicit `load`. Avoid clashes with an alias:
`load("go_idiomatic", evaluate="eval")`.

| Limit | Value |
| --- | --- |
| Source | 4096 bytes; 512 non-whitespace bytes, including literals/comments |
| AST | Depth 64; 512 nodes |
| Each copied input/output graph | Depth 64; 4096 visited edges, including repeated references |
| Each graph's scalar/name payload | 1 MiB |
| Integer magnitude | 4096 bits |

Only ASCII space, tab, CR and LF are excluded from the source count. Put large
data in `vars`. These quotas do not cap native intermediate allocations, total
CPU or process memory; configure caller execution limits for your workload.

### Go callers

The Go helper returns `(starlark.Value, error)`, not an `error_info` dict:

```go
import (
    "github.com/1set/starlet/lib/goidiomatic"
    "go.starlark.net/starlark"
)

func evaluateQuantity(qty int) (starlark.Value, error) {
    thread := &starlark.Thread{Name: "expression"}
    thread.SetMaxExecutionSteps(10000)
    return goidiomatic.Eval(thread, "qty * 2", starlark.StringDict{
        "qty": starlark.MakeInt(qty),
    })
}
```

Own the Thread and own/synchronize input data, or deeply freeze it first.
Bridge direct Go cancellation to `Thread.Cancel`; Machine does this already.
Use `errors.As` with `*goidiomatic.Error` to inspect `Phase` and `Kind`; error
causes remain available through `errors.As`/`errors.Is`. Results are not converted
to Go/JSON. Keep upstream Universe standard at startup and never mutate it
concurrently. See [`Eval`](eval.go) for the host contract.

The module remains Log + Process and is already in Starbox Safe/Network.
Use host module policy for opt-in; hard CPU/memory isolation requires a separately
validated process boundary.

### Performance

`eval` reuses the caller's Thread and builds a private environment from explicit
variables. Every call parses and compiles the expression, copies and validates
input/output data, and checks the caller's execution limits. Reusing `vars`
does not cache the compiled expression or remove this preparation work.

Benchmarks against the same expression in a compiled Starlark function found
that this preparation has a noticeable cost for small calculations. With fixed
source and small input/output graphs, more computation can reduce the relative
cost. The total time difference is not a single constant: longer source, larger
data graphs, allocation and GC also affect it. Large-workload variation does not
establish that dynamic evaluation is faster or has no overhead.

Use direct script code for fixed calculations and hot loops; use `eval` when the
expression itself is supplied at runtime. Timings depend on CPU, OS,
architecture, Go/compiler and system load. Measure on your target system with
`BenchmarkEvalComparison` and `BenchmarkEvalWorkload`; [benchmark notes](BENCHMARKS.md)
provide the workloads, sample results and reproduction commands.

## Types

### `shared_dict`

A mutable dictionary returned by `shared_dict()` and `make_shared_dict()`. Indexing and dictionary methods lock the top-level mapping; this does not protect mutable nested values or make every operation safe during concurrent writes. In particular, format/print the dictionary only while writers are stopped or externally synchronized. Its `type()` is `shared_dict` by default, or the custom name passed to `make_shared_dict`. It supports indexing (`sd["k"] = v`, `sd["k"]`), membership (`in`), the `.len()` method, and `==`/`!=` comparison with another `shared_dict`. It is **not** iterable, and the builtin `len()` does not apply — use the `.len()` method.

Beyond the standard dict methods it inherits (`clear`, `get`, `items`, `keys`, `pop`, `popitem`, `setdefault`, `update`, `values`), it adds the methods below.

| method | description |
|--------|-------------|
| `len() -> int` | Number of items in the dictionary. |
| `perform(fn)` | Call `fn(dict)` with the underlying plain dict while holding the lock, for atomic compound updates. Do not retain that dict or reenter this `shared_dict` from the callback. |
| `to_dict() -> dict` | Copy top-level entries into a plain dict. Inserting/removing clone entries is independent; nested values remain shared. |
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

`to_dict` copies the top-level mapping; nested values remain shared:

```python
load("go_idiomatic", "make_shared_dict")
sd = make_shared_dict("", {"items": [1]})
clone = sd.to_dict()
clone["new"] = 3
clone["items"].append(2)
print("new" in sd, sd["items"])
# Output: False [1, 2]
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

- **`length` vs `len`.** `length` counts Unicode code points for strings; the builtin `len` counts UTF-8 bytes. They agree for sequences.
- **`distinct` ordering.** First-seen order is preserved for list/tuple/custom inputs; dict-key order from `.keys()` is **not** guaranteed (sort it if you need stability). Sets are returned unchanged.
- **Process effects.** `sleep` blocks (and respects context cancellation); `exit`/`quit` do not stop the goroutine themselves — they return `ErrSystemExit` for the host runner to act on, and set the thread-local `exit_code`.
- **Determinism.** `struct`/`make_struct`/`module` print fields in sorted order; `to_dict` of a `GoStruct` keeps the Go field names verbatim.
- **`shared_dict` scope.** Locking protects a single instance across threads; it does not make the values it holds deeply immutable or persist anything beyond the process.
