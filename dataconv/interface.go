// Package dataconv provides helper functions to convert between Starlark and Go types.
//
// It works like package starlight, but only supports common Starlark and Go types, and won't wrap any custom types or functions.
//
// For data type conversion, it provides functions to convert between Starlark and Go types:
//
//     +---------+   Marshal   +------------+   MarshalStarlarkJSON   +----------+
//     |         | ----------> |            | ----------------------> |          |
//     |  Go     |             |  Starlark  |                         | JSON     |
//     |  Value  | <---------- |  Value     | <---------------------- | String   |
//     |         |  Unmarshal  |            |   UnmarshalStarlarkJSON |          |
//     +---------+             +------------+                         +----------+
//
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
