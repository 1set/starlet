# file

`file` provides functions to read, write, append, inspect, and copy files on the local file system, plus a small BOM helper. It is inspired by file helpers from common Go toolkits. Capability profile: **FileSystem** — every function except `trim_bom` touches the host file system (reads, writes, stats, or copies real files).

## Functions

### Read

| function | description |
|----------|-------------|
| `read_bytes(name) -> bytes` | Read the whole file and return its contents as bytes. |
| `read_string(name) -> str` | Read the whole file and return its contents as a string. |
| `read_lines(name) -> list` | Read the whole file and return its lines (without line endings) as a list of strings. |
| `read_json(name) -> value` | Read the file and decode its contents as a single JSON document. |
| `read_jsonl(name) -> list` | Read the file as JSON Lines (one JSON document per line; blank lines are skipped) and return a list of values. |
| `head_lines(name, n) -> list` | Return the first `n` lines of the file (or fewer if the file is shorter). |
| `tail_lines(name, n) -> list` | Return the last `n` lines of the file (or fewer if the file is shorter). |
| `count_lines(name) -> int` | Count the number of lines in the file. |

### Write (overwrite)

| function | description |
|----------|-------------|
| `write_bytes(name, data) -> None` | Create or truncate the file and write `data` (string or bytes). |
| `write_string(name, data) -> None` | Create or truncate the file and write `data` (string or bytes) as text. |
| `write_lines(name, data) -> None` | Create or truncate the file and write each item of `data` as a line. |
| `write_json(name, data) -> None` | Create or truncate the file and write `data` as a JSON document. |
| `write_jsonl(name, data) -> None` | Create or truncate the file and write each item of `data` as a JSON line. |

### Append

| function | description |
|----------|-------------|
| `append_bytes(name, data) -> None` | Append `data` (string or bytes) to the file, creating it if absent. |
| `append_string(name, data) -> None` | Append `data` (string or bytes) as text to the file, creating it if absent. |
| `append_lines(name, data) -> None` | Append each item of `data` as a line, creating the file if absent. |
| `append_json(name, data) -> None` | Append `data` as a JSON document, creating the file if absent. |
| `append_jsonl(name, data) -> None` | Append each item of `data` as a JSON line, creating the file if absent. |

### Inspect & copy & utility

| function | description |
|----------|-------------|
| `stat(name, follow=False) -> FileStat` | Return a `FileStat` describing the file or directory. |
| `copyfile(src, dst, overwrite=False) -> str` | Copy a regular file and return the destination path. |
| `trim_bom(rd) -> str \| bytes` | Strip a leading UTF-8 BOM from a string or bytes value (no file I/O). |

## Types

### `FileStat`

Returned by `stat`. A struct (printed type name `file_stat`) carrying file metadata fields plus four hashing methods that read the file on demand.

| member | type | description |
|--------|------|-------------|
| `name` | `str` | Base name of the file or directory. |
| `path` | `str` | Absolute path of the file or directory. |
| `ext` | `str` | File extension including the dot (e.g. `.txt`), or `''` if none. |
| `size` | `int` | Size in bytes. |
| `type` | `str` | One of `file`, `dir`, `symlink`, `fifo`, `socket`, `char`, `block`, `irregular`, `unknown`. |
| `modified` | `time.Time` | Last modification time. |
| `get_md5() -> str` | method | Hex MD5 of the file contents. |
| `get_sha1() -> str` | method | Hex SHA-1 of the file contents. |
| `get_sha256() -> str` | method | Hex SHA-256 of the file contents. |
| `get_sha512() -> str` | method | Hex SHA-512 of the file contents. |

The `get_*` methods open and read the file each time they are called; they error if the path is a directory or unreadable (e.g. `is a directory`, `permission denied`).

## Details & examples

### Reading

`read_bytes(name)`, `read_string(name)`, and `read_lines(name)` read the entire file. `read_lines` strips the trailing newline of each line and handles both `\n` and `\r\n` endings; an empty file yields `[]`. All three error if the path does not exist (e.g. `open no-such-file:`).

```python
load('file', 'read_bytes', 'read_string', 'read_lines')
print(read_bytes('testdata/aloha.txt') == b'ALOHA\n')  # bytes, trailing newline kept
print(read_string('testdata/aloha.txt') == 'ALOHA\n')  # same content as a string
print(read_lines('testdata/line_win.txt'))             # newline stripped, \r\n handled
# Output:
# True
# True
# ["Line 1", "Line 2", "Line 3"]
```

`read_json(name)` decodes the whole file as one JSON document and returns the matching Starlark value (`dict`, `list`, `int`, `bool`, `None`, etc.). `read_jsonl(name)` decodes one JSON document per line, skipping blank lines, and returns a list. Both error if the file is missing (`open no-such-file:`) or the content is not valid JSON (`read_json` → `json.decode: at offset ...`; `read_jsonl` → `line N: json.decode: at offset ...`).

```python
load('file', 'read_json')
data = read_json('testdata/json1.json')
print(data['num'], data['bool'], data['arr'])
# Output: 42 True [1, 2, 3]
```

`head_lines(name, n)` and `tail_lines(name, n)` return at most `n` lines from the start or end of the file. `n` must be a positive integer, otherwise they error (`expected positive integer, got -7`); a missing file errors as `open no-such-file:`. `count_lines(name)` returns the line count (0 for an empty file).

```python
load('file', 'head_lines', 'tail_lines', 'count_lines')
print(head_lines('testdata/line_win.txt', 2))
print(tail_lines('testdata/line_win.txt', 2))
print(count_lines('testdata/line_mac.txt'))
# Output:
# ["Line 1", "Line 2"]
# ["Line 2", "Line 3"]
# 3
```

### Writing and appending

The `write_*` functions create the file if absent and truncate it if present; the `append_*` functions create the file if absent and append to it otherwise. All of them return `None` and error if the target cannot be opened (e.g. writing to a directory path: `open testdata/:`). Both arguments are required; calling without `name` or `data` errors with `missing argument for name` / `missing argument for data`.

- `write_bytes` / `append_bytes` and `write_string` / `append_string` accept a string or bytes value; any other type errors with `expected string or bytes, got <type>`.
- `write_lines` / `append_lines` accept a `list`, `tuple`, or `set` (each item is rendered with one trailing newline); a bare string or bytes is treated as a single line. Other types error with `expected list/tuple/set, got <type>`. Non-string items are stringified (e.g. `123`, `[True, False]`).
- `write_json` / `append_json` write a string or bytes value verbatim; any other value is JSON-encoded. Values that cannot be encoded error (e.g. a lambda: `json.encode: cannot encode function as JSON`).
- `write_jsonl` / `append_jsonl` write a string or bytes value as one line; for a `list`/`tuple`/`set` each item is JSON-encoded onto its own line; any other value is encoded onto a single line.

```python
load('file', 'write_lines', 'append_lines')
fp = 'out.txt'
write_lines(fp, ['Hello', 'World'])
append_lines(fp, ['Great', 'Job'])
print(read_string(fp))
# Output:
# Hello
# World
# Great
# Job
#
```

```python
load('file', 'write_jsonl')
write_jsonl('out.jsonl', [{'a': 520}, {'b': True}])
# out.jsonl now contains:
# {"a":520}
# {"b":true}
```

### `stat(name, follow=False) -> FileStat`

Returns a `FileStat` for the path. By default symbolic links are reported as links (`lstat`); pass `follow=True` to resolve the link and stat its target. Errors if the path does not exist (`file.stat: lstat <name>: ...`).

```python
load('file', 'stat')
s = stat('testdata/aloha.txt')
print(s.name, s.size, s.type, s.ext)
print(s.get_md5())
# Output:
# aloha.txt 6 file .txt
# 6a12867bd5e0810f2dae51da4a51f001
```

### `copyfile(src, dst, overwrite=False) -> str`

Copies the regular file at `src` to `dst` and returns the destination path. If `dst` is an existing directory, the file is copied into it under its original base name. Symbolic links are followed; the file mode and access/modification times are preserved on a best-effort basis (errors setting them are ignored). It errors when: `src` or `dst` is empty (`source path is empty` / `destination path is empty`), `src` is not a regular file (`source file is not a regular file`), `src` and `dst` resolve to the same file (`source and destination are the same file`), or `dst` exists and `overwrite` is `False` (`file already exists`).

```python
load('file', 'copyfile')
dst = copyfile('testdata/aloha.txt', 'copy.txt')
print(dst)
# Output: copy.txt
```

### `trim_bom(rd) -> str | bytes`

Removes a leading UTF-8 byte order mark (`\xef\xbb\xbf`) from a string or bytes value, returning the same type it received. Input without a BOM is returned unchanged. It is the only pure function here (no file I/O). It takes exactly one argument (`takes exactly one argument (0 given)`) and rejects other types (`expected string or bytes, got int`).

```python
load('file', 'trim_bom')
print(trim_bom(b'\xef\xbb\xbfhello'))  # bytes in, bytes out; print shows the content
print(trim_bom('hello'))               # no BOM, returned unchanged
# Output:
# hello
# hello
```

## Notes / boundaries

- **Side effects.** All functions except `trim_bom` perform real file system I/O; this is a FileSystem-capability module and must be gated as such by the host.
- **Line endings.** Readers strip line endings and accept both `\n` and `\r\n`; the line writers always emit `\n`.
- **JSON engine.** JSON encoding/decoding goes through `dataconv` (the same engine as the `json` module); only JSON-encodable Starlark values can be written, and decode errors surface the offset.
- **Atomicity.** Writes are not atomic — a truncating write that fails partway can leave a partially written file. Appends use `O_APPEND`.
- All exported names are `snake_case`; the `FileStat` hashing methods (`get_md5`, `get_sha1`, `get_sha256`, `get_sha512`) follow the same convention.
