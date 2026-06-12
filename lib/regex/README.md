# regex

`regex` provides regular expression functions for Starlark, a subset of [Python's **re** module](https://docs.python.org/3/library/re.html) backed by Go's [RE2 engine](https://golang.org/s/re2syntax).

It is the successor to the legacy `re` module (which is frozen): `regex` adds **Match objects**, named-group extraction, the `IGNORECASE`/`MULTILINE`/`DOTALL` flags, `\1`/`\g<name>` replacement and function replacements, and the full `compile`/`fullmatch`/`finditer`/`subn`/`escape` surface.

**Python re semantics subset (RE2 engine, no lookaround / no backreferences).** RE2 matches in linear time — no catastrophic backtracking / ReDoS — which suits running untrusted or LLM-generated scripts in a sandbox. Where RE2 genuinely differs from Python — lookahead/lookbehind (`(?=...)`, `(?<=...)`) and in-pattern backreferences (`\1`) — the pattern fails to **compile** with a clear error rather than silently misbehaving.

## Functions

| function | description |
|---|---|
| `compile(pattern, flags=0)` | compile a pattern into a `Pattern` object |
| `search(pattern, string, flags=0)` | first match anywhere → `Match` or `None` |
| `match(pattern, string, flags=0)` | match anchored at the **start** → `Match` or `None` |
| `fullmatch(pattern, string, flags=0)` | match the **whole** string → `Match` or `None` |
| `findall(pattern, string, flags=0)` | all matches as a list (Python group shaping, see below) |
| `finditer(pattern, string, flags=0)` | a tuple of `Match` objects |
| `sub(pattern, repl, string, count=0, flags=0)` | replace matches → string |
| `subn(pattern, repl, string, count=0, flags=0)` | replace → `(string, count)` |
| `split(pattern, string, maxsplit=0, flags=0)` | split into a list, **including capture-group text** (Python semantics) |
| `escape(pattern)` | escape regex metacharacters |

`try_compile` / `try_search` are also provided, returning a `(value, error)` pair instead of aborting the script — the same shape as the `json`/`csv`/`http` modules.

**findall shaping** (matches Python): the result is a **list**; no capture group → the full match text; one group → that group's text; two or more groups → a **tuple** of the group texts per match. (`split` likewise returns a list — the same as Python's `re` and starlet's legacy `re` module.)

**sub replacement**: `repl` is either a string template — `\1`/`\g<name>` reference capture groups, `\n`/`\t`/`\r` are the usual escapes, a literal `$` is preserved — or a function called with each `Match`, returning the replacement string. `count` limits the number of replacements (`0` = all, negative = none).

## Flags

`I`/`IGNORECASE`, `M`/`MULTILINE`, `S`/`DOTALL` — integer constants (Python's `re` values) that may be combined with `|` and translate to RE2 inline flags.

```python
load("regex", "search", "I")
print(search("hello", "HELLO WORLD", I).group(0))
# Output: "HELLO"
```

## Types

### `Pattern`

A compiled regular expression. Methods mirror the module functions without the `pattern` argument: `search`, `match`, `fullmatch`, `findall`, `finditer`, `sub`, `subn`, `split`. Attributes: `pattern` (the source), `flags`, `groups` (number of capture groups).

### `Match`

The result of a successful `search`/`match`/`fullmatch`.

| member | description |
|---|---|
| `group(n=0, ...)` | group text by index or name; several args → a tuple; group 0 is the whole match |
| `groups(default=None)` | a tuple of all capture groups (`default` for non-participating groups) |
| `groupdict(default=None)` | a dict of named groups to their text |
| `start(n=0)` / `end(n=0)` | start/end byte index of a group |
| `span(n=0)` | `(start, end)` of a group |
| `expand(template)` | apply a `\1`/`\g<name>` template against the match |
| `string` | the subject string that was matched |
| `re` | the `Pattern` that produced the match |

A non-participating optional group reports `None`. `Match` objects are unhashable (they cannot be used as dict keys).

```python
load("regex", "search")
m = search(r'(?P<user>\w+)@(?P<host>\w+)', 'ann@example.com')
print(m.group('user'), m.group('host'))   # ann example
print(m.groupdict())                       # {"user": "ann", "host": "example"}
print(m.span(0))                           # (0, 11)
```
