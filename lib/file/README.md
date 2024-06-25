# file

`file` provides functions to interact with the file system. The library is inspired by file helpers from Amoy.

## Functions

### `trim_bom(rd) string`

Removes the Byte Order Mark (BOM) from a byte literal string or bytes.

#### Parameters

| name | type    | description |
|------|---------|-------------|
| `rd` | `string | byes`       |

#### Examples

**basic**

Removes the Byte Order Mark (BOM) from a string.

```python
load("file", "trim_bom")
s = b'\xef\xbb\xbfhello'
print(trim_bom(s))
# Output: hello
```

### `count_lines(name) int`

Counts the total lines in a file located at the provided path.

#### Parameters

| name   | type     | description                                        |
|--------|----------|----------------------------------------------------|
| `name` | `string` | The path of the file whose lines are to be counted |

#### Examples

**basic**

Count the lines of a file.

```python
load("file", "count_lines")
name = 'path/to/file.txt'
print(count_lines(name))
# Output: 10
```

### `head_lines(name, n) []string`

Returns the first 'n' lines of a file.

#### Parameters

| name   | type     | description                                     |
|--------|----------|-------------------------------------------------|
| `name` | `string` | The path of the file                            |
| `n`    | `int`    | The number of lines from the top to be returned |

#### Examples

**basic**

Get the top 10 lines of a file.

```python
load('file', 'head_lines')
print(head_lines('path/to/file.txt', 10))
# Output: ['line1', 'line2', ... 'line10']
```

### `tail_lines(name, n) []string`

Returns the last 'n' lines of a file.

#### Parameters

| name   | type     | description                                        |
|--------|----------|----------------------------------------------------|
| `name` | `string` | The path of the file                               |
| `n`    | `int`    | The number of lines from the bottom to be returned |

#### Examples

**basic**

Get the bottom 10 lines of a file.

```python
load('file', 'tail_lines')
print(tail_lines('path/to/file.txt', 10))
# Output: ['line91', 'line92', ... 'line100']
```

### `read_bytes(name)`

Reads a file and returns its contents as bytes.

#### Parameters

| name   | type     | description                     |
|--------|----------|---------------------------------|
| `name` | `string` | The path of the file to be read |

#### Examples

**basic**

Read a file in bytes.

```python
load('file', 'read_bytes')
print(read_bytes('path/to/file.txt'))
# Output: b'file_content'
```

### `read_string(name)`

Reads a file and returns its contents as string.

#### Parameters

| name   | type     | description                     |
|--------|----------|---------------------------------|
| `name` | `string` | The path of the file to be read |

#### Examples

**basic**

Read a file in string.

```python
load('file', 'read_string')
print(read_string('path/to/file.txt'))
# Output: 'file_content'
```

### `read_lines(name)`

Reads a file and returns its contents as a list of lines.

#### Parameters

| name   | type     | description                     |
|--------|----------|---------------------------------|
| `name` | `string` | The path of the file to be read |

#### Examples

**basic**

Get lines of a file in a list.

```python
load('file', 'read_lines')
print(read_lines('path/to/file.txt'))
# Output: ['line1', 'line2', 'line3', ....]
```

### `read_json(name) dict`

Reads a file and decodes its contents as JSON, returning the corresponding Starlark object (dict or any types).

#### Parameters

| name   | type     | description                          |
|--------|----------|--------------------------------------|
| `name` | `string` | The path of the JSON file to be read |

#### Examples

**basic**

Read a JSON file.

```python
load('file', 'read_json')
data = read_json('path/to/file.json')
print(data)
# Output: {'key': 'value', 'array': [1, 2, 3]}
```

### `read_jsonl(name) list`

Reads a file with each line containing a JSON object and returns a list of Starlark objects.

#### Parameters

| name   | type     | description                           |
|--------|----------|---------------------------------------|
| `name` | `string` | The path of the JSONL file to be read |

#### Examples

**basic**

Read a JSONL file.

```python
load('file', 'read_jsonl')
data = read_jsonl('path/to/file.jsonl')
print(data)
# Output: [{'key1': 'value1'}, {'key2': 'value2'}]
```

### `write_bytes(name, data)`

Writes/overwrites bytes or a byte literal string to a file. If the file isn't present, a new file would be created.

#### Parameters

| name   | type     | description                                                |
|--------|----------|------------------------------------------------------------|
| `name` | `string` | The path of the file to be written to                      |
| `data` | `string` | The byte literal string or bytes to be written to the file |

#### Examples

**basic**

Write a byte string to a file.

```python
load('file', 'write_bytes')
name = 'new_file.txt'
data = b'Hello, This is a new file.'
write_bytes(name, data)
```

### `write_string(name, data)`

Writes/overwrites a string to a file. If the file isn't present, a new file would be created.

#### Parameters

| name   | type     | description                           |
|--------|----------|---------------------------------------|
| `name` | `string` | The path of the file to be written to |
| `data` | `string` | The string to be written to the file  |

#### Examples

**basic**

Write a string to a file.

```python
load('file', 'write_string')
write_string('new_file.txt', 'Hello, This is a new file.')
```

### `write_lines(name, data)`

Writes/overwrites a list, tuple or set of lines to a file. If the file isn't present, a new file would be created.

#### Parameters

| name   | type     | description                           |
|--------|----------|---------------------------------------|
| `name` | `string` | The path of the file to be written to |
| `data` | `list    | set                                   |

#### Examples

**List**

Write a list of lines to a file.

```python
load('file', 'write_lines')
lines = ['This is line1', 'This is line2', 'This is line3']
write_lines('new_file.txt', lines)
```

### `write_json(name, data)`

Writes the given Starlark object as JSON to a file. If the file exists, it will be overwritten.

#### Parameters

| name   | type     | description                           |
|--------|----------|---------------------------------------|
| `name` | `string` | The path of the file to be written to |
| `data` | `dict    | list                                  |

#### Examples

**basic**

Write a dictionary as JSON to a file.

```python
load('file', 'write_json')
data = {"key": "value", "array": [1, 2, 3]}
write_json('new_file.json', data)
```

### `write_jsonl(name, data)`

Writes the given data as JSON lines to a file. If the file exists, it will be overwritten.

#### Parameters

| name   | type     | description                           |
|--------|----------|---------------------------------------|
| `name` | `string` | The path of the file to be written to |
| `data` | `list    | set                                   |

#### Examples

**basic**

Write a list of JSON objects to a file as JSONL.

```python
load('file', 'write_jsonl')
data = [{"key1": "value1"}, {"key2": "value2"}]
write_jsonl('new_file.jsonl', data)
```

### `append_bytes(name, data)`

Appends bytes or a byte literal string to a file. If the file isn't present, a new file would be created.

#### Parameters

| name   | type     | description                                                 |
|--------|----------|-------------------------------------------------------------|
| `name` | `string` | The path of the file to be written to                       |
| `data` | `string` | The byte literal string or bytes to be appended to the file |

#### Examples

**basic**

Append a byte string to a file.

```python
load('file', 'append_bytes')
append_bytes('existing_file.txt', b'Hello, This is appended data.')
```

### `append_string(name, data)`

Appends a string to a file. If the file isn't present, a new file would be created.

#### Parameters

| name   | type     | description                           |
|--------|----------|---------------------------------------|
| `name` | `string` | The path of the file to be written to |
| `data` | `string` | The string to be appended to the file |

#### Examples

**basic**

Append a string to a file.

```python
load('file', 'append_string')
append_string('existing_file.txt', 'Hello, This is appended data.')
```

### `append_lines(name, data)`

Appends a list, tuple or set of lines to a file. If the file isn't present, a new file would be created.

#### Parameters

| name   | type     | description                           |
|--------|----------|---------------------------------------|
| `name` | `string` | The path of the file to be written to |
| `data` | `list    | set                                   |

#### Examples

**basic**

Append a list of lines to a file.

```python
load('file', 'append_lines')
append_lines('existing_file.txt', ['This is line1', 'This is line2', 'This is line3'])
```

### `append_json(name, data)`

Appends the given Starlark object as JSON to a file. If the file does not exist, it will be created.

#### Parameters

| name   | type     | description                            |
|--------|----------|----------------------------------------|
| `name` | `string` | The path of the file to be appended to |
| `data` | `dict    | list                                   |

#### Examples

**basic**

Append a dictionary as JSON to a file.

```python
load('file', 'append_json')
data = {"key": "value"}
append_json('existing_file.json', data)
```

### `append_jsonl(name, data)`

Appends the given data as JSON lines to a file. If the file does not exist, it will be created.

#### Parameters

| name   | type     | description                            |
|--------|----------|----------------------------------------|
| `name` | `string` | The path of the file to be appended to |
| `data` | `list    | set                                    |

#### Examples

**basic**

Append a list of JSON objects to a file as JSONL.

```python
load('file', 'append_jsonl')
data = [{"key1": "value1"}, {"key2": "value2"}]
append_jsonl('existing_file.jsonl', data)
```

### `stat(name, follow=False) FileStat`

Returns a FileStat object representing information about the given file or directory.

#### Parameters

| name     | type     | description                           |
|----------|----------|---------------------------------------|
| `name`   | `string` | The path of the file or directory.    |
| `follow` | `bool`   | If true, symbolic links are followed. |

#### Examples

**file information**

Retrieve information about a file.

```python
load('file', 'stat')
info = stat('path/to/file.txt')
print(info.name, info.size, info.type)
# Output: file.txt 3759 file
```

**directory information**

Retrieve information about a directory.

```python
load('file', 'stat')
info = stat('path/to/folder', follow=True)
print(info.name, info.size, info.type)
# Output: folder 448 dir
```

### `copyfile(src, dst, overwrite=False) string`

Copies a file from source to destination, and returns the destination file path.
If the destination exists and overwrite is set to False, an error is returned. If the destination is a directory, the file is copied into that directory with its original filename. Symbolic links are followed. Mode, access, and modification times are preserved.

#### Parameters

| name        | type     | description                                                                       |
|-------------|----------|-----------------------------------------------------------------------------------|
| `src`       | `string` | The path of the source file to be copied.                                         |
| `dst`       | `string` | The path of the destination file or directory. The parent directory must exist.   |
| `overwrite` | `bool`   | If true, allows overwriting the destination file if it exists. Defaults to False. |

#### Examples

**basic copy**

Copy a file to a new location without overwrite.

```python
load('file', 'copyfile')
src = 'path/to/source.txt'
dst = 'path/to/destination.txt'
copyfile(src, dst)
# The file at 'path/to/source.txt' is copied to 'path/to/destination.txt'
```

**overwrite copy**

Copy a file to a new location with overwrite enabled.

```python
load('file', 'copyfile')
src = 'path/to/source.txt'
dst = 'path/to/existing_destination.txt'
copyfile(src, dst, overwrite=True)
# The file at 'path/to/source.txt' is copied to 'path/to/existing_destination.txt', overwriting it.
```

**copy to directory**

Copy a file into a directory.

```python
load('file', 'copyfile')
src = 'path/to/source.txt'
dst = 'path/to/directory'
copyfile(src, dst)
# The file at 'path/to/source.txt' is copied into 'path/to/directory' with its original filename.
```

## Types

### `FileStat`

Represents information about a file.

**Fields**

| name           | type        | description                                                                                                                                                                                                                                                                                |
|----------------|-------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`         | `string`    | The name of the file.                                                                                                                                                                                                                                                                      |
| `path`         | `string`    | The full path of the file.                                                                                                                                                                                                                                                                 |
| `ext`          | `string`    | The file extension.                                                                                                                                                                                                                                                                        |
| `size`         | `int`       | The size of the file in bytes.                                                                                                                                                                                                                                                             |
| `type`         | `string`    | The type of the file: `file` for regular file, `dir` for directory, `symlink` for symbolic link, `fifo` for FIFO pipe, `socket` for network socket, `char` for character device file, `block` for block device file, `irregular` for irregular file type, `unknown` for unknown file type. |
| `modified`     | `time.Time` | The last modified time of the file.                                                                                                                                                                                                                                                        |
| `get_md5()`    | `function`  | Returns the MD5 hash of the file contents.                                                                                                                                                                                                                                                 |
| `get_sha1()`   | `function`  | Returns the SHA-1 hash of the file contents.                                                                                                                                                                                                                                               |
| `get_sha256()` | `function`  | Returns the SHA-256 hash of the file contents.                                                                                                                                                                                                                                             |
| `get_sha512()` | `function`  | Returns the SHA-512 hash of the file contents.                                                                                                                                                                                                                                             |
