package types

import (
	"fmt"

	"go.starlark.net/starlark"
)

// Unpacker is an interface for types that can be unpacked from Starlark values.
var (
	_ starlark.Unpacker = (*NullableString)(nil)
	_ starlark.Unpacker = (*NullableDict)(nil)
)

// NullableString is an Unpacker that converts a Starlark None or string to Go's string.
type NullableString struct {
	str *string
}

// NewNullableString creates and returns a new NullableString.
func NewNullableString(s string) *NullableString {
	return &NullableString{str: &s}
}

// Unpack implements Unpacker.
func (p *NullableString) Unpack(v starlark.Value) error {
	switch v := v.(type) {
	case starlark.String:
		s := string(v)
		p.str = &s
	case starlark.Bytes:
		s := string(v)
		p.str = &s
	case starlark.NoneType:
		p.str = nil
	default:
		return fmt.Errorf("got %s, want string, bytes or None", v.Type())
	}
	return nil
}

// GoString returns the Go string representation of the NullableString, if the underlying value is nil, it returns an empty string.
func (p *NullableString) GoString() string {
	if p == nil || p.str == nil {
		return ""
	}
	return *p.str
}

// IsNull returns true if the underlying value is nil.
func (p *NullableString) IsNull() bool {
	return p == nil || p.str == nil
}

// IsNullOrEmpty returns true if the underlying value is nil or an empty string.
func (p *NullableString) IsNullOrEmpty() bool {
	return p.IsNull() || p.GoString() == emptyStr
}

// NullableDict is an Unpacker that converts a Starlark None or Dict to Go's *starlark.Dict.
type NullableDict struct {
	dict *starlark.Dict
}

// Unpack implements Unpacker.
func (p *NullableDict) Unpack(v starlark.Value) error {
	switch v := v.(type) {
	case *starlark.Dict:
		p.dict = v
	case starlark.NoneType:
		p.dict = nil
	default:
		return fmt.Errorf("got %s, want dict or None", v.Type())
	}
	return nil
}

// AsDict returns the *starlark.Dict representation of the NullableDict, if the underlying dict is nil, it returns an new empty dict.
func (p *NullableDict) AsDict() *starlark.Dict {
	if p.dict == nil {
		return starlark.NewDict(0)
	}
	return p.dict
}
