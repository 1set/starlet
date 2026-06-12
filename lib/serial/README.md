# serial

`serial` serializes Starlark **data values** to and from a compact JSON envelope, round-tripping the types plain JSON cannot: `bytes`, `set`, `tuple`, arbitrary-precision `int`, `time`, and dicts with **non-string keys**. It is the persistence companion to the `json` module — where `json.encode`/`decode` speak the JSON subset, `serial.dumps`/`loads` preserve the full Starlark data shape so a value written by one script reads back identically in another (caches, content-addressed keys, cross-run state).

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

## Functions

| function | description |
|---|---|
| `dumps(value) string` | serialize a data value to a JSON-envelope string (deterministic; usable as a cache key) |
| `loads(s) value` | reconstruct the value from a `dumps` string |
| `try_dumps(value) tuple` | `dumps` variant returning `(result, error)` instead of aborting |
| `try_loads(s) tuple` | `loads` variant returning `(result, error)` |

`loads` returns a fresh, **unfrozen** value (the same as `json.decode`), so scripts can read or mutate the result.

## Encoding

Non-JSON-native values are wrapped in a `{"$t": <tag>, "v": <payload>}` envelope: `bytes` (base64), `set`, `tuple`, `bigint` (decimal string), `time` (RFC 3339), and `mapkv` for a dict with non-string keys (a list of `[key, value]` pairs). A real dict that itself contains a `"$t"` key is wrapped in an `object` envelope so it is never mistaken for a tagged value on the way back. Output is **deterministic**: object keys are sorted and set elements are ordered by their encoded form, so the same value always dumps to the same bytes.

## Examples

**Round-trip the types JSON drops**

```python
load('serial', 'dumps', 'loads')
v = {'id': 2**80, 'tags': set(['a', 'b']), 'raw': b'\x00\x01', 'pair': (1, 2)}
assert loads(dumps(v)) == v
```

**Deterministic, usable as a cache key**

```python
load('serial', 'dumps')
print(dumps({'b': 2, 'a': 1}))
# Output: {"a":1,"b":2}
```

**Handle failure without aborting**

```python
load('serial', 'try_dumps')
out, err = try_dumps(lambda x: x)
print(out, err)
# Output: None serial.dumps: cannot serialize function: it is code — store the .star script and load() it instead
```
