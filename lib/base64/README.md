# base64

`base64` provides base64 encoding and decoding for Starlark, commonly used to represent binary data as ASCII text. It wraps Go's `encoding/base64` and supports the standard and URL-safe alphabets, each with padded and raw (unpadded) variants. Capability profile: **pure** (no filesystem, network, process, or log side effects).

## Functions

| function | description |
|----------|-------------|
| `encode(data, encoding="standard") -> string` | base64-encode a string or bytes, returning the encoded text |
| `decode(data, encoding="standard") -> string` | base64-decode a string or bytes, returning the decoded text |

## Encoding dialects

The optional `encoding` argument selects the alphabet and padding. An empty string is treated as `"standard"`. Any other value raises `unsupported encoding format: <value>`.

| encoding | meaning |
|----------|---------|
| `"standard"` | standard base64 with padding, RFC 4648 (default) |
| `"standard_raw"` | standard base64 without padding, RFC 4648 §3.2 |
| `"url"` | URL- and filename-safe base64 with padding, RFC 4648 |
| `"url_raw"` | URL- and filename-safe base64 without padding |

## Details & examples

### `encode`

`encode(data, encoding="standard") -> string`

`data` is the input to encode and accepts a `string` or `bytes`; any other type raises `base64.encode: for parameter data: got <type>, want string or bytes`. The result is always a `string`. Errors on an unknown `encoding` value (see the dialects table).

```python
load("base64", "encode")
print(encode("hello"))
print(encode("hello", encoding="standard_raw"))
print(encode("hello friend!", encoding="url"))
print(encode("hello friend!", encoding="url_raw"))
# Output:
# aGVsbG8=
# aGVsbG8
# aGVsbG8gZnJpZW5kIQ==
# aGVsbG8gZnJpZW5kIQ
```

### `decode`

`decode(data, encoding="standard") -> string`

`data` is the base64-encoded input and accepts a `string` or `bytes`; any other type raises `base64.decode: for parameter data: got <type>, want string or bytes`. The result is the decoded text as a `string`. Errors on an unknown `encoding` value, and on malformed input for the chosen dialect (e.g. decoding the unpadded `"aGVsbG8"` with the default `standard` encoding raises `illegal base64 data at input byte 4` — use `encoding="standard_raw"` for unpadded input).

```python
load("base64", "decode")
print(decode("aGVsbG8="))
print(decode("aGVsbG8", encoding="standard_raw"))
print(decode("aGVsbG8gZnJpZW5kIQ==", encoding="url"))
print(decode("aGVsbG8gZnJpZW5kIQ", encoding="url_raw"))
# Output:
# hello
# hello
# hello friend!
# hello friend!
```

## Notes / boundaries

- Engine: thin wrapper over Go's standard `encoding/base64`; behavior and error messages follow that package.
- The padded dialects (`standard`, `url`) require correct `=` padding on decode; the raw dialects (`standard_raw`, `url_raw`) require its absence. Mixing them raises an `illegal base64 data` error.
- Deterministic: the same input and dialect always yield the same output.
- `decode` returns a `string`; decoded bytes that are not valid UTF-8 are still returned as a Starlark string holding those bytes.
