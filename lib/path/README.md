# path

`path` manipulates directories and file paths for Starlark. The **lexical** functions follow Python's `posixpath` semantics — pure string work on `/`-separated paths, no implicit cleaning, identical on every OS — while the **filesystem** functions operate on the real, OS-native filesystem.

Capability profile: **FileSystem + Process**. The filesystem functions touch the host disk; `chdir` changes the working directory of the **whole process** (every machine and goroutine in the host, persisting after the script ends), so this is a Process effect too. Never expose this module to untrusted scripts; host runtimes typically exclude `path` from restricted module sets.

## Functions

### Lexical (pure string, `posixpath` semantics)

| function | description |
|---|---|
| `join(*paths) -> string` | Join path elements with `/`; an absolute component resets the result, an empty component adds a trailing separator, no cleaning. |
| `basename(path) -> string` | Final component (everything after the last `/`); empty when `path` ends in `/`. |
| `dirname(path) -> string` | Directory part (everything before the last `/`), trailing slashes stripped unless all-slashes. |
| `normpath(path) -> string` | Collapse redundant separators and `..`/`.` lexically. |
| `split(path) -> (head, tail)` | Split into `(dirname, basename)`. |
| `splitext(path) -> (root, ext)` | Split off the extension at the last dot of the final component. |
| `isabs(path) -> bool` | Whether `path` is absolute (starts with `/`). |
| `relpath(path, start=".") -> string`, `try_relpath(path, start=".") -> (string, error)` | Relative path from `start` to `path`, computed lexically. `try_relpath` returns a `(value, error)` pair instead of aborting. |

### Filesystem (real OS-native disk)

| function | description |
|---|---|
| `abs(path) -> string`, `try_abs(path) -> (string, error)` | Absolute representation of `path`. `try_abs` returns a `(value, error)` pair instead of aborting. |
| `expanduser(path) -> string` | Replace a leading `~` with the current user's home directory (reads the host environment). |
| `exists(path) -> bool` | Whether `path` exists (symlinks are followed). |
| `is_file(path) -> bool` | Whether `path` exists and is a regular file (symlinks followed). |
| `is_dir(path) -> bool` | Whether `path` exists and is a directory (symlinks followed). |
| `is_link(path) -> bool` | Whether `path` exists and is a symbolic link (not followed). |
| `listdir(path, recursive=False, filter=None) -> list[string]` | List directory contents, optionally recursively and filtered. |
| `getcwd() -> string` | Current working directory of the process. |
| `chdir(path) -> None` | Change the **process-wide** working directory (global side effect). |
| `mkdir(path, mode=0o755) -> None` | Create a directory (and parents); existing directories are not an error. |

## Migrating from v0.1.x

`join` previously used Go's `filepath.Join`, which cleaned the result. With Python semantics:

| call | v0.1.x | now |
|---|---|---|
| `join("a", "/b")` | `a/b` | `/b` (an absolute component resets the result) |
| `join("a", "")` | `a` | `a/` (an empty component adds a trailing separator) |
| `join("a", "../b")` | `b` | `a/../b` (no cleaning — apply `normpath` to collapse) |
| `join("a//b", "c")` | `a/b/c` | `a//b/c` |

Use `normpath(join(...))` where the old cleaned result is wanted.

## Lexical functions

### `join(*paths) -> string`

Joins one or more string elements. Components are joined with `/`; an absolute component (one starting with `/`) resets the result; an empty component contributes a trailing separator; no lexical cleaning is applied. Requires at least one argument; every argument must be a string.

Errors on: zero arguments (`got 0 arguments, want at least 1`); a non-string argument (`for parameter path: got int, want string`).

```python
load("path", "join")
print(join("a", "b", "c"))     # 'a/b/c'
print(join("a", "/b", "c"))    # '/b/c'   (absolute resets)
print(join("a", "b", ""))      # 'a/b/'   (empty adds trailing /)
print(join("a/b", "../../xyz")) # 'a/b/../../xyz' (no cleaning)
# Output: a/b/c
# /b/c
# a/b/
# a/b/../../xyz
```

### `basename(path) -> string`

Everything after the last `/`. A path ending in `/` has an empty basename (unlike Go's `filepath.Base`). Errors on a non-string argument.

```python
load("path", "basename")
print(basename("a/b/c.txt")) # 'c.txt'
print(basename("a/b/"))      # ''
print(basename("/"))         # ''
print(basename("plain"))     # 'plain'
# Output: c.txt
#
#
# plain
```

### `dirname(path) -> string`

Everything before the last `/`, with trailing slashes stripped unless the result is all slashes (so `dirname("//a")` keeps `//`). Errors on a non-string argument.

```python
load("path", "dirname")
print(dirname("a/b/c.txt")) # 'a/b'
print(dirname("a/b/"))      # 'a/b'
print(dirname("plain"))     # ''
print(dirname("/a"))        # '/'
print(dirname("//a"))       # '//'
# Output: a/b
# a/b
#
# /
# //
```

### `normpath(path) -> string`

Collapses redundant separators and up-level references lexically: `a//b`, `a/./b` and `a/c/../b` all become `a/b`. Exactly two leading slashes are preserved (POSIX gives them implementation-defined meaning); three or more collapse to one; an empty path normalizes to `.`. Errors on a non-string argument.

```python
load("path", "normpath")
print(normpath("a/c/../b")) # 'a/b'
print(normpath("a/../../b")) # '../b'
print(normpath("//a"))      # '//a'
print(normpath("///a"))     # '/a'
print(normpath(""))         # '.'
# Output: a/b
# ../b
# //a
# /a
# .
```

### `split(path) -> (head, tail)`

Returns the `(dirname, basename)` pair. `split("/a")` is `("/", "a")`; `split("a/b/")` is `("a/b", "")`. Errors on a non-string argument.

```python
load("path", "split")
print(split("a/b/c.txt")) # ('a/b', 'c.txt')
print(split("/a"))        # ('/', 'a')
print(split("a/b/"))      # ('a/b', '')
print(split("plain"))     # ('', 'plain')
# Output: ("a/b", "c.txt")
# ("/", "a")
# ("a/b", "")
# ("", "plain")
```

### `splitext(path) -> (root, ext)`

Splits off the extension — the suffix beginning at the last dot of the final component. Leading dots do not count, so `splitext(".bashrc")` is `(".bashrc", "")`. A dot only in a directory component is ignored. Errors on a non-string argument.

```python
load("path", "splitext")
print(splitext("a/b.tar.gz")) # ('a/b.tar', '.gz')
print(splitext(".bashrc"))    # ('.bashrc', '')
print(splitext("a/.bashrc"))  # ('a/.bashrc', '')
print(splitext("a.b/c"))      # ('a.b/c', '')
# Output: ("a/b.tar", ".gz")
# (".bashrc", "")
# ("a/.bashrc", "")
# ("a.b/c", "")
```

### `isabs(path) -> bool`

Reports whether `path` is absolute in the POSIX sense (starts with `/`). The empty string is not absolute. Errors on a non-string argument.

```python
load("path", "isabs")
print(isabs("/a/b")) # True
print(isabs("a/b"))  # False
print(isabs(""))     # False
# Output: True
# False
# False
```

### `relpath(path, start=".") -> string` / `try_relpath(path, start=".") -> (string, error)`

Returns a relative path from `start` to `path`, computed lexically on the normalized inputs. `start` defaults to `"."` (an empty `start` is also treated as `"."`).

Errors on: an empty `path` (`no path specified`); mixing an absolute path with a relative `start` (`cannot mix an absolute path with a relative start`) — resolving the mix would silently depend on the process working directory, so call `abs()` first when that is intended; a non-string argument. `try_relpath` returns these as the error half of a `(value, error)` pair rather than aborting.

```python
load("path", "relpath")
print(relpath("/a/b/c", "/a"))   # 'b/c'
print(relpath("/a/b", "/a/b"))   # '.'
print(relpath("/a/b", "/a/c/d")) # '../../b'
print(relpath("a/b", "a"))       # 'b'
print(relpath("a/b"))            # 'a/b'  (start defaults to '.')
# Output: b/c
# .
# ../../b
# b
# a/b
```

```python
load("path", "try_relpath")
v, err = try_relpath("/a/b/c", "/a")
print(v, err)                 # 'b/c' None
v2, err2 = try_relpath("/a", "c")
print(v2, "cannot mix" in err2) # None True
# Output: b/c None
# None True
```

## Filesystem functions

### `abs(path) -> string` / `try_abs(path) -> (string, error)`

Returns an absolute representation of `path`. A relative path is joined with the current working directory; the result is not guaranteed to be unique. Errors on a missing argument (`missing argument for path`) or a non-string argument. `try_abs` returns a `(value, error)` pair instead of aborting.

```python
load("path", "abs")
p = abs("path_test.go")
print(p.endswith("lib/path/path_test.go"))
# Output: True
```

```python
load("path", "try_abs")
v, err = try_abs(".")
print(err, len(v) > 0)
# Output: None True
```

### `expanduser(path) -> string`

Replaces a leading `~` with the current user's home directory. Only the bare `~` or `~/...` form expands; `~user/...` is returned unchanged (matching Python when the user lookup is unavailable), and so is the path when the home directory cannot be determined. Reads the host environment. Errors on a non-string argument.

```python
load("path", "expanduser")
print(expanduser("~") != "~")          # True (home resolved)
print(expanduser("~/x").endswith("/x")) # True
print(expanduser("~user/x"))           # '~user/x'
print(expanduser("plain"))             # 'plain'
# Output: True
# True
# ~user/x
# plain
```

### `exists(path) -> bool`

True if `path` exists; a symbolic link is followed. The empty string and non-existent paths return `False`. Errors only on a missing or non-string argument.

```python
load("path", "exists")
print(exists("path_test.go")) # True
print(exists("."))            # True
print(exists("nope"))         # False
# Output: True
# True
# False
```

### `is_file(path) -> bool`

True if `path` exists and is a regular file (symlinks followed). Directories and non-existent paths return `False`.

```python
load("path", "is_file")
print(is_file("path_test.go")) # True
print(is_file("."))            # False
print(is_file("nope"))         # False
# Output: True
# False
# False
```

### `is_dir(path) -> bool`

True if `path` exists and is a directory (symlinks followed).

```python
load("path", "is_dir")
print(is_dir("."))             # True
print(is_dir("path_test.go"))  # False
print(is_dir("nope"))          # False
# Output: True
# False
# False
```

### `is_link(path) -> bool`

True if `path` exists and is a symbolic link. The link itself is inspected (not followed), so plain files, directories, and non-existent paths return `False`.

```python
load("path", "is_link")
print(is_link("path_test.go")) # False
print(is_link("."))            # False
# Output: False
# False
```

### `listdir(path, recursive=False, filter=None) -> list[string]`

Returns a list of directory contents. Listing a non-directory (e.g. a file) returns an empty list. With `recursive=True` the walk descends into subdirectories. `filter` is a callable taking one path argument and returning a bool; paths for which it returns `False` are excluded.

Errors on: a missing or non-string `path`; a non-existent path (`lstat ...`); an unreadable directory (`open ...`); a `filter` that is neither callable nor `None` (`expected <nil> or None, got int`); a filter returning a non-bool (`got int, want bool`); or a filter that itself fails (the inner error propagates).

```python
load("path", "listdir")
p = listdir(".")
print("path_test.go" in p)
# Output: True
```

```python
load("path", "listdir")
p = listdir(".", filter=lambda x: not x.endswith(".go"))
print("path_test.go" not in p)
# Output: True
```

### `getcwd() -> string`

Returns the current working directory of the process. Takes no arguments; passing any is an error (`got 1 arguments, want 0`).

```python
load("path", "getcwd")
p = getcwd()
print(p.endswith("path"))
# Output: True
```

### `chdir(path) -> None`

Changes the current working directory **of the whole process**.

> WARNING: the effect is global — it applies to every machine and goroutine in the host, persists after the script ends, and concurrent machines calling it race with each other. Never expose this module to untrusted scripts.

Errors on a missing or non-string argument, or when the target cannot be entered (a non-existent path or a file both error with `chdir ...`).

```python
load("path", "chdir", "abs")
a = abs(".")
chdir(".")
b = abs(".")
print(a == b)
# Output: True
```

### `mkdir(path, mode=0o755) -> None`

Creates a directory at `path`, creating parent directories as needed (like `mkdir -p`). An already-existing directory is not an error. `mode` is the octal permission bits for newly-created directories (default `0o755`); `path` may be a string or bytes.

Errors on a missing argument, or when a path component is an existing non-directory (`not a directory`).

```python
load("path", "mkdir")
mkdir("new_directory")          # default 0o755
mkdir("secure_directory", 0o700) # explicit mode
# Output:
```

## Notes / boundaries

- **Lexical vs filesystem.** `join`, `basename`, `dirname`, `normpath`, `split`, `splitext`, `isabs`, and `relpath` are pure string operations on `/`-separated paths and never touch disk — they match CPython's `posixpath` (not Go's `path/filepath`, which cleans eagerly and uses the OS separator). The rest read or mutate the real filesystem.
- **`expanduser` and the filesystem functions are OS-native**, so paths use the host separator and `abs`/`getcwd` return host-absolute paths; examples above that assert OS-specific shapes are skipped on Windows in the test suite.
- **`chdir` is a process-global, persistent side effect** shared across all machines — the Process capability. The test suite saves and restores the working directory around each case for this reason.
- **`try_abs` / `try_relpath`** never abort the script: they return a `(value, error)` tuple with `None` error on success, the same shape as the `json`/`csv`/`http` modules' `try_*` functions.
- All names are snake_case; the only non-alphabetic members are the `try_` prefixes on `try_abs` and `try_relpath`.
