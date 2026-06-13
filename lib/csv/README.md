# csv

`csv` parses and writes comma-separated values, mirroring a subset of Go's `encoding/csv`: read a CSV string into rows or header-keyed dicts, and write rows or dicts back to a CSV string. Capability profile: **Pure** — it operates only on in-memory strings and has no filesystem, network, process, or log side effects.

Every function has a `try_`-prefixed variant (`try_read_all`, `try_read_dict`, `try_write_all`, `try_write_dict`) that never aborts the script: it returns a `(value, error)` tuple where exactly one side is `None` (the value on success, the error string on failure), the same shape as the `json` module's `try_*` functions.

## Functions

| function | description |
|----------|-------------|
| `read_all(source, comma=",", comment="", lazy_quotes=False, trim_leading_space=False, fields_per_record=0, skip=0, limit=0) -> list` / `try_read_all(...) -> tuple` | read all rows from a CSV string into a list of string lists |
| `read_dict(source, comma=",", comment="", lazy_quotes=False, trim_leading_space=False, fields_per_record=0, skip=0, limit=0) -> list` / `try_read_dict(...) -> tuple` | read a CSV string whose first row (after `skip`) is the header into a list of dicts keyed by the header fields |
| `write_all(data, comma=",") -> string` / `try_write_all(...) -> tuple` | write a list of lists to a CSV-encoded string |
| `write_dict(data, header, comma=",") -> string` / `try_write_dict(...) -> tuple` | write a list of dicts to a CSV-encoded string using the given header columns |

The `try_*` variants take the same arguments as their base function and return a `(value, error)` tuple: `(result, None)` on success, `(None, "<error message>")` on failure. They never raise.

## Constants

| constant | meaning |
|----------|---------|
| `ModuleName` | Go-side constant (`"csv"`) naming this module for `load()`; not a script-visible member. |

## Reading

### `read_all`

`read_all(source, comma=",", comment="", lazy_quotes=False, trim_leading_space=False, fields_per_record=0, skip=0, limit=0) -> list`

Reads `source` and returns a list of rows, each row a list of strings. A UTF-8 BOM at the start of `source` is stripped. `source` may be a string or bytes.

Parameters:

| name | type | description |
|------|------|-------------|
| `source` | `string`/`bytes` | input CSV data |
| `comma` | `string` | field delimiter, a single character; defaults to `","`. Must not be `\r`, `\n`, or U+FFFD. |
| `comment` | `string` | if not `""`, a single comment character; lines beginning with it (without preceding whitespace) are skipped. Must not equal `comma`. |
| `lazy_quotes` | `bool` | if `True`, a quote may appear in an unquoted field and a non-doubled quote in a quoted field |
| `trim_leading_space` | `bool` | if `True`, leading white space in a field is ignored |
| `fields_per_record` | `int` | expected fields per record: positive requires exactly that many; `0` (default) pins the count to the first kept record; negative disables the check (rows may vary in length) |
| `skip` | `int` | number of rows to skip before reading; skipped rows are omitted from the result |
| `limit` | `int` | maximum number of rows to return; `0` reads all rows after `skip` |

Rows are read one at a time: a positive `limit` stops parsing at that many rows, so malformed content beyond the limit is never reached. Rows consumed by `skip` do not pin the expected field count when `fields_per_record` is `0` — the first kept row does.

Errors on: a `comma` or `comment` that is not exactly one character (or `comment` equal to `comma`); a record whose field count violates `fields_per_record`; a malformed/unterminated-quote parse error that is actually reached. An empty `source` returns `[]`.

```python
load("csv", "read_all")
data_str = """type,name,number_of_legs
dog,spot,4
cat,spot,3
spider,samantha,8
"""
data = read_all(data_str)
print(data)
# Output: [["type", "name", "number_of_legs"], ["dog", "spot", "4"], ["cat", "spot", "3"], ["spider", "samantha", "8"]]
```

With `skip` and `limit` (skip the header, keep one data row):

```python
load("csv", "read_all")
data_str = """type,name,number_of_legs
dog,spot,4
cat,spot,3
spider,samantha,8
"""
data = read_all(data_str, skip=1, limit=1)
print(data)
# Output: [["dog", "spot", "4"]]
```

With a custom delimiter and comment character:

```python
load("csv", "read_all")
csv_string = """a|b|c
#1,2,3
4|5|6
7|8|9
"""
print(read_all(csv_string, comma="|", comment="#"))
# Output: [["a", "b", "c"], ["4", "5", "6"], ["7", "8", "9"]]
```

A malformed row past `limit` is never reached, so this succeeds:

```python
load("csv", "read_all")
csv_string = 'a,b\nc,d\n"bad\n'
print(read_all(csv_string, limit=2))
# Output: [["a", "b"], ["c", "d"]]
```

### `read_dict`

`read_dict(source, comma=",", comment="", lazy_quotes=False, trim_leading_space=False, fields_per_record=0, skip=0, limit=0) -> list`

Takes the same parameters as `read_all`. The first remaining row (after `skip`) is treated as the header; each subsequent row becomes a dict keyed by the header fields. `limit` counts data rows only — the header is not included.

Errors on: a duplicate header field; the same `comma`/`comment`/parse/`fields_per_record` errors as `read_all`. An empty `source` returns `[]`. With `fields_per_record=-1`, rows shorter than the header simply omit the missing keys, and cells beyond the header are dropped.

```python
load("csv", "read_dict")
data_str = """a,b
1,2
3,4
"""
print(read_dict(data_str))
# Output: [{"a": "1", "b": "2"}, {"a": "3", "b": "4"}]
```

Variable-length rows with `fields_per_record=-1` map only the fields present:

```python
load("csv", "read_dict")
print(read_dict('a\n1,2\n3\n', fields_per_record=-1))
# Output: [{"a": "1"}, {"a": "3"}]
```

## Writing

Cell values are rendered by type, never with Go's default formatting:

- `string` — written as-is
- `int` / `float` — plain decimal, never scientific notation (`1000000.0` → `1000000`, `0.00001` → `0.00001`)
- `bool` — `true` / `false` (lowercase, matching `json.encode`)
- `None` — an empty cell
- `time` value — RFC 3339 (e.g. `2023-01-15T12:30:45Z`)

A non-finite float (`nan`, `inf`) errors (`float value ... is not representable in CSV`); any other type, including nested lists/dicts, errors (`unsupported cell type ...`) rather than being written as Go syntax.

### `write_all`

`write_all(data, comma=",") -> string`

`data` must be a list of lists (rows of cells). Returns the CSV text, each record terminated by `\n`.

Errors on: a `comma` that is not exactly one character; `data` that is not an array; a row that is not an array; a cell of an unsupported or non-finite-float type. An unconvertible Starlark value (e.g. a function) errors at the conversion step (`unrecognized starlark type: ...`).

```python
load("csv", "write_all")
data = [
    ["type", "name", "number_of_legs"],
    ["dog", "spot", "4"],
    ["cat", "spot", "3"],
    ["spider", "samantha", "8"],
]
print(write_all(data))
# Output: type,name,number_of_legs
# dog,spot,4
# cat,spot,3
# spider,samantha,8
#
```

Per-type cell rendering:

```python
load("csv", "write_all")
print(write_all([[1000000.0, 0.00001, -2.5]]))
print(write_all([[None, True, False]]))
# Output: 1000000,0.00001,-2.5
#
# ,true,false
#
```

### `write_dict`

`write_dict(data, header, comma=",") -> string`

`data` must be a list of dicts; `header` is an iterable of column-name strings. The header row is written first, then one row per dict, taking cells in `header` order. A key missing from a dict produces an empty cell (the same as an explicit `None`); dict keys not in `header` are ignored.

Errors on: an empty `header`; a `header` element that is not a string; a `comma` that is not exactly one character; `data` that is not an array; an element of `data` that is not a dict; a cell of an unsupported or non-finite-float type.

```python
load("csv", "write_dict")
data = [
    {"type": "dog", "name": "spot", "number_of_legs": 4},
    {"type": "cat", "name": "spot", "number_of_legs": 3},
    {"type": "spider", "name": "samantha", "number_of_legs": 8},
]
print(write_dict(data, header=["type", "name", "number_of_legs"]))
# Output: type,name,number_of_legs
# dog,spot,4
# cat,spot,3
# spider,samantha,8
#
```

Missing keys and extra keys (`number_of_legs` is absent in the second row; `C` is not in `header`):

```python
load("csv", "write_dict")
x = write_dict([{"a": 200, "b": 100, "c": 500}, {"b": 1024, "C": 2048}], header=["c", "b"])
print(x)
# Output: c,b
# 500,100
# ,1024
#
```

## try_* variants

Each base function has a `try_`-prefixed variant — `try_read_all`, `try_read_dict`, `try_write_all`, `try_write_dict` — that returns `(value, None)` on success and `(None, "<error message>")` on failure instead of aborting the script. Argument-unpacking errors are captured the same way.

```python
load("csv", "try_read_all")
rows, err = try_read_all('a,b\nc,d\n')
print(rows, err)
bad, err2 = try_read_all('"bad\n')
print(bad, "parse error" in err2)
# Output: [["a", "b"], ["c", "d"]] None
# None True
```

```python
load("csv", "try_write_all")
text, err = try_write_all([[1, 2]])
print(text, err)
bad, err2 = try_write_all([[[1]]])
print(bad, "unsupported cell type" in err2)
# Output: 1,2
#  None
# None True
```

## Notes / boundaries

- **Engine.** Backed by Go's `encoding/csv` (RFC 4180 semantics). Reading transparently strips a leading UTF-8 BOM and normalizes lone `\r` line endings.
- **Pure.** No filesystem, network, process, or logging effects; functions operate only on the strings/bytes passed in and return strings or values.
- **Honest boundary.** Writing never silently corrupts: unsupported cell types and non-finite floats raise rather than emitting Go-syntax text. Use `try_*` to handle failures inline instead of aborting.
- **Reading returns strings.** Every cell from `read_all` / `read_dict` is a string; no numeric or boolean coercion is performed on input.
- **All member names are snake_case** — no non-standard identifiers in this module.
