// Package dataconv provides helper functions to convert between Starlark and Go types.
// It works like package starlight, but only supports common Starlark and Go types, and won't wrap any custom types or functions.
package dataconv

import "go.starlark.net/starlark"

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