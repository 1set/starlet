# path

`path` provides functions to manipulate directories and file paths. It follows one rule:

- **Lexical functions** — `join`, `basename`, `dirname`, `normpath`, `split`, `splitext`, `isabs`, `relpath` — use **Python (`posixpath`) semantics**: pure string work on `/`-separated paths, no implicit cleaning, identical results on every OS.
- **Filesystem functions** — `abs`, `exists`, `is_file`, `is_dir`, `is_link`, `listdir`, `getcwd`, `chdir`, `mkdir` — operate on the real, OS-native filesystem.

## Migrating from v0.1.x

`join` previously used Go's `filepath.Join`, which cleaned the result. With Python semantics:

| call | v0.1.x | now |
|---|---|---|
| `join("a", "/b")` | `a/b` | `/b` (an absolute component resets the result) |
| `join("a", "")` | `a` | `a/` (an empty component adds a trailing separator) |
| `join("a", "../b")` | `b` | `a/../b` (no cleaning — apply `normpath` to collapse) |
| `join("a//b", "c")` | `a/b/c` | `a//b/c` |

Use `normpath(join(...))` where the old cleaned result is wanted.

## Functions

### `abs(path) string`

Returns an absolute representation of path. If the path is not absolute it will be joined with the current working directory to turn it into an absolute path. The absolute path name for a given file is not guaranteed to be unique.

#### Parameters

| name   | type     | description                                        |
|--------|----------|----------------------------------------------------|
| `path` | `string` | The file path to be converted to its absolute form |

#### Examples

**basic**

Convert a relative path to an absolute path.

```python
load("path", "abs")
p = abs('.')
print(p)
# Output: '/current/absolute/path'
```

### `join(path, *paths) string`

Joins one or more path elements with Python (`posixpath`) semantics: components are joined with `/`, an absolute component resets the result, an empty component contributes a trailing separator, and no cleaning is applied (see the migration notes above).

#### Parameters

| name       | type     | description                    |
|------------|----------|--------------------------------|
| `paths...` | `string` | The path elements to be joined |

#### Examples

**basic**

Join multiple path parts.

```python
load("path", "join")
p = join('a', 'b', 'c')
print(p)
# Output: 'a/b/c'
```

### `exists(path) bool`

Returns true if the path exists.

#### Parameters

| name   | type     | description            |
|--------|----------|------------------------|
| `path` | `string` | The path to be checked |

#### Examples

**basic**

Check if a path exists.

```python
load("path", "exists")
p = exists('path_test.go')
print(p)
# Output: True
```

### `is_file(path) bool`

Returns true if the path exists and is a file.

#### Parameters

| name   | type     | description            |
|--------|----------|------------------------|
| `path` | `string` | The path to be checked |

#### Examples

**basic**

Check if a path is a file.

```python
load("path", "is_file")
p = is_file('path_test.go')
print(p)
# Output: True
```

### `is_dir(path) bool`

Returns true if the path exists and is a directory.

#### Parameters

| name   | type     | description            |
|--------|----------|------------------------|
| `path` | `string` | The path to be checked |

#### Examples

**basic**

Check if a path is a directory.

```python
load("path", "is_dir")
p = is_dir('.')
print(p)
# Output: True
```

### `is_link(path) bool`

Returns true if the path exists and is a symbolic link.

#### Parameters

| name   | type     | description            |
|--------|----------|------------------------|
| `path` | `string` | The path to be checked |

#### Examples

**basic**

Check if a path is a symbolic link.

```python
load("path", "is_link")
p = is_link('link_to_path_test.go')
print(p)
# Output: False
```

### `listdir(path, recursive=False, filter=None) []string`

Returns a list of directory contents. Optionally applies a filter function to each path to decide inclusion in the final list.

#### Parameters

| name        | type       | description                                                                                                                                                                                            |
|-------------|------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `path`      | `string`   | The directory path to be listed                                                                                                                                                                        |
| `recursive` | `bool`     | If true, list contents recursively                                                                                                                                                                     |
| `filter`    | `callable` | A callable object (e.g., lambda or function) that takes a single argument (a path) and returns a boolean value. Paths for which the filter function returns `False` are excluded from the result list. |

#### Examples

**basic**

List directory contents.

```python
load("path", "listdir")
p = listdir('.')
print(p)
# Output: ['file1', 'file2', ...]
```

**recursive**

List directory contents recursively.

```python
load("path", "listdir")
p = listdir('.', True)
print(p)
# Output: ['file1', 'file2', 'subdir/file3', ...]
```

**filtered**

List directory contents with a filter function.

```python
load("path", "listdir")
is_not_go_file = lambda p: not p.endswith('.go')
p = listdir('.', filter=is_not_go_file)
print(p)
# Output: ['file1.py', 'file2.txt', ...]
```

**filtered_recursive**

List directory contents recursively with a filter function.

```python
load("path", "listdir")
is_not_go_file = lambda p: not p.endswith('.go')
p = listdir('.', True, filter=is_not_go_file)
print(p)
# Output: ['file1.py', 'file2.txt', 'subdir/file3']
```

### `getcwd() string`

Returns the current working directory.

#### Examples

**basic**

Get the current working directory.

```python
load("path", "getcwd")
p = getcwd()
print(p)
# Output: '/current/directory'
```

### `chdir(path)`

Changes the current working directory **of the whole process**.

> ⚠️ The effect is global: it applies to every machine and goroutine in the host, persists after the script ends, and concurrent machines calling it race with each other. Never expose this module to untrusted scripts; host runtimes typically exclude `path` from restricted module sets for this reason.

#### Parameters

| name   | type     | description                           |
|--------|----------|---------------------------------------|
| `path` | `string` | The path to the new current directory |

#### Examples

**basic**

Change the current working directory.

```python
load("path", "chdir")
chdir('/new/directory')
# Current directory is now '/new/directory'
```

### `mkdir(path, mode=0o755)`

Creates a directory with the given name. If the directory already exists, no error is thrown. It's capable of creating nested directories.

#### Parameters

| name   | type     | description                                                                                                           |
|--------|----------|-----------------------------------------------------------------------------------------------------------------------|
| `path` | `string` | The directory path to be created                                                                                      |
| `mode` | `int`    | The file mode (permissions) to use for the newly-created directory, represented as an octal number. Defaults to 0755. |

#### Examples

**default**

Create a new directory.

```python
load("path", "mkdir")
mkdir('new_directory')
# New directory named 'new_directory' is created with default permissions
```

**permission**

Create a new directory with specific permissions.

```python
load("path", "mkdir")
mkdir('secure_directory', 0o700)
# New directory named 'secure_directory' is created with permissions set to 0700
```

### `basename(path) string`

Returns the final component of a path — everything after the last `/`; a path ending in `/` has an empty basename (`basename("a/b/")` is `""`, unlike Go's `filepath.Base`).

### `dirname(path) string`

Returns the directory part of a path — everything before the last `/`, with trailing slashes stripped unless the result is all slashes.

### `normpath(path) string`

Collapses redundant separators and up-level references lexically: `a//b`, `a/./b` and `a/c/../b` all become `a/b`. Exactly two leading slashes are preserved (POSIX gives them special meaning); an empty path normalizes to `.`.

### `split(path) (string, string)`

Splits a path into a `(head, tail)` pair: `split("/a")` is `("/", "a")`, `split("a/b/")` is `("a/b", "")`.

### `splitext(path) (string, string)`

Splits a path into a `(root, extension)` pair: `splitext("a/b.tar.gz")` is `("a/b.tar", ".gz")`; leading dots do not count, so `splitext(".bashrc")` is `(".bashrc", "")`.

### `isabs(path) bool`

Reports whether the path is absolute (starts with `/`).

### `relpath(path, start=".") string`

Returns a relative path to `path` from `start`, computed lexically on the normalized inputs. Mixing an absolute path with a relative one is an error here — resolving the mix would silently depend on the process working directory; call `abs()` first when that is intended. `try_relpath` returns a `(value, error)` pair instead of aborting.

### `expanduser(path) string`

Replaces a leading `~` with the current user's home directory. Only the bare `~`/`~/...` form expands; `~user/...` is returned unchanged. Note that this reads the host environment.

### `try_abs(path) (string, error)` / `try_relpath(path, start=".") (string, error)`

The `(value, error)` pair forms of `abs` and `relpath`, never aborting the script — the same shape as the `json`/`csv`/`http` modules' `try_*` functions.
