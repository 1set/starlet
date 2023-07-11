# JSON

`json` defines utilities for converting Starlark values to/from JSON strings. The most recent IETF standard for JSON is https://www.ietf.org/rfc/rfc7159.txt .

## Functions

### `encode(x) string`

The encode function accepts one required positional argument, which it converts to JSON by cases:
- A Starlark value that implements Go's standard `json.Marshal` interface defines its own JSON encoding.
- `None`, `True`, and `False` are converted to `null`, `true`, and `false`, respectively.
- Starlark int values, no matter how large, are encoded as decimal integers. Some decoders may not be able to decode very large integers.
- Starlark float values are encoded using decimal point notation, even if the value is an integer. It is an error to encode a non-finite floating-point value.
- Starlark strings are encoded as JSON strings, using UTF-16 escapes.
- a Starlark IterableMapping (e.g. dict) is encoded as a JSON object. It is an error if any key is not a string.
- any other Starlark Iterable (e.g. list, tuple) is encoded as a JSON array.
- a Starlark HasAttrs (e.g. struct) is encoded as a JSON object.
  It an application-defined type matches more than one the cases describe above, (e.g. it implements both `Iterable` and `HasFields`), the first case takes precedence. Encoding any other value yields an error.

#### Examples

**Basic**

Encode a Starlark dict to a JSON string.

```python
load('json', 'encode')
print(encode({'a': 1, 'b': 2}))
# Output: {"a":1,"b":2}
```

### `decode(x[, default]) string`

The decode function has one required positional parameter, a JSON string. It returns the Starlark value that the string denotes.
- Numbers are parsed as int or float, depending on whether they contain a decimal point.
- JSON objects are parsed as new unfrozen Starlark dicts.
- JSON arrays are parsed as new unfrozen Starlark lists.
  If x is not a valid JSON string, the behavior depends on the "default" parameter: if present, Decode returns its value; otherwise, Decode fails.

#### Examples

**Basic**

Decode a JSON string to a Starlark dict.

```python
load('json', 'decode')
print(decode('{"a":10,"b":20}'))
# Output: {'a': 10, 'b': 20}
```

### `indent(str, *, prefix="", indent="\t") string`

The indent function pretty-prints a valid JSON encoding, and returns a string containing the indented form.
It accepts one required positional parameter, the JSON string, and two optional keyword-only string parameters, prefix and indent, that specify a prefix of each new line, and the unit of indentation.

#### Examples

**Basic**

Indent a JSON string.

```python
load('json', 'indent')
print(indent('{"a":10,"b":20}', indent="  "))
# Output:
# {
#   "a": 10,
#   "b": 20
# }
```

### `dumps(obj, indent=0) string`

The dumps function converts a Starlark value to a JSON string, and returns it.
It accepts one required positional parameter, the Starlark value, and one optional integer parameter, indent, that specifies the unit of indentation.

#### Examples

**Basic**

Dump a Starlark dict to a JSON string with indentation.

```python
load('json', 'dumps')
print(dumps({'a': 10, 'b': 20}, indent=2))
# Output:
# {
#   "a": 10,
#   "b": 20
# }
```
