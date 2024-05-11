package types

import (
	"fmt"

	"go.starlark.net/starlark"
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
		return fmt.Errorf("got %s, want %s or None", v.Type(), p.defaultValue.Type())
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

// NullableFloat is an Unpacker that converts a Starlark None or Float.
type NullableFloat = Nullable[starlark.Float]

// NullableBool is an Unpacker that converts a Starlark None or Bool.
type NullableBool = Nullable[starlark.Bool]

// NullableList is an Unpacker that converts a Starlark None or List.
type NullableList = Nullable[*starlark.List]

// NullableTuple is an Unpacker that converts a Starlark None or Tuple.
type NullableTuple = Nullable[starlark.Tuple]

// NullableSet is an Unpacker that converts a Starlark None or Set.
type NullableSet = Nullable[*starlark.Set]

// NullableIterable is an Unpacker that converts a Starlark None or Iterable.
type NullableIterable = Nullable[starlark.Iterable]

// NullableCallable is an Unpacker that converts a Starlark None or Callable.
type NullableCallable = Nullable[starlark.Callable]

// Unpacker is an interface for types that can be unpacked from Starlark values.
var (
	_ starlark.Unpacker = (*NullableInt)(nil)
	_ starlark.Unpacker = (*NullableFloat)(nil)
	_ starlark.Unpacker = (*NullableBool)(nil)
	_ starlark.Unpacker = (*NullableList)(nil)
	_ starlark.Unpacker = (*NullableTuple)(nil)
	_ starlark.Unpacker = (*NullableSet)(nil)
	_ starlark.Unpacker = (*NullableCallable)(nil)
	_ starlark.Unpacker = (*NullableIterable)(nil)
)
