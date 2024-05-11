// Package types provides wrappers for Starlark types that can be unpacked by the Unpack helper functions to interpret call arguments.
package types

import (
	"fmt"

	"go.starlark.net/starlark"
)

// Unpacker is an interface for types that can be unpacked from Starlark values.
var (
	_ starlark.Unpacker = (*FloatOrInt)(nil)
	_ starlark.Unpacker = (*NumericValue)(nil)
	_ starlark.Unpacker = (*StringOrBytes)(nil)
	_ starlark.Unpacker = (*NullableStringOrBytes)(nil)
)

var (
	emptyStr string
)

// FloatOrInt is an Unpacker that converts a Starlark int or float to Go's float64.
// There is no constructor for this type because it is a simple type alias of float64.
type FloatOrInt float64

// Unpack implements Unpacker.
func (p *FloatOrInt) Unpack(v starlark.Value) error {
	switch v := v.(type) {
	case starlark.Int:
		*p = FloatOrInt(v.Float())
	case starlark.Float:
		*p = FloatOrInt(v)
	default:
		return fmt.Errorf("got %s, want float or int", v.Type())
	}
	return nil
}

// GoFloat returns the Go float64 representation of the FloatOrInt.
func (p FloatOrInt) GoFloat() float64 {
	return float64(p)
}

// GoInt returns the Go int representation of the FloatOrInt.
func (p FloatOrInt) GoInt() int {
	return int(p)
}

// NumericValue holds a Starlark numeric value and tracks its type.
// It can represent an integer or a float, and it prefers integers over floats.
type NumericValue struct {
	intValue   starlark.Int
	floatValue starlark.Float
	hasFloat   bool
}

// NewNumericValue creates and returns a new NumericValue.
func NewNumericValue() *NumericValue {
	return &NumericValue{intValue: starlark.MakeInt(0), floatValue: starlark.Float(0)}
}

// Unpack implements Unpacker.
func (n *NumericValue) Unpack(v starlark.Value) error {
	switch v := v.(type) {
	case starlark.Int:
		n.intValue = v
	case starlark.Float:
		n.floatValue = v
		n.hasFloat = true
	default:
		return fmt.Errorf("got %s, want float or int", v.Type())
	}
	return nil
}

// Add takes a Starlark Value and adds it to the NumericValue.
// It returns an error if the given value is neither an int nor a float.
func (n *NumericValue) Add(value starlark.Value) error {
	switch value := value.(type) {
	case starlark.Int:
		n.intValue = n.intValue.Add(value)
	case starlark.Float:
		n.floatValue += value
		n.hasFloat = true
	case starlark.NoneType:
		// do nothing
	case nil:
		// do nothing
	default:
		return fmt.Errorf("unsupported type: %s, expected float or int", value.Type())
	}
	return nil
}

// AsFloat returns the float representation of the NumericValue.
func (n *NumericValue) AsFloat() float64 {
	return float64(n.floatValue + n.intValue.Float())
}

// Value returns the Starlark Value representation of the NumericValue.
func (n *NumericValue) Value() starlark.Value {
	if n.hasFloat {
		return starlark.Float(n.AsFloat())
	}
	return n.intValue
}

// StringOrBytes is an Unpacker that converts a Starlark string or bytes to Go's string.
// It works because Go Starlark strings and bytes are both represented as Go strings.
// There is no constructor for this type because it is a simple type alias of string.
type StringOrBytes string

// Unpack implements Unpacker.
func (p *StringOrBytes) Unpack(v starlark.Value) error {
	switch v := v.(type) {
	case starlark.String:
		*p = StringOrBytes(v)
	case starlark.Bytes:
		*p = StringOrBytes(v)
	default:
		return fmt.Errorf("got %s, want string or bytes", v.Type())
	}
	return nil
}

// GoBytes returns the Go byte slice representation of the StringOrBytes.
func (p StringOrBytes) GoBytes() []byte {
	return []byte(p)
}

// GoString returns the Go string representation of the StringOrBytes.
func (p StringOrBytes) GoString() string {
	return string(p)
}

// StarlarkString returns the Starlark string representation of the StringOrBytes.
func (p StringOrBytes) StarlarkString() starlark.String {
	return starlark.String(p)
}

// NullableStringOrBytes is an Unpacker that converts a Starlark None or string to Go's string.
type NullableStringOrBytes struct {
	str *string
}

// NewNullableStringOrBytes creates and returns a new NullableStringOrBytes.
func NewNullableStringOrBytes(s string) *NullableStringOrBytes {
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
