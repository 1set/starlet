// Package types provides wrappers for Starlark types that can be unpacked by the Unpack helper functions to interpret call arguments.
package types

import (
	"errors"
	"fmt"
	"math"

	"go.starlark.net/starlark"
)

// Unpacker is an interface for types that can be unpacked from Starlark values.
var (
	_ starlark.Unpacker = (*FloatOrInt)(nil)
	_ starlark.Unpacker = (*FloatOrIntList)(nil)
	_ starlark.Unpacker = (*NumericValue)(nil)
	_ starlark.Unpacker = (*StringOrBytes)(nil)
	_ starlark.Unpacker = (*NullableStringOrBytes)(nil)
)

var (
	emptyStr       string
	errNilReceiver = errors.New("nil pointer")
	gotStarType    = func(v starlark.Value) string {
		if v == nil {
			return "nil"
		}
		return v.Type()
	}
)

// FloatOrInt is an Unpacker that converts a Starlark int or float to Go's float64.
// There is no constructor for this type because it is a simple type alias of float64.
type FloatOrInt float64

// Unpack implements Unpacker.
func (p *FloatOrInt) Unpack(v starlark.Value) error {
	if p == nil {
		return errNilReceiver
	}
	switch v := v.(type) {
	case starlark.Int:
		*p = FloatOrInt(v.Float())
	case starlark.Float:
		*p = FloatOrInt(v)
	default:
		return fmt.Errorf("got %s, want float or int", gotStarType(v))
	}
	return nil
}

// GoFloat returns the Go float64 representation of the FloatOrInt.
func (p FloatOrInt) GoFloat() float64 {
	return float64(p)
}

// GoFloat32 returns the Go float32 representation of the FloatOrInt.
func (p FloatOrInt) GoFloat32() float32 {
	return float32(p)
}

// GoFloat64 returns the Go float64 representation of the FloatOrInt.
func (p FloatOrInt) GoFloat64() float64 {
	return float64(p)
}

// GoInt returns the Go int representation of the FloatOrInt.
func (p FloatOrInt) GoInt() int {
	return int(clampToRange(float64(p), float64(math.MinInt), float64(math.MaxInt)))
}

// GoInt32 returns the Go int32 representation of the FloatOrInt.
func (p FloatOrInt) GoInt32() int32 {
	return int32(clampToRange(float64(p), float64(math.MinInt32), float64(math.MaxInt32)))
}

// GoInt64 returns the Go int64 representation of the FloatOrInt.
func (p FloatOrInt) GoInt64() int64 {
	return int64(clampToRange(float64(p), float64(math.MinInt64), float64(math.MaxInt64)))
}

// clampToRange clamps the input float64 value to the specified min and max range
func clampToRange(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// FloatOrIntList is an Unpacker that converts a Starlark list of floats or ints to a Go slice of float64.
type FloatOrIntList []float64

// Unpack implements Unpacker.
func (l *FloatOrIntList) Unpack(v starlark.Value) error {
	// ensure the input is iterable
	if l == nil {
		return errNilReceiver
	}
	list, ok := v.(starlark.Iterable)
	if !ok {
		return fmt.Errorf("got %s, want iterable", gotStarType(v))
	}

	// create a new slice to hold the unpacked values
	result := make([]float64, 0)

	// iterate through the Starlark list
	iter := list.Iterate()
	defer iter.Done()
	var item starlark.Value
	for i := 0; iter.Next(&item); i++ {
		//for iter.Next(&item) {
		switch v := item.(type) {
		case starlark.Int:
			// Convert Starlark int to float64
			result = append(result, float64(v.Float()))
		case starlark.Float:
			// Append Starlark float directly
			result = append(result, float64(v))
		default:
			// Return an error for unsupported types
			return fmt.Errorf("at index %d, got unsupported type %s, want float or int", i, gotStarType(v))
		}
	}

	// assign the result to the receiver
	*l = result
	return nil
}

// GoSlice returns the underlying Go []float64 slice.
func (l FloatOrIntList) GoSlice() []float64 {
	return []float64(l)
}

// StarlarkList returns a new Starlark list containing the values as Starlark floats.
func (l FloatOrIntList) StarlarkList() *starlark.List {
	items := make([]starlark.Value, len(l))
	for i, v := range l {
		items[i] = starlark.Float(v)
	}
	return starlark.NewList(items)
}

// Len returns the length of the FloatOrIntList.
func (l FloatOrIntList) Len() int {
	return len(l)
}

// IsEmpty returns true if the FloatOrIntList is empty.
func (l FloatOrIntList) IsEmpty() bool {
	return len(l) == 0
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
	if n == nil {
		return errNilReceiver
	}
	switch v := v.(type) {
	case starlark.Int:
		n.intValue = v
	case starlark.Float:
		n.floatValue = v
		n.hasFloat = true
	default:
		return fmt.Errorf("got %s, want float or int", gotStarType(v))
	}
	return nil
}

// Add takes a Starlark Value and adds it to the NumericValue.
// It returns an error if the given value is neither an int nor a float.
func (n *NumericValue) Add(v starlark.Value) error {
	switch v := v.(type) {
	case starlark.Int:
		n.intValue = n.intValue.Add(v)
	case starlark.Float:
		n.floatValue += v
		n.hasFloat = true
	case starlark.NoneType:
		// do nothing
	case nil:
		// do nothing
	default:
		return fmt.Errorf("unsupported type: %s, expected float or int", gotStarType(v))
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
	if p == nil {
		return errNilReceiver
	}
	switch v := v.(type) {
	case starlark.String:
		*p = StringOrBytes(v)
	case starlark.Bytes:
		*p = StringOrBytes(v)
	default:
		return fmt.Errorf("got %s, want string or bytes", gotStarType(v))
	}
	return nil
}

// GoString returns the Go string representation of the StringOrBytes.
func (p StringOrBytes) GoString() string {
	return string(p)
}

// GoBytes returns the Go byte slice representation of the StringOrBytes.
func (p StringOrBytes) GoBytes() []byte {
	return []byte(p)
}

// StarlarkString returns the Starlark string representation of the StringOrBytes.
func (p StringOrBytes) StarlarkString() starlark.String {
	return starlark.String(p)
}

// StarlarkBytes returns the Starlark bytes representation of the StringOrBytes.
func (p StringOrBytes) StarlarkBytes() starlark.Bytes {
	return starlark.Bytes(p)
}

// IsEmpty returns true if the underlying value is an empty string.
func (p StringOrBytes) IsEmpty() bool {
	return p.GoString() == emptyStr
}

// NullableStringOrBytes is an Unpacker that converts a Starlark None or string to Go's string.
type NullableStringOrBytes struct {
	str *string
}

// NewNullableStringOrBytes creates and returns a new NullableStringOrBytes.
func NewNullableStringOrBytes(s string) *NullableStringOrBytes {
	return &NullableStringOrBytes{str: &s}
}

// NewNullableStringOrBytesNoDefault creates and returns a new NullableStringOrBytes without a default value.
func NewNullableStringOrBytesNoDefault() *NullableStringOrBytes {
	return &NullableStringOrBytes{str: nil}
}

// Unpack implements Unpacker.
func (p *NullableStringOrBytes) Unpack(v starlark.Value) error {
	if p == nil {
		return errNilReceiver
	}
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
		return fmt.Errorf("got %s, want string, bytes or None", gotStarType(v))
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

// GoBytes returns the Go byte slice representation of the NullableStringOrBytes, if the underlying value is nil, it returns nil.
func (p *NullableStringOrBytes) GoBytes() []byte {
	if p == nil || p.str == nil {
		return nil
	}
	return []byte(*p.str)
}

// StarlarkString returns the Starlark string representation of the NullableStringOrBytes, if the underlying value is nil, it returns a Starlark string with an empty string.
func (p *NullableStringOrBytes) StarlarkString() starlark.String {
	if p == nil || p.str == nil {
		return ""
	}
	return starlark.String(*p.str)
}

// StarlarkBytes returns the Starlark bytes representation of the NullableStringOrBytes, if the underlying value is nil, it returns a Starlark bytes with an empty string.
func (p *NullableStringOrBytes) StarlarkBytes() starlark.Bytes {
	if p == nil || p.str == nil {
		return ""
	}
	return starlark.Bytes(*p.str)
}

// IsNull returns true if the underlying value is nil.
func (p *NullableStringOrBytes) IsNull() bool {
	return p == nil || p.str == nil
}

// IsNullOrEmpty returns true if the underlying value is nil or an empty string.
func (p *NullableStringOrBytes) IsNullOrEmpty() bool {
	return p.IsNull() || p.GoString() == emptyStr
}
