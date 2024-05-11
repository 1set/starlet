package types

import (
	"fmt"

	"go.starlark.net/starlark"
)

// Unpacker is an interface for types that can be unpacked from Starlark values.
var (
	_ starlark.Unpacker = (*NullableInt)(nil)
)

// Nullable is an Unpacker that converts a Starlark None or T to Go's starlark.Value.
type Nullable[T starlark.Value] struct {
	value        *T
	defaultValue T
}

// NewNullable creates and returns a new Nullable with the given default value.
func NewNullable[T starlark.Value](defaultValue T) *Nullable[T] {
	return &Nullable[T]{value: nil, defaultValue: defaultValue}
}

// Unpack implements Unpacker.
func (p *Nullable[T]) Unpack(v starlark.Value) error {
	if p == nil {
		return fmt.Errorf("nil pointer")
	}
	if _, ok := v.(starlark.NoneType); ok {
		p.value = nil
	} else if t, ok := v.(T); ok {
		p.value = &t
	} else {
		return fmt.Errorf("got %s, want %T or None", v.Type(), p.defaultValue.Type())
	}
	return nil
}

// IsNull returns true if the underlying value is nil.
func (p *Nullable[T]) IsNull() bool {
	return p == nil || p.value == nil
}

// Value returns the underlying value or default value if the underlying value is nil.
func (p *Nullable[T]) Value() T {
	if p.IsNull() {
		return p.defaultValue
	}
	return *p.value
}

// NullableInt is an Unpacker that converts a Starlark None or Int.
type NullableInt = Nullable[starlark.Int]
