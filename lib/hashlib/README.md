# hashlib

`hashlib` provides cryptographic hash primitives for Starlark — MD5, SHA-1, SHA-256, and SHA-512 digests of string or bytes input, returned as lowercase hex. Capability profile: **pure** (no filesystem, network, process, or log side effects).

Migrated from [qri-io/starlib](https://github.com/qri-io/starlib/tree/master/hash).

## Functions

| function | description |
| --- | --- |
| `md5(data) -> string` | Lowercase hex MD5 digest of `data` (string or bytes). |
| `sha1(data) -> string` | Lowercase hex SHA-1 digest of `data` (string or bytes). |
| `sha256(data) -> string` | Lowercase hex SHA-256 digest of `data` (string or bytes). |
| `sha512(data) -> string` | Lowercase hex SHA-512 digest of `data` (string or bytes). |

All four share the same signature and behavior; only the algorithm differs. `data` is a required positional argument (also accepted as the keyword `data`). It must be a Starlark `string` or `bytes`; the two encode identically, so a string and the equivalent `bytes` produce the same digest. The return is the digest hex-encoded as a lowercase `string`.

**Errors:** each function takes exactly one argument — passing more raises `got N arguments, want at most 1`. Passing a non-string/bytes value (e.g. an `int`) raises `for parameter data: got <type>, want string or bytes`.

## Examples

### `md5`

```python
load("hashlib", "md5")
print(md5(""))
print(md5("Aloha!"))
print(md5(b"Aloha!"))
# Output:
# d41d8cd98f00b204e9800998ecf8427e
# de424bf3e7dcba091c27d652ada485fb
# de424bf3e7dcba091c27d652ada485fb
```

The string and bytes forms of `"Aloha!"` hash to the same value.

### `sha1`

```python
load("hashlib", "sha1")
print(sha1(""))
print(sha1("Aloha!"))
# Output:
# da39a3ee5e6b4b0d3255bfef95601890afd80709
# c3dd37312ba987e1cc40ae021bc202c4a52d8afe
```

### `sha256`

```python
load("hashlib", "sha256")
print(sha256(""))
print(sha256("Aloha!"))
# Output:
# e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
# dea7e28aee505f2dd033de1427a517793e38b7605e8fc24da40151907e52cea3
```

### `sha512`

```python
load("hashlib", "sha512")
print(sha512(""))
print(sha512("Aloha!"))
# Output:
# cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e
# d9cb95ad9d916a0781b3339424d5eb11c476405dfba7af7fabf4981fdd3291c27e8006e4cca617beae70dd00ab86a0213c44ed461229b16b45db45f64691049e
```

### Error: wrong argument count

```python
load("hashlib", "md5")
md5("Aloha!", "Hello!")
# Output:
# Error: hash.md5: got 2 arguments, want at most 1
```

### Error: wrong input type

```python
load("hashlib", "md5")
md5(123)
# Output:
# Error: hash.md5: for parameter data: got int, want string or bytes
```

## Notes / boundaries

- **Engine:** Go standard library `crypto/md5`, `crypto/sha1`, `crypto/sha256`, `crypto/sha512`; output is `encoding/hex` lowercase.
- **Deterministic:** identical input always yields identical output; no salt, no randomness.
- **Pure:** no host effects of any kind.
- **No streaming/incremental API:** each call hashes the full `data` value in one shot; there is no reusable hasher object or `update`/`hexdigest` protocol like CPython's `hashlib`. The module exposes only the four one-shot functions above — no constants and no custom types.
- **MD5 and SHA-1 are not collision-resistant.** They are provided for checksums and legacy interop; do not use them for security-sensitive integrity or signatures — prefer `sha256` or `sha512`.
