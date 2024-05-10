package internal

import (
	"fmt"

	"go.starlark.net/starlark"
)

var (
	emptyStr string
)

// Unpacker is an interface for types that can be unpacked from Starlark values.
var (
	_ starlark.Unpacker = (*FloatOrInt)(nil)
	_ starlark.Unpacker = (*StringOrBytes)(nil)
	_ starlark.Unpacker = (*NullableString)(nil)
	_ starlark.Unpacker = (*NullableDict)(nil)
	//_ starlark.Unpacker = (*NumericValue)(nil)
)

// FloatOrInt is an Unpacker that converts a Starlark int or float to Go's float64.
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

// StringOrBytes is an Unpacker that converts a Starlark string or bytes to Go's string.
// It works because Go Starlark strings and bytes are both represented as Go strings.
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
