// Package internal contains types and utilities that are not part of the public API, and may change without notice.
// It should be only imported by the custom Starlark modules under starlet/lib folders, and not by the Starlet main package to avoid cyclic import.
package internal

import (
	"fmt"

	"go.starlark.net/starlark"
)

var (
	emptyStr string
)

// FloatOrInt is an Unpacker that converts a Starlark int or float to Go's float64.
type FloatOrInt float64

// Unpack implements Unpacker.
func (p *FloatOrInt) Unpack(v starlark.Value) error {
	switch v := v.(type) {
	case starlark.Int:
		*p = FloatOrInt(v.Float())
		return nil
	case starlark.Float:
		*p = FloatOrInt(v)
		return nil
	}
	return fmt.Errorf("got %s, want float or int", v.Type())
}

// StringOrBytes is an Unpacker that converts a Starlark string or bytes to Go's string.
// It works because Go Starlark strings and bytes are both represented as Go strings.
type StringOrBytes string

// Unpack implements Unpacker.
func (p *StringOrBytes) Unpack(v starlark.Value) error {
	switch v := v.(type) {
	case starlark.String:
		*p = StringOrBytes(v)
		return nil
	case starlark.Bytes:
		*p = StringOrBytes(v)
		return nil
	}
	return fmt.Errorf("got %s, want string or bytes", v.Type())
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

// Add takes a Starlark Value and adds it to the NumericValue.
// It returns an error if the given value is neither an int nor a float.
func (n *NumericValue) Add(value starlark.Value) error {
	switch value := value.(type) {
	case starlark.Int:
		n.intValue = n.intValue.Add(value)
	case starlark.Float:
		n.floatValue += value
		n.hasFloat = true
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
