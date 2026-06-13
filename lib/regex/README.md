# regex

`regex` provides regular-expression functions for Starlark — a subset of [Python's **re** module](https://docs.python.org/3/library/re.html) backed by Go's [RE2 engine](https://golang.org/s/re2syntax). **Capability profile: Pure** (no filesystem, network, process, or log side effects; deterministic for a given pattern and input).

It succeeds the legacy `re` module (which is frozen): `regex` adds `Match` objects, named-group extraction, the `IGNORECASE`/`MULTILINE`/`DOTALL` flags, `\1`/`\g<name>` and function replacements, and the full `compile`/`fullmatch`/`finditer`/`subn`/`escape` surface. RE2 matches in linear time — no catastrophic backtracking / ReDoS — which suits running untrusted or LLM-generated scripts in a sandbox. Where RE2 genuinely differs from Python — lookahead/lookbehind (`(?=...)`, `(?<=...)`) and in-pattern backreferences (`\1`) — the pattern fails to **compile** with a clear error rather than silently misbehaving.

## Functions

| function | description |
|---|---|
| `compile(pattern, flags=0) -> Pattern` | compile a pattern into a reusable `Pattern` object |
| `search(pattern, string, flags=0) -> Match` | first match anywhere → `Match`, or `None` |
| `match(pattern, string, flags=0) -> Match` | match anchored at the **start** → `Match`, or `None` |
| `fullmatch(pattern, string, flags=0) -> Match` | match the **whole** string → `Match`, or `None` |
| `findall(pattern, string, flags=0) -> list` | all matches as a list (Python group shaping, see below) |
| `finditer(pattern, string, flags=0) -> tuple` | a tuple of `Match` objects |
| `sub(pattern, repl, string, count=0, flags=0) -> str` | replace matches → the result string |
| `subn(pattern, repl, string, count=0, flags=0) -> (str, int)` | replace → `(result, num_replacements)` |
| `split(pattern, string, maxsplit=0, flags=0) -> list` | split into a list, **including capture-group text** (Python semantics) |
| `escape(pattern) -> str` | escape regex metacharacters in `pattern` |
| `try_compile(...) -> (value, error)`, `try_search(...) -> (value, error)` | non-raising variants of `compile` / `search`: return a `(value, error)` pair (error is `None` on success) instead of aborting the script — the same shape as the `json`/`csv`/`http` modules |

**findall shaping** (matches Python): the result is always a **list**; no capture group → the full match text; one group → that group's text; two or more groups → a **tuple** of the group texts per match. `split` likewise returns a **list** (same as Python's `re` and starlet's legacy `re`).

## Constants

Integer flag constants (Python `re` values), OR-able with `|` and passed as the `flags` argument; they translate to RE2 inline flags (`(?i)`, `(?m)`, `(?s)`). Each flag has a short and a long spelling that are equal.

| constant | meaning |
|---|---|
| `I` / `IGNORECASE` | case-insensitive matching (`(?i)`, value `2`) |
| `M` / `MULTILINE` | `^`/`$` match at line boundaries (`(?m)`, value `8`) |
| `S` / `DOTALL` | `.` matches newlines too (`(?s)`, value `16`) |

A `flags` value outside the supported set (e.g. `1024`) errors with `unsupported flags value`.

## Types

### `Pattern`

A compiled regular expression — the value returned by `compile` (`type` is `regex.Pattern`). Its methods mirror the module-level functions **without** the leading `pattern` argument. `Pattern` is hashable, so it may be used as a dict key.

| member | signature | description |
|---|---|---|
| `search` | `p.search(string) -> Match` | first match anywhere → `Match`, or `None` |
| `match` | `p.match(string) -> Match` | match anchored at the start → `Match`, or `None` |
| `fullmatch` | `p.fullmatch(string) -> Match` | match the whole string → `Match`, or `None` |
| `findall` | `p.findall(string) -> list` | all matches (Python group shaping) |
| `finditer` | `p.finditer(string) -> tuple` | a tuple of `Match` objects |
| `split` | `p.split(string, maxsplit=0) -> list` | split, including capture-group text |
| `sub` | `p.sub(repl, string, count=0) -> str` | replace matches → string |
| `subn` | `p.subn(repl, string, count=0) -> (str, int)` | replace → `(result, count)` |
| `pattern` | attribute → `str` | the source pattern string |
| `flags` | attribute → `int` | the flags passed at compile time |
| `groups` | attribute → `int` | number of capture groups |

### `Match`

The result of a successful `search` / `match` / `fullmatch` (`type` is `regex.Match`). `Match` objects are **unhashable** — using one as a dict key errors with `unhashable type: regex.Match`. A non-participating optional group reports `None` (or the supplied default).

| member | signature | description |
|---|---|---|
| `group` | `m.group(n=0, ...) -> str` | group text by index or name; no arg → group 0 (the whole match); several args → a tuple |
| `groups` | `m.groups(default=None) -> tuple` | a tuple of all capture groups; `default` substitutes for non-participating groups |
| `groupdict` | `m.groupdict(default=None) -> dict` | a dict of named groups → their text |
| `start` | `m.start(group=0) -> int` | start byte index of a group (index or name) |
| `end` | `m.end(group=0) -> int` | end byte index of a group |
| `span` | `m.span(group=0) -> (int, int)` | `(start, end)` of a group |
| `expand` | `m.expand(template) -> str` | apply a `\1`/`\g<name>` template against the match |
| `string` | attribute → `str` | the subject string that was matched |
| `re` | attribute → `Pattern` | the `Pattern` that produced the match |

## Details & examples

### Matching: `search`, `match`, `fullmatch`

`search` finds the first match anywhere; `match` anchors at the start of the string; `fullmatch` requires the whole string to match. Each returns a `Match` on success or `None` on no match. All three error with `cannot compile pattern` if `pattern` is invalid or uses an RE2-unsupported construct (lookaround, backreferences).

```python
load("regex", "search", "match", "fullmatch")
m = search(r'(\w+)@(\w+)', 'reach ann@host now')
print(m.group(0), m.group(1), m.group(2), m.span(0))
print(match('world', 'hello world'))
print(fullmatch('a+', 'aaa').group(0))
# Output:
# ann@host ann host (6, 14)
# None
# aaa
```

### Named groups

Groups can be addressed by name (`(?P<name>...)`) as well as index, in `group`, `start`, `end`, and `span`; `groupdict` returns just the named groups.

```python
load("regex", "search")
m = search(r'(?P<user>\w+)@(?P<host>\w+)', 'ann@example')
print(m.group('user'), m.group('host'))
print(m.groupdict())
print(m.group('user', 'host'))
# Output:
# ann example
# {"user": "ann", "host": "example"}
# ("ann", "example")
```

`m.group` / `m.start` / `m.span` error with `no such group` for an out-of-range index or unknown name; passing a non-int/non-string selector errors with `group index must be an int or string`.

```python
load("regex", "search")
m = search(r'(a)(b)?', 'a')
print(m.groups())
print(m.groups('X'))
# Output:
# ("a", None)
# ("a", "X")
```

### `findall` and `finditer`

`findall` returns a list shaped by capture-group count (see *findall shaping* above); `finditer` returns a tuple of `Match` objects. Both error with `cannot compile pattern` on a bad pattern.

```python
load("regex", "findall", "finditer")
print(findall(r'\d+', 'a1 b22 c333'))
print(findall(r'(\w)(\d)', 'a1 b2'))
print(findall('z', 'abc'))
ms = finditer(r'\d+', 'a1 b22')
print([m.group(0) for m in ms], [m.span(0) for m in ms])
# Output:
# ["1", "22", "333"]
# [("a", "1"), ("b", "2")]
# []
# ["1", "22"] [(1, 2), (4, 6)]
```

### `sub` and `subn`

`repl` is either a string template or a function called with each `Match`. In a string template, `\1`/`\g<name>` reference capture groups, `\n`/`\t`/`\r`/`\\` are the usual escapes, an unrecognized `\x` or unclosed `\g<` is left literal, and a literal `$` is preserved. A function `repl` is called with the `Match` and must return a string. `count` limits the number of replacements: `0` = all, a positive `n` = at most `n`, a negative value = none. `subn` additionally returns the replacement count.

`sub`/`subn` error with `cannot compile pattern` on a bad pattern, `repl must be a string or a function` for a wrong `repl` type, and `repl function must return a string` if a function `repl` returns a non-string.

```python
load("regex", "sub", "subn")
print(sub(r'(\w+)@(\w+)', r'\2.\1', 'ann@host'))
print(sub(r'(?P<x>\d)', r'[\g<x>]', 'a1b2'))
print(sub('a', 'X', 'aaa', 2))
print(sub('a', 'X', 'aaa', -1))
print(subn('a', 'X', 'aaa'))
# Output:
# host.ann
# a[1]b[2]
# XXa
# aaa
# ("XXX", 3)
```

```python
load("regex", "sub")
def up(m):
    return m.group(0).upper()
print(sub(r'[a-z]+', up, 'aa bb'))
# Output: AA BB
```

### `split`

Splits `string` on matches of `pattern`, returning a list. The text of capture groups is included between the pieces (Python semantics). A non-positive `maxsplit` means no limit; otherwise at most `maxsplit` splits are made. Errors with `cannot compile pattern` on a bad pattern.

```python
load("regex", "split")
print(split(r'\s+', 'a b  c'))
print(split(r'(\s+)', 'a b'))
print(split(',', 'a,b,c', 1))
# Output:
# ["a", "b", "c"]
# ["a", " ", "b"]
# ["a", "b,c"]
```

### `escape`

Escapes all regex metacharacters in `pattern` so the result matches the input literally.

```python
load("regex", "escape", "search")
p = escape('a.b*c')
print(search(p, 'a.b*c').group(0))
print(search(p, 'axbyc'))
# Output:
# a.b*c
# None
```

### Flags

```python
load("regex", "search", "findall", "I", "M", "S")
print(search('hello', 'HELLO', I).group(0))
print(findall('^x', 'x\nx\ny', M))
print(search('a.b', 'a\nb', S).group(0))
print(search('a.b', 'a\nb'))
# Output:
# HELLO
# ["x", "x"]
# a
# b
# None
```

### `compile` and `Pattern`

`compile` returns a reusable `Pattern`; its methods drop the leading `pattern` argument. `expand` applies a replacement template against an existing `Match`. `compile` errors with `cannot compile pattern` on an invalid or RE2-unsupported pattern.

```python
load("regex", "compile")
p = compile(r'(?P<n>\d+)')
print(p.pattern, p.groups)
print(p.search('x42').group('n'))
print(p.findall('1 2 3'))
print(p.sub('#', 'a1b2'))
print(p.match('x5'))
# Output:
# (?P<n>\d+) 1
# 42
# ["1", "2", "3"]
# a#b#
# None
```

```python
load("regex", "search")
m = search(r'(\w+) (\w+)', 'hello world')
print(m.expand(r'\2 \1'))
print(m.string, m.re.pattern)
# Output:
# world hello
# hello world (\w+) (\w+)
```

### `try_compile` and `try_search`

The `try_*` variants return a `(value, error)` tuple with a `None` error on success, instead of aborting the script. On a compile failure the value is `None` and the error is a string containing `cannot compile`. `try_search` returns `(None, None)` when the pattern is valid but does not match.

```python
load("regex", "try_compile", "try_search")
p, err = try_compile('a+')
print(err, p != None)
bad, err2 = try_compile('(')
print(bad, 'cannot compile' in err2)
res, err3 = try_search('z', 'abc')
print(res, err3)
# Output:
# None True
# None True
# None None
```

## Notes / boundaries

- **Engine: Go RE2** — linear-time matching, no catastrophic backtracking / ReDoS. RE2-unsupported Python constructs (lookahead/lookbehind, in-pattern backreferences) **fail to compile** with `cannot compile pattern`; they are never silently approximated.
- **Indices are byte offsets**, not rune offsets — `start`/`end`/`span` report positions into the UTF-8 bytes of the subject string.
- **Determinism & purity** — no host effects; the same pattern and input always yield the same result. Suitable for sandboxed/untrusted scripts.
- **Python-parity carve-outs** — `findall` and `split` return **lists** (not tuples), matching CPython; `Match` is unhashable like CPython, while `Pattern` is hashable here so it can serve as a dict key. A negative `count` to `sub`/`subn` replaces nothing.
