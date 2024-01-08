package dataconv

import "go.starlark.net/starlark"

// Unmarshaler is the interface use to unmarshal Starlark values to custom types.
type Unmarshaler interface {
	// UnmarshalStarlark unmarshal a Starlark object to custom type.
	UnmarshalStarlark(starlark.Value) error
}

// Marshaler is the interface use to marshal Starlark from custom types.
type Marshaler interface {
	// MarshalStarlark marshal a custom type to Starlark object.
	MarshalStarlark() (starlark.Value, error)
}
