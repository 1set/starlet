package types

import (
	"fmt"

	"go.starlark.net/starlark"
)

// Unpacker is an interface for types that can be unpacked from Starlark values.
var (
	_ starlark.Unpacker = (*NullableStringOrBytes)(nil)
	_ starlark.Unpacker = (*NullableDict)(nil)
)

// NullableStringOrBytes is an Unpacker that converts a Starlark None or string to Go's string.
type NullableStringOrBytes struct {
	str *string
}

// NewNullableString creates and returns a new NullableStringOrBytes.
func NewNullableString(s string) *NullableStringOrBytes {
	return &NullableStringOrBytes{str: &s}
}

// Unpack implements Unpacker.
func (p *NullableStringOrBytes) Unpack(v starlark.Value) error {
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

// GoString returns the Go string representation of the NullableStringOrBytes, if the underlying value is nil, it returns an empty string.
func (p *NullableStringOrBytes) GoString() string {
	if p == nil || p.str == nil {
		return ""
	}
	return *p.str
}

// IsNull returns true if the underlying value is nil.
func (p *NullableStringOrBytes) IsNull() bool {
	return p == nil || p.str == nil
}

// IsNullOrEmpty returns true if the underlying value is nil or an empty string.
func (p *NullableStringOrBytes) IsNullOrEmpty() bool {
	return p.IsNull() || p.GoString() == emptyStr
}
