# json

`json` converts Starlark values to and from JSON text and offers a small toolkit around it: pretty-printing, JSONPath query/evaluation, LLM-output repair, and JSON Schema validation. It extends Go Starlark's stdlib `json` (`encode`/`decode`/`indent`) with `dumps`-style and `try_*` helpers. **Capability profile: pure** — no filesystem, network, or process side effects (external schema `$ref` is deliberately blocked to keep it so).

Every host-error function ships a `try_*` twin that, instead of aborting the script, returns a `(result, error)` tuple — `error` is `None` on success, and `result` is `None` on failure. `try_validate` is the one exception, distinguishing three outcomes (see below).

## Functions

| function | description |
| --- | --- |
| `encode(x) -> string` | Encode a Starlark value to compact JSON text (go.starlark.net stdlib). |
| `decode(x[, default]) -> value` | Decode a JSON string to a Starlark value; on bad input returns `default` if given, else errors. |
| `indent(str, *, prefix="", indent="\t") -> string` | Pretty-print valid JSON text with the given prefix/indent unit (stdlib). |
| `dumps(obj, indent=0) -> string` / `try_dumps(obj, indent=0) -> tuple` | Encode a Starlark value (incl. struct/module) to JSON text, optionally indented by `indent` spaces. `try_dumps` returns `(text, error)`. |
| `encode(x)` / `try_encode(x) -> tuple` | `try_encode` is the tuple-returning variant of `encode`. |
| `decode(x)` / `try_decode(x) -> tuple` | `try_decode` is the tuple-returning variant of `decode` (no `default` param). |
| `indent(...)` / `try_indent(str, prefix="", indent="\t") -> tuple` | `try_indent` is the tuple-returning variant of `indent`. |
| `path(data, path) -> list` / `try_path(data, path) -> tuple` | Run a JSONPath query over `data`, returning the list of matches. `try_path` returns `(list, error)`. |
| `eval(data, expr) -> value` / `try_eval(data, expr) -> tuple` | Evaluate a JSONPath expression (aggregates, arithmetic, comparisons) over `data`. `try_eval` returns `(value, error)`. |
| `repair(text) -> string` / `try_repair(text) -> tuple` | Recover valid JSON *text* from messy/LLM output (fences, prose, single quotes, trailing commas, truncation). `try_repair` returns `(text, error)`. |
| `validate(data, schema) -> None` / `try_validate(data, schema) -> tuple` | Validate a JSON document against a JSON Schema. `validate` returns `None` or errors; `try_validate` returns one of three outcomes. |

`data` and `schema` arguments to `path` / `eval` / `validate` accept a JSON `string`, `bytes`, or any encodable Starlark value (dict, list, struct, …).

This module exposes no custom Starlark types — every result is a standard Starlark value (`dict`, `list`, `string`, `int`, `float`, `bool`, or `None`).

## Details & examples

### `encode` / `try_encode`

`encode(x) -> string` converts a Starlark value to compact JSON using go.starlark.net's stdlib rules:

- `None`, `True`, `False` → `null`, `true`, `false`.
- Starlark ints (any size) → decimal integers; floats → decimal-point notation. Non-finite floats are an error.
- Strings → JSON strings (UTF-16 escapes); `dict`/IterableMapping → object (non-string keys error); other Iterable (`list`, `tuple`) → array; `HasAttrs` (`struct`) → object.

It **errors** when a value cannot be encoded (e.g. a function: `cannot encode function as JSON`). `try_encode(x)` returns `(text, None)` on success or `(None, message)` on failure.

```python
load('json', 'encode')
load("struct.star", "struct")
s = struct(a="Aloha", b=0x10, c=True, d=[1,2,3])
print(encode(s))
# Output: {"a":"Aloha","b":16,"c":true,"d":[1,2,3]}
```

```python
load('json', 'try_encode')
result, error = try_encode({'a': 10, 'b': 20})
print(result, error)
# Output: {"a":10,"b":20} None
```

### `decode` / `try_decode`

`decode(x[, default]) -> value` parses a JSON string into a Starlark value: numbers become `int` or `float` (by presence of a decimal point), objects become unfrozen `dict`s, arrays become unfrozen `list`s. On invalid input it returns `default` if supplied, otherwise it **errors**. `try_decode(x)` (no `default`) returns `(value, None)` or `(None, message)`.

```python
load('json', 'decode')
print(decode('{"a":10,"b":20}'))
# Output: {"a": 10, "b": 20}
```

```python
load('json', 'try_decode')
result, error = try_decode('{"a": "b"}')
print(result, error)
# Output: {"a": "b"} None
```

### `indent` / `try_indent`

`indent(str, *, prefix="", indent="\t") -> string` re-formats already-valid JSON text. `prefix` and `indent` are keyword-only on the stdlib `indent`; on `try_indent` they are ordinary optional params. It **errors** on invalid JSON text (e.g. `invalid character ...`). `try_indent(str, prefix="", indent="\t")` returns `(text, error)`.

```python
load('json', 'indent')
print(indent('{"a":10,"b":20}', indent="  "))
# Output:
# {
#   "a": 10,
#   "b": 20
# }
```

### `dumps` / `try_dumps`

`dumps(obj, indent=0) -> string` encodes any Starlark value via the internal marshaler (which, unlike `encode`, also handles host structs/modules). `indent` is the number of spaces per level; `0` or negative produces compact output. It **errors** when a value cannot be marshaled (e.g. a function: `unrecognized starlark type: *starlark.Function`). `try_dumps(obj, indent=0)` returns `(text, error)`.

```python
load('json', 'dumps')
print(dumps({'a': 10, 'b': 20}, indent=2))
# Output:
# {
#   "a": 10,
#   "b": 20
# }
```

```python
load('json', 'try_dumps')
result, error = try_dumps(1, indent=-7)
print(result, error)
# Output: 1 None
```

Note: `dumps`/`encode` can differ for host structs. A struct carrying a `star` tag encodes by its Go field names under `encode` (`{"Message":...}`) but by its struct values under `dumps` — see the test suite for the exact shapes.

### `path` / `try_path`

`path(data, path) -> list` runs a JSONPath query and returns the list of matching elements (empty list if nothing matches). Numeric matches come back as `int` when integral, else `float`. It **errors** on a malformed JSONPath expression (`wrong symbol 'X' at N`) or on `data` that is neither valid JSON nor an encodable value (`unrecognized starlark type`). `try_path(data, path)` returns `(list, error)`.

```python
load('json', 'path')
data = '''{"store":{"book":[{"title":"Sayings of the Century","price":8.95},{"title":"Sword of Honour","price":12.99}]}}'''
print(path(data, '$.store.book[*].title'))
# Output: ["Sayings of the Century", "Sword of Honour"]
```

```python
load('json', 'try_path')
data = {'items': [{'value': 5}, {'value': 10}, {'value': 15}]}
result, error = try_path(data, '$.items[?(@.value > 7)].value')
print(result, error)
# Output: [10, 15] None
```

### `eval` / `try_eval`

`eval(data, expr) -> value` evaluates a JSONPath *expression* — aggregates (`sum`, `avg`, `size`), arithmetic, comparisons, string concatenation, and built-in constants (`pi`) — and returns a single `value` (number, string, bool, list, dict, or `None`). It **errors** on an unknown function (`'invalid' is not a function`), bad syntax, division by zero, invalid `data`, or an unencodable value. `try_eval(data, expr)` returns `(value, error)`.

```python
load('json', 'eval')
data = '''{"store":{"book":[{"price":8.95},{"price":12.99},{"price":8.99},{"price":22.99}],"bicycle":{"price":19.95}}}'''
print(eval(data, 'avg($..price)'))
# Output: 14.774000000000001
```

```python
load('json', 'try_eval')
result, error = try_eval({'value': 10}, '$.value > 5')
print(result, error)
# Output: True None
```

### `repair` / `try_repair`

`repair(text) -> string` recovers valid JSON **text** from messy model output so it can then be `decode`d — the idiom is `decode(repair(x))`. It strips code fences (```` ```json … ``` ````), surrounding prose, fixes single quotes, trailing commas, comments, Python literals (`True`/`False`/`None`), and completes truncated JSON.

- **Idempotent on good input**: already-valid JSON is returned byte-for-byte unchanged (so it never mangles valid escapes — calling it defensively is safe).
- Because it returns *text*, a recovered bare scalar is honest output (`repair('The answer is 42')` decodes to `42`, not a dict); scripts needing structure should check the decoded type.
- It **errors** on truly unrepairable input (e.g. `{,,,}`). `try_repair(text)` returns `(text, error)` — including on a non-string argument.

```python
load('json', 'repair', 'decode')
messy = '''Here is the result:
```json
{'name': 'Ann', 'tags': ['a', 'b',],}
```
'''
print(decode(repair(messy)))
# Output: {"name": "Ann", "tags": ["a", "b"]}
```

```python
load('json', 'try_repair')
result, error = try_repair('{"a": 1,}')
print(result, error)
# Output: {"a": 1} None
```

### `validate` / `try_validate`

`validate(data, schema) -> None` checks a JSON document against a [JSON Schema](https://json-schema.org) (drafts 4, 6, 7, 2019-09, 2020-12 — detected from `$schema`, default 2020-12). It returns `None` when the data conforms; otherwise it **errors** with one line per violation, each prefixed by its [JSON Pointer](https://datatracker.ietf.org/doc/html/rfc6901) location (e.g. `at /age: must be >= 0 but found -3`; long lists are capped with `... and N more`).

Schemas must be **self-contained**: an external `$ref` (`file://`, `http://`) is rejected (`not allowed`) — this is what keeps the module pure. A bad schema or malformed data text reports `invalid schema` / `invalid data`. Compiled schemas are cached (bounded), so repeated validation against the same schema text avoids recompilation.

`try_validate(data, schema)` distinguishes three outcomes:

- `(True, None)` — the data conforms.
- `(False, details)` — the data was checked and is invalid; `details` lists the violations.
- `(None, error)` — validation could not run (invalid schema, malformed JSON text, or bad arguments).

```python
load('json', 'validate')
schema = {'type': 'object', 'required': ['name'], 'properties': {'name': {'type': 'string'}, 'age': {'type': 'integer', 'minimum': 0}}}
print(validate({'name': 'Ann', 'age': 3}, schema))
# Output: None
```

```python
load('json', 'try_validate')
ok, err = try_validate('{"age":-3}', '{"type":"object","properties":{"age":{"type":"integer","minimum":0}}}')
print(ok, err)
# Output: False at /age: must be >= 0 but found -3
```

## Notes & boundaries

- **Engines.** `encode`/`decode`/`indent` are go.starlark.net's stdlib `json`; `dumps` uses starlet's internal marshaler (handles host structs/modules); `path`/`eval` use [ajson](https://github.com/spyzhov/ajson) JSONPath; `repair` uses a vendored, frozen [jsonrepair](https://github.com/kaptinlin/jsonrepair) (golden-locked); `validate` uses [santhosh-tekuri/jsonschema](https://github.com/santhosh-tekuri/jsonschema).
- **Purity.** No file or network access. JSON Schema `$ref` to external resources is blocked by design.
- **Number shaping.** `path`/`eval` return integral numbers as `int` and non-integral as `float`; JSON `null` becomes `None`.
- **`repair` vs `validate`.** `repair` fixes *text* and is idempotent on valid input; `validate` never mutates — it only reports conformance.
- All function names are snake_case; `try_*` variants mirror their base function and never abort the script.
- There is **no** `encode_indent` function: indentation is a parameter, not a separate call — use `dumps(obj, indent=N)` for indented encoding, or `indent(...)` to re-format existing JSON text.
