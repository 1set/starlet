# HashLib

`hashlib` defines hash primitives for Starlark.

## Functions

### `md5(data) string`

Returns an MD5 hash for a string or bytes.

#### Examples

**Basic**

Calculate an MD5 checksum for "hello world".

```python
load("hashlib", "md5")
sum = md5("hello world!")
print(sum)
# Output: fc3ff98e8c6a0d3087d515c0473f8677
```

### `sha1(data) string`

Returns a SHA-1 hash for a string or bytes.

#### Examples

**Basic**

Calculate an SHA-1 checksum for "hello world".

```python
load("hashlib", "sha1")
sum = sha1("hello world!")
print(sum)
# Output: 430ce34d020724ed75a196dfc2ad67c77772d169
```

### `sha256(data) string`

Returns an SHA-256 hash for a string or bytes.

#### Examples

**Basic**

Calculate an SHA-256 checksum for "hello world".

```python
load("hashlib", "sha256")
sum = sha256("hello world!")
print(sum)
# Output: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9
```

### `sha512(data) string`

Returns an SHA-512 hash for a string or bytes.

#### Examples

**Basic**

Calculate an SHA-512 checksum for "hello world".

```python
load("hashlib", "sha512")
sum = sha512("hello world!")
print(sum)
# Output: db9b1cd3262dee37756a09b9064973589847caa8e53d31a9d142ea2701b1b28abd97838bb9a27068ba305dc8d04a45a1fcf079de54d607666996b3cc54f6b67c
```
