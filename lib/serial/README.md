# serial

`serial` serializes Starlark **data values** to and from a compact JSON envelope, round-tripping the types plain JSON cannot: `bytes`, `set`, `tuple`, arbitrary-precision `int`, `time`, and dicts with **non-string keys**. It is the persistence companion to the `json` module — where `json.encode`/`decode` speak the JSON subset, `serial.dumps`/`loads` preserve the full Starlark data shape so a value written by one script reads back identically in another (caches, content-addressed keys, cross-run state).

**Capability profile: Pure.** No filesystem, network, process, or logging side effects — it only transforms a value to a string and back.

## Functions

| function | description |
|---|---|
| `dumps(value) -> string` | serialize a data value to a JSON-envelope string (deterministic; usable as a cache key) |
| `loads(s) -> value` | reconstruct the value from a `dumps` string; returns a fresh, unfrozen value |
| `try_dumps(value) -> tuple` | `dumps` variant returning `(result, error)` instead of aborting |
| `try_loads(s) -> tuple` | `loads` variant returning `(result, error)` instead of aborting |

The `try_*` variants never abort the script: on success they return `(result, None)`; on failure they return `(None, message)` where `message` is the error string. The non-`try` `dumps`/`loads` raise the same error instead.

## The contract: lossless, or a clear error — never silently lossy

A value either round-trips **losslessly** or `dumps` fails with an **actionable error**. There is no quietly-lossy middle ground, because flattening an object would drop its type identity, methods, or live host binding without telling you.

**Round-trips losslessly** — `None`, `bool`, `int` (including big integers), finite `float`, `str`, `bytes`, `list`, `tuple`, `dict` (string-keyed and non-string-keyed), `set`, `time`.

**Rejected with an error** (convert to data first, or store differently):

| value | why | what to do |
|---|---|---|
| function / lambda / builtin | code/closures can't be serialized | store the `.star` script and `load()` it ("persist data, replay code") |
| `struct` | the constructor / type identity can't be persisted | convert it to a dict first |
| host Go objects (Go-backed wrappers) | the Go type, methods and live binding would be lost | convert to a dict on the host side first |
| non-finite `float` (`NaN`, `±Inf`) | JSON has no representation | guard before dumping |
| a value that refers to itself (cycle) | infinite structure | break the cycle |

This directly answers "will serializing a complex object lose information?" — serial won't let you serialize an object at all; it tells you to flatten it first, so nothing is dropped behind your back.

## Encoding

Non-JSON-native values are wrapped in a `{"$t": <tag>, "v": <payload>}` envelope. The tags are:

| tag | for | payload |
|---|---|---|
| `bytes` | `bytes` | base64 string |
| `bigint` | `int` outside int64 | decimal string |
| `tuple` | `tuple` | JSON array of encoded elements |
| `set` | `set` | JSON array, sorted by encoded form |
| `time` | `time` | RFC 3339 (nanosecond) string |
| `mapkv` | dict with non-string keys | array of `[key, value]` pairs, sorted by encoded key |
| `object` | a real dict that itself contains a `"$t"` key | the dict, so it is never mistaken for an envelope on the way back |

An int that fits `int64` is a plain JSON number; an all-string-keyed dict is a plain JSON object. Output is **deterministic**: object keys are sorted (by `json.Marshal`), and set elements and `mapkv` pairs are ordered by their encoded bytes, so the same value always dumps to the same string — safe to use directly as a cache key.

## Details & examples

### `dumps(value) -> string`

Serialize `value` to a JSON-envelope string. Walks the value directly (never via re-parse, which would collapse type information). Returns the deterministic JSON text.

**Errors on**: a function/lambda/builtin (`cannot serialize function: it is code …`), a `struct` (`convert it to a dict first`), a host Go object (`serial round-trips data, not host objects`), a non-finite float (`cannot serialize non-finite float …`), or a reference cycle (`cannot serialize a value that refers to itself (cycle …)`). The error propagates from any depth — an unserializable element inside a list, tuple, set, dict value, or non-string dict key fails the whole `dumps`.

```python
load('serial', 'dumps')
print(dumps({'b': 2, 'a': 1}))
# Output: {"a":1,"b":2}
```

### `loads(s) -> value`

Reconstruct the value from a `dumps` string, interpreting the type tags. The result is a fresh, **unfrozen** value (the same as `json.decode`), so scripts can read or mutate it.

**Errors on**: invalid JSON (`serial.loads: …`), trailing content after the value (`unexpected trailing data after JSON value` — a second JSON document or garbage; trailing whitespace is fine), an unknown type tag (`unknown type tag "…"`), or a malformed envelope payload whose `v` has the wrong shape (`invalid bytes payload`, `invalid bigint payload`, `invalid time payload`, `invalid tuple payload`, `invalid set payload`, `invalid mapkv payload`/`invalid mapkv entry`, `invalid object payload`). A `set` with an unhashable element or a `mapkv` with an unhashable key errors (`unhashable`). Bare JSON numbers decode without a tag: an integer (any precision) to `int`, a number with `.`/`e`/`E` to `float`.

```python
load('serial', 'loads')
print(loads('{"a":1,"b":[2,3]}'))
# Output: {"a": 1, "b": [2, 3]}
```

### Round-trip the types JSON drops

`loads(dumps(x))` reproduces `x` exactly, preserving the type — a `tuple` stays a `tuple`, `bytes` stay `bytes`, a `set` stays a `set`, and a big integer keeps full precision.

```python
load('serial', 'dumps', 'loads')
def rt(x): return loads(dumps(x))
v = {'id': 1267650600228229401496703205376, 'tags': set(['a', 'b']), 'raw': b'abc', 'pair': (1, 2), 'm': {1: 'x'}}
print(rt(v) == v, type(rt((1, 2))), type(rt(b'x')))
# Output: True tuple bytes
```

### `try_dumps(value) -> tuple` / `try_loads(s) -> tuple`

Same as `dumps`/`loads` but report failure as a `(result, error)` tuple instead of aborting the script. On success `error` is `None`; on failure `result` is `None` and `error` is the message string.

```python
load('serial', 'try_dumps', 'try_loads')
out, err = try_dumps(42)
print(out, err)
# Output: 42 None
```

```python
load('serial', 'try_loads')
val, e1 = try_loads('[1,2,3]')
bad, e2 = try_loads('not json')
print(val, e1, bad, e2 != None)
# Output: [1, 2, 3] None None True
```

## Notes / boundaries

- **Engine** — encoding/decoding uses the Go standard `encoding/json`; decoding preserves number precision via `json.Number` (`UseNumber`), so large integers do not lose precision through a float round-trip.
- **Determinism** — output is byte-stable for a given value: object keys sorted, set elements and `mapkv` pairs ordered by encoded form. This is what makes a `dumps` string usable as a cache key.
- **`$t` reserved key** — `"$t"` is the envelope discriminator. A real string-keyed dict that contains a `"$t"` key is automatically wrapped in an `object` envelope and unwrapped on `loads`, so such dicts still round-trip; you do not need to avoid the key.
- **Difference from `json`** — `json.encode`/`decode` cover only JSON-native types and silently lose `tuple`/`set`/`bytes`/big-int/`time`/non-string keys; `serial` preserves all of them and instead refuses (with an actionable error) the things that genuinely cannot be persisted as data (code, structs, host objects, non-finite floats, cycles).
- **`module name`** — load as `load('serial', 'dumps', 'loads')`; the module constant `ModuleName` is `"serial"`.
- All exported names are snake_case; no non-conforming identifiers.
