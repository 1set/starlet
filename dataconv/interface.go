// Package dataconv provides helper functions to convert between Starlark and Go types.
//
// It works like package starlight, but only supports common Starlark and Go types, and won't wrap any custom types or functions.
//
// For data type conversion, it provides functions to convert between Starlark and Go types:
//
//	+---------+   Marshal   +------------+   MarshalStarlarkJSON   +----------+
//	|         | ----------> |            | ----------------------> |          |
//	|  Go     |             |  Starlark  |                         | JSON     |
//	|  Value  | <---------- |  Value     | <---------------------- | String   |
//	|         |  Unmarshal  |            |   UnmarshalStarlarkJSON |          |
//	+---------+             +------------+                         +----------+
//
// Which function to use:
//
//   - Starlark value -> native Go value: Unmarshal. Ints stay platform int
//     in range and degrade to uint64/*big.Int beyond it; dicts come back as
//     map[string]interface{} with non-string keys stringified (JSON-ready).
//   - Go value -> Starlark value: Marshal (common types only; package
//     starlight wraps everything else).
//   - Starlark value -> JSON text and back, via the Go shapes above:
//     MarshalStarlarkJSON / UnmarshalStarlarkJSON. The decode direction
//     applies TypeConvert heuristics (RFC3339-looking strings become time
//     values) and maps numbers by literal form — an integer literal becomes
//     an int (exact, arbitrary precision), a number with a decimal point or
//     exponent becomes a float.
//   - Starlark value -> JSON text and back, staying inside Starlark types:
//     EncodeStarlarkJSON / DecodeStarlarkJSON (the interpreter's own json
//     encoder: big ints work, bytes/time are errors, no heuristics).
//
// The starlight convert package remains the third conversion surface used
// by the Machine API itself (Run/Call results): its FromValue returns
// int64-first integers and fidelity-preserving typed dict keys, where this
// package prefers JSON-friendly shapes.
package dataconv

import "go.starlark.net/starlark"

// StarlarkFunc is a function that can be called from Starlark.
type StarlarkFunc func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)

// Unmarshaler is the interface use to unmarshal Starlark values to custom types, i.e. Starlark to Go.
type Unmarshaler interface {
	// UnmarshalStarlark unmarshal a Starlark object to custom type.
	UnmarshalStarlark(starlark.Value) error
}

// Marshaler is the interface use to marshal Starlark from custom types, i.e. Go to Starlark.
type Marshaler interface {
	// MarshalStarlark marshal a custom type to Starlark object.
	MarshalStarlark() (starlark.Value, error)
}
