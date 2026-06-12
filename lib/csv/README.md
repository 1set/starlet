# csv

csv parses and writes comma-separated values files (csv).

Every function has a `try_`-prefixed variant (`try_read_all`, `try_read_dict`, `try_write_all`, `try_write_dict`) that never aborts the script: it returns a `(value, error)` pair where exactly one side is `None`, the same shape as the `json` module's `try_*` functions.

## Functions

### `read_all(source, comma=",", comment="", lazy_quotes=False, trim_leading_space=False, fields_per_record=0, skip=0, limit=0) [][]string`

read all rows from a source string, returning a list of string lists

Rows are read one at a time: a positive `limit` stops parsing at that many rows (malformed content beyond the limit is never reached), and rows consumed by `skip` do not pin the expected field count when `fields_per_record` is 0 — the first kept row does.

#### Parameters

| name                 | type     | description                                                                                                                                                                                                                                                                                                                                                                                                             |
|----------------------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `source`             | `string` | input string of csv data                                                                                                                                                                                                                                                                                                                                                                                                |
| `comma`              | `string` | comma is the field delimiter, defaults to "," (a comma). comma must be a valid character and must not be \r, \n, or the Unicode replacement character (0xFFFD).                                                                                                                                                                                                                                                         |
| `comment`            | `string` | comment, if not "", is the comment character. Lines beginning with the comment character without preceding whitespace are ignored. With leading whitespace the comment character becomes part of the field, even if trim_leading_space is True. comment must be a valid character and must not be \r, \n, or the Unicode replacement character (0xFFFD). It must also not be equal to comma.                            |
| `lazy_quotes`        | `bool`   | If lazy_quotes is True, a quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field.                                                                                                                                                                                                                                                                                                   |
| `trim_leading_space` | `bool`   | If trim_leading_space is True, leading white space in a field is ignored. This is done even if the field delimiter, comma, is white space.                                                                                                                                                                                                                                                                              |
| `fields_per_record`  | `int`    | fields_per_record is the number of expected fields per record. If fields_per_record is positive, read_all requires each record to have the given number of fields. If fields_per_record is 0, read_all sets it to the number of fields in the first record, so that future records must have the same field count. If fields_per_record is negative, no check is made and records may have a variable number of fields. |
| `skip`               | `int`    | Number of rows to skip before starting to read, omitting from returned rows.                                                                                                                                                                                                                                                                                                                                            |
| `limit`              | `int`    | Maximum number of rows to read, stops reading when this limit is reached. If limit is 0, all rows after skip are read.                                                                                                                                                                                                                                                                                                  |

#### Examples

**basic**

read a csv string into a list of string lists

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

**skip_and_limit**

read a csv string with skip and limit

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

### `read_dict(source, comma=",", comment="", lazy_quotes=False, trim_leading_space=False, fields_per_record=0, skip=0, limit=0) []dict`

read csv data whose first row (after `skip`) is the header, returning a list of dicts keyed by the header fields

Takes the same parameters as `read_all`. `limit` counts data rows (the header is not included). A duplicate header field is an error; an empty source yields an empty list. With the default `fields_per_record=0` every data row must have as many fields as the header.

#### Examples

**basic**

read a csv string into a list of dicts

```python
load("csv", "read_dict")
data_str = """type,name,number_of_legs
dog,spot,4
cat,spot,3
"""
data = read_dict(data_str)
print(data)
# Output: [{"type": "dog", "name": "spot", "number_of_legs": "4"}, {"type": "cat", "name": "spot", "number_of_legs": "3"}]
```

### `write_all(source, comma=",") string`

write all rows from source to a csv-encoded string

Cell values are rendered per type: strings as-is; ints and floats in plain decimal notation (never scientific, e.g. `1000000.0` → `1000000`); `True`/`False` as `true`/`false`; `None` as an empty cell; time values as RFC 3339. Nested lists/dicts and non-finite floats (`nan`, `inf`) are reported as errors instead of being silently written in Go syntax.

#### Parameters

| name     | type         | description                                                                                                                                                     |
|----------|--------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `source` | `[][]string` | array of arrays of strings to write to csv                                                                                                                      |
| `comma`  | `string`     | comma is the field delimiter, defaults to "," (a comma). comma must be a valid character and must not be \r, \n, or the Unicode replacement character (0xFFFD). |

#### Examples

**basic**

write a list of string lists to a csv string

```python
load("csv", "write_all")
data = [
["type", "name", "number_of_legs"],
["dog", "spot", "4"],
["cat", "spot", "3"],
["spider", "samantha", "8"],
]
csv_str = write_all(data)
print(csv_str)
# Output: "type,name,number_of_legs\ndog,spot,4\ncat,spot,3\nspider,samantha,8\n"
```

### `write_dict(data, header, comma=",") string`

write a list of dictionaries to a csv string based on the provided header

Cell values are rendered with the same per-type rules as `write_all`; a key missing from a row produces an empty cell, the same as an explicit `None`.

#### Parameters

| name     | type       | description                                                                                                                                                     |
|----------|------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `data`   | `[]dict`   | array of dictionaries where each dictionary is a row with field names as keys                                                                                   |
| `header` | `[]string` | array of strings representing the header (column names) of the csv                                                                                              |
| `comma`  | `string`   | comma is the field delimiter, defaults to "," (a comma). comma must be a valid character and must not be \r, \n, or the Unicode replacement character (0xFFFD). |

#### Examples

**basic**

write a list of dictionaries to a csv string based on header

```python
load("csv", "write_dict")
data = [
{"type": "dog", "name": "spot", "number_of_legs": 4},
{"type": "cat", "name": "spot", "number_of_legs": 3},
{"type": "spider", "name": "samantha", "number_of_legs": 8},
]
csv_str = write_dict(data, header=["type", "name", "number_of_legs"])
print(csv_str)
# Output: "type,name,number_of_legs\ndog,spot,4\ncat,spot,3\nspider,samantha,8\n"
```
