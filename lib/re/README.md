# re

`re` defines regular expression functions, it's intended to be a subset of [Python's **re** module](https://docs.python.org/3/library/re.html) for Starlark, built on Go's [RE2 syntax](https://golang.org/s/re2syntax).

Notable differences from Python's `re`:

- **flags must be `0`**: numeric `re.*` flags are not supported and are rejected with an error. Use inline pattern flags like `(?i)`, `(?m)`, `(?s)` instead.
- **`sub` replacement templates use Go syntax**: `$1` or `${name}` refers to a capture group and `$$` is a literal dollar sign. Python backslash references like `\1` are **not** interpreted, and a function `repl` is not supported.
- **`split` does not include capture-group text** in its result.
- There is no Match object: `match` returns a list of tuples and `search` returns an index pair (see below).

## Functions

### `compile(pattern, flags=0) Pattern`

Compile a regular expression pattern into a regular expression object, which
can be used for matching using its match(), search() and other methods.

#### Parameters

| name      | type     | description                                            |
|-----------|----------|--------------------------------------------------------|
| `pattern` | `string` | regular expression pattern string                      |
| `flags`   | `int`    | must be 0; use inline flags like `(?i)` in the pattern |

### `search(pattern, string, flags=0)`

Scan through string looking for the first location where the regular expression pattern
produces a match, and return the `[start, end]` byte-index pair of that match as a list.
Return None if no position in the string matches the pattern; note that this is different
from finding a zero-length match at some point in the string.

#### Parameters

| name      | type     | description                                            |
|-----------|----------|--------------------------------------------------------|
| `pattern` | `string` | regular expression pattern string                      |
| `string`  | `string` | input string to search                                 |
| `flags`   | `int`    | must be 0; use inline flags like `(?i)` in the pattern |

### `findall(pattern, text, flags=0)`

Returns all non-overlapping matches of pattern in string, as a tuple of strings.
The string is scanned left-to-right, and matches are returned in the order found.
If one group is present in the pattern, the group text is returned instead of the
full match; with several groups, each element is a tuple of the group texts.
Empty matches are included in the result.

#### Parameters

| name      | type     | description                                            |
|-----------|----------|--------------------------------------------------------|
| `pattern` | `string` | regular expression pattern string                      |
| `text`    | `string` | string to find within                                  |
| `flags`   | `int`    | must be 0; use inline flags like `(?i)` in the pattern |

### `split(pattern, text, maxsplit=0, flags=0)`

Split text by the occurrences of pattern. If maxsplit is positive, at most maxsplit
splits occur, and the remainder of the string is returned as the final element of the
result; a negative maxsplit means no splits happen at all. Note that unlike Python,
the text of capture groups in the pattern is **not** included in the result.

#### Parameters

| name       | type     | description                                                            |
|------------|----------|------------------------------------------------------------------------|
| `pattern`  | `string` | regular expression pattern string                                      |
| `text`     | `string` | input string to split                                                  |
| `maxsplit` | `int`    | maximum number of splits. 0 (default) splits everywhere, negative none |
| `flags`    | `int`    | must be 0; use inline flags like `(?i)` in the pattern                 |

### `sub(pattern, repl, text, count=0, flags=0)`

Return the string obtained by replacing the leftmost non-overlapping occurrences of pattern
in string by the replacement repl. If the pattern isn't found, string is returned unchanged.
repl must be a string and uses Go's template syntax: `$1` or `${name}` refers to a capture
group and `$$` is a literal dollar sign; Python backslash references like `\1` are **not**
interpreted.

#### Parameters

| name      | type     | description                                                                |
|-----------|----------|----------------------------------------------------------------------------|
| `pattern` | `string` | regular expression pattern string                                          |
| `repl`    | `string` | replacement template (`$1`/`${name}` for groups, `$$` for a literal `$`)   |
| `text`    | `string` | input string to replace                                                    |
| `count`   | `int`    | number of replacements. 0 (default) replaces all matches, negative none    |
| `flags`   | `int`    | must be 0; use inline flags like `(?i)` in the pattern                     |

### `match(pattern, string, flags=0)`

If zero or more characters at the **beginning** of string match the regular expression
pattern, return a list with a single tuple holding the full match followed by the text of
every capture group. Return an empty list if the beginning of the string does not match
the pattern — a match elsewhere in the string does not count; use `search()` or
`findall()` for that.

#### Parameters

| name      | type     | description                                            |
|-----------|----------|--------------------------------------------------------|
| `pattern` | `string` | regular expression pattern string                      |
| `string`  | `string` | input string to match                                  |
| `flags`   | `int`    | must be 0; use inline flags like `(?i)` in the pattern |

## Types

### `Pattern`

**Methods**

#### `search(text, flags=0)`

#### `match(text, flags=0)`

#### `findall(text, flags=0)`

#### `split(text, maxsplit=0, flags=0)`

#### `sub(repl, text, count=0, flags=0)`
