# json

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

### `try_dumps(obj, indent=0) tuple`

The try_dumps function is a variant of dumps that handles errors gracefully.
It accepts the same parameters as dumps, but returns a tuple of (result, error).
If successful, error will be None. If an error occurs, result will be None and error will contain the error message.

#### Examples

**Basic**

Try to dump a Starlark dict to a JSON string and handle potential errors.

```python
load('json', 'try_dumps')
result, error = try_dumps({'a': 10, 'b': 20}, indent=2)
print("Result:", result)
print("Error:", error)
# Output:
# Result: {
#   "a": 10,
#   "b": 20
# }
# Error: None
```

### `try_encode(x) tuple`

The try_encode function is a variant of encode that handles errors gracefully.
It accepts the same parameter as encode, but returns a tuple of (result, error).
If successful, error will be None. If an error occurs, result will be None and error will contain the error message.

#### Examples

**Basic**

Try to encode a Starlark dict to a JSON string and handle potential errors.

```python
load('json', 'try_encode')
result, error = try_encode({'a': 10, 'b': 20})
print("Result:", result)
print("Error:", error)
# Output:
# Result: {"a":10,"b":20}
# Error: None
```

### `try_decode(x) tuple`

The try_decode function is a variant of decode that handles errors gracefully.
It accepts the same parameter as decode, but returns a tuple of (result, error).
If successful, error will be None. If an error occurs, result will be None and error will contain the error message.

#### Examples

**Basic**

Try to decode a JSON string to a Starlark dict and handle potential errors.

```python
load('json', 'try_decode')
result, error = try_decode('{"a":10,"b":20}')
print("Result:", result)
print("Error:", error)
# Output:
# Result: {'a': 10, 'b': 20}
# Error: None
```

### `try_indent(str, prefix="", indent="\t") tuple`

The try_indent function is a variant of indent that handles errors gracefully.
It accepts the same parameters as indent, but returns a tuple of (result, error).
If successful, error will be None. If an error occurs, result will be None and error will contain the error message.

#### Examples

**Basic**

Try to indent a JSON string and handle potential errors.

```python
load('json', 'try_indent')
result, error = try_indent('{"a":10,"b":20}', indent="  ")
print("Result:", result)
print("Error:", error)
# Output:
# Result: {
#   "a": 10,
#   "b": 20
# }
# Error: None
```

### `path(data, path) list`

The path function performs a JSONPath query on the given JSON data and returns the matching elements.
It accepts two positional arguments:
- data: JSON data as a string, bytes, or Starlark value (dict, list, etc.)
- path: A JSONPath expression string
  It returns a list of matching elements. If no matches are found, an empty list is returned.
  If the JSONPath expression is invalid, an error is raised.

#### Examples

**Basic**

Query JSON data using JSONPath expressions.

```python
load('json', 'path')
data = '''{"store":{"book":[{"title":"Moby Dick","price":8.99},{"title":"War and Peace","price":12.99}]}}'''
titles = path(data, '$.store.book[*].title')
print(titles)
# Output: ['Moby Dick', 'War and Peace']
prices = path(data, '$..price')
print(prices)
# Output: [8.99, 12.99]
```

### `try_path(data, path) tuple`

The try_path function is a variant of path that handles errors gracefully.
It accepts the same parameters as path, but returns a tuple of (result, error).
If successful, error will be None. If an error occurs, result will be None and error will contain the error message.

#### Examples

**Basic**

Try to query JSON data using JSONPath and handle potential errors.

```python
load('json', 'try_path')
data = '''{"store":{"book":[{"title":"Moby Dick","price":8.99},{"title":"War and Peace","price":12.99}]}}'''
result, error = try_path(data, '$..price')
print("Result:", result)
print("Error:", error)
# Output:
# Result: [8.99, 12.99]
# Error: None
```

### `eval(data, expr) value`

The eval function evaluates a JSONPath expression on the given JSON data and returns the evaluation result.
It accepts two positional arguments:
- data: JSON data as a string, bytes, or Starlark value (dict, list, etc.)
- expr: A JSONPath expression string to evaluate
  It returns the result of the evaluation, which can be a number, string, boolean, list, dict, or None.
  If the expression is invalid, an error is raised.

#### Examples

**Basic**

Evaluate JSONPath expressions on JSON data.

```python
load('json', 'eval')
data = '''{"store":{"book":[{"price":8.99},{"price":12.99},{"price":5.99}]}}'''
avg_price = eval(data, 'avg($..price)')
print(avg_price)
# Output: 9.323333333333334
sum_price = eval(data, 'sum($..price)')
print(sum_price)
# Output: 27.97
```

### `try_eval(data, expr) tuple`

The try_eval function is a variant of eval that handles errors gracefully.
It accepts the same parameters as eval, but returns a tuple of (result, error).
If successful, error will be None. If an error occurs, result will be None and error will contain the error message.

#### Examples

**Basic**

Try to evaluate JSONPath expressions on JSON data and handle potential errors.

```python
load('json', 'try_eval')
data = '''{"store":{"book":[{"price":8.99},{"price":12.99},{"price":5.99}]}}'''
result, error = try_eval(data, 'avg($..price)')
print("Result:", result)
print("Error:", error)
# Output:
# Result: 9.323333333333334
# Error: None
```
