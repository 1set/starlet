# re

`re` provides regular-expression functions for Starlark — a small subset of [Python's `re` module](https://docs.python.org/3/library/re.html), built on Go's [RE2 syntax](https://golang.org/s/re2syntax). It is **pure** (no filesystem, network, process, or log side effects).

Capability profile: **Pure** — and **legacy/frozen**: this module is superseded by the `regex` module, which returns Python-shaped `list` results, supports more of the API, and is the one new code should use. `re` is kept only for backward compatibility; no new features are added here.

## Functions

| function | description |
|----------|-------------|
| `compile(pattern, flags=0) -> regexp` | Compile `pattern` into a reusable `regexp` object exposing `match`/`search`/`findall`/`split`/`sub` methods. |
| `search(pattern, string, flags=0) -> list \| None` | Find the first match anywhere in `string`; return its `[start, end]` byte-index pair, or `None` if there is no match. |
| `match(pattern, string, flags=0) -> list` | Match only at the **beginning** of `string`; return `[(full_match, group1, ...)]` on success, or `[]` on no match. |
| `split(pattern, string, maxsplit=0, flags=0) -> tuple` | Split `string` on `pattern`. Capture-group text is **not** included (unlike Python). |
| `findall(pattern, string, flags=0) -> tuple` | Return all non-overlapping matches; full match text with no groups, the single group's text with one group, or a tuple of group texts with several. |
| `sub(pattern, repl, string, count=0, flags=0) -> string` | Replace non-overlapping matches of `pattern` in `string` with the Go-template `repl`. |

## Constants

This module exposes no constants. In particular, no `re.*` flag constants exist — the `flags` parameter must be `0` (see Notes).

## Types

### `regexp`

A compiled pattern returned by `compile`. Its `Type()` string is `regexp`. It is immutable (frozen) and hashable. Each method mirrors the module-level function of the same name but drops the leading `pattern` argument.

| method | signature | description |
|--------|-----------|-------------|
| `search` | `r.search(string, flags=0) -> list \| None` | As `search`, against the compiled pattern. |
| `match` | `r.match(string, flags=0) -> list` | As `match`, anchored at the start. |
| `findall` | `r.findall(string, flags=0) -> tuple` | As `findall`. |
| `split` | `r.split(string, maxsplit=0, flags=0) -> tuple` | As `split`. |
| `sub` | `r.sub(repl, string, count=0, flags=0) -> string` | As `sub`. |

`dir()` on a compiled object lists exactly `["findall", "match", "search", "split", "sub"]`.

## Details & examples

### `compile(pattern, flags=0) -> regexp`

Compile a pattern once and reuse it. Errors if `pattern` is not a string, if it is not valid RE2, or if `flags` is non-zero.

```python
load('re', 'compile')
foo_r = compile("foo")
print(foo_r.findall("foo bar baz"))
# Output:
# ("foo",)
```

### `search(pattern, string, flags=0) -> list | None`

Scans the whole string for the first match and returns its `[start, end]` byte indices, or `None` when nothing matches. Errors if `pattern`/`string` is not a string, if `pattern` is invalid RE2, or if `flags` is non-zero. Note this returns an index pair, not match text (use `findall` for text).

```python
load('re', 'compile')
b = compile('b')
print(b.search('abc'))
print(b.search('xyz'))
# Output:
# [1, 2]
# None
```

### `match(pattern, string, flags=0) -> list`

Matches only at the **beginning** of the string (Python `re.match` semantics — a match elsewhere does not count). On success returns a one-element list holding a tuple of the full match followed by each capture group; on failure returns an empty list `[]` (which is falsy, so `if match(...)` works as ported from Python). Same error conditions as `search`.

```python
load('re', 'match')
print(match('world', 'hello world'))
print(match('hello', 'hello world'))
print(match('(h)(e)', 'hello'))
# Output:
# []
# [("hello",)]
# [("he", "h", "e")]
```

### `split(pattern, string, maxsplit=0, flags=0) -> tuple`

Splits on `pattern`. `maxsplit=0` (default) splits everywhere; a positive `maxsplit` keeps the remainder as the final element; a negative `maxsplit` performs no split at all. An astronomically large `maxsplit` is treated as "no limit". Capture-group text is **not** inserted into the result (a deliberate difference from Python). Same error conditions as `search`.

```python
load('re', 'split')
print(split(',', 'a,b,c', 1))
print(split(',', 'a,b,c', maxsplit=2))
print(split(',', 'a,b,c', -1))
# Output:
# ("a", "b,c")
# ("a", "b", "c")
# ("a,b,c",)
```

### `findall(pattern, string, flags=0) -> tuple`

Returns all non-overlapping matches, left to right. With no capture groups each element is the full match text; with one group it is that group's text; with several groups it is a tuple of the group texts. Empty matches are included. Returns an empty tuple when nothing matches. Same error conditions as `search`.

```python
load('re', 'findall')
print(findall(r'(\w)(\d)', 'a1 b2'))
print(findall(r'\w(\d)', 'a1 b2'))
print(findall('foo', 'bar baz'))
# Output:
# (("a", "1"), ("b", "2"))
# ("1", "2")
# ()
```

### `sub(pattern, repl, string, count=0, flags=0) -> string`

Replaces non-overlapping matches with `repl`. `count=0` (default) replaces all; a positive `count` replaces only the first `count`; a negative `count` replaces nothing (returns the input unchanged). `repl` uses **Go template syntax**: `$1` or `${name}` refers to a capture group and `$$` is a literal `$` — Python backslash references like `\1` are **not** interpreted, and a function `repl` is not supported. Errors if any of `pattern`/`repl`/`string` is not a string, if `pattern` is invalid RE2, or if `flags` is non-zero.

```python
load('re', 'sub')
print(sub('a', 'X', 'aaa', 1))
print(sub('a', 'X', 'aaa', count=2))
print(sub('a', 'X', 'aaa', -1))
print(sub('(a)(b)', '${2}${1}', 'ab ab'))
print(sub('a', '$$5', 'a'))
# Output:
# Xaa
# XXa
# aaa
# ba ba
# $5
```

## Notes / boundaries

- **Engine:** Go's `regexp` (RE2). RE2 guarantees linear-time matching but has no backreferences or lookaround; patterns needing those fail to compile with an error rather than being silently approximated.
- **`flags` must be `0`.** This module exports no flag constants. Any non-zero `flags` (e.g. a ported `re.IGNORECASE == 2`) is rejected with an error — historically it was silently ignored, which matched with default behavior and was wrong. Use **inline pattern flags** like `(?i)`, `(?m)`, `(?s)` instead.
- **Return shapes differ from CPython** by design: there is no `Match` object; `match` returns a list of tuples, `search` returns a `[start, end]` index pair, and `findall`/`split` return **tuples**. The successor `regex` module returns Python-shaped lists — prefer it for new code.
- **`split` drops capture-group text**, unlike Python's `re.split`.
- **`sub` templates are Go-style** (`$1`, `${name}`, `$$`), not Python-style (`\1`); `repl` must be a string.
- **Determinism:** matching is deterministic; results depend only on the pattern and input string.
- **Not implemented** from Python's `re`: `fullmatch`, `finditer`, `subn`, `escape`, and numeric flag constants.
