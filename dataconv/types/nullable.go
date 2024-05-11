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

type (
	// NullableInt is an Unpacker that converts a Starlark None or Int.
	NullableInt = Nullable[starlark.Int]

	// NullableFloat is an Unpacker that converts a Starlark None or Float.
	NullableFloat = Nullable[starlark.Float]

	// NullableBool is an Unpacker that converts a Starlark None or Bool.
	NullableBool = Nullable[starlark.Bool]

	// NullableString is an Unpacker that converts a Starlark None or String.
	NullableString = Nullable[starlark.String]

	// NullableBytes is an Unpacker that converts a Starlark None or Bytes.
	NullableBytes = Nullable[starlark.Bytes]

	// NullableList is an Unpacker that converts a Starlark None or List.
	NullableList = Nullable[*starlark.List]

	// NullableTuple is an Unpacker that converts a Starlark None or Tuple.
	NullableTuple = Nullable[starlark.Tuple]

	// NullableSet is an Unpacker that converts a Starlark None or Set.
	NullableSet = Nullable[*starlark.Set]

	// NullableDict is an Unpacker that converts a Starlark None or Dict.
	NullableDict = Nullable[*starlark.Dict]

	// NullableIterable is an Unpacker that converts a Starlark None or Iterable.
	NullableIterable = Nullable[starlark.Iterable]

	// NullableCallable is an Unpacker that converts a Starlark None or Callable.
	NullableCallable = Nullable[starlark.Callable]
)

// Unpacker is an interface for types that can be unpacked from Starlark values.
var (
	_ starlark.Unpacker = (*NullableInt)(nil)
	_ starlark.Unpacker = (*NullableFloat)(nil)
	_ starlark.Unpacker = (*NullableBool)(nil)
	_ starlark.Unpacker = (*NullableString)(nil)
	_ starlark.Unpacker = (*NullableBytes)(nil)
	_ starlark.Unpacker = (*NullableList)(nil)
	_ starlark.Unpacker = (*NullableTuple)(nil)
	_ starlark.Unpacker = (*NullableSet)(nil)
	_ starlark.Unpacker = (*NullableDict)(nil)
	_ starlark.Unpacker = (*NullableCallable)(nil)
	_ starlark.Unpacker = (*NullableIterable)(nil)
)

var (
	// NewNullableInt creates and returns a new NullableInt with the given default value.
	NewNullableInt = func(dv starlark.Int) *NullableInt { return NewNullable[starlark.Int](dv) }

	// NewNullableFloat creates and returns a new NullableFloat with the given default value.
	NewNullableFloat = func(dv starlark.Float) *NullableFloat { return NewNullable[starlark.Float](dv) }

	// NewNullableBool creates and returns a new NullableBool with the given default value.
	NewNullableBool = func(dv starlark.Bool) *NullableBool { return NewNullable[starlark.Bool](dv) }

	// NewNullableString creates and returns a new NullableString with the given default value.
	NewNullableString = func(dv starlark.String) *NullableString { return NewNullable[starlark.String](dv) }

	// NewNullableBytes creates and returns a new NullableBytes with the given default value.
	NewNullableBytes = func(dv starlark.Bytes) *NullableBytes { return NewNullable[starlark.Bytes](dv) }

	// NewNullableList creates and returns a new NullableList with the given default value.
	NewNullableList = func(dv *starlark.List) *NullableList { return NewNullable[*starlark.List](dv) }

	// NewNullableTuple creates and returns a new NullableTuple with the given default value.
	NewNullableTuple = func(dv starlark.Tuple) *NullableTuple { return NewNullable[starlark.Tuple](dv) }

	// NewNullableSet creates and returns a new NullableSet with the given default value.
	NewNullableSet = func(dv *starlark.Set) *NullableSet { return NewNullable[*starlark.Set](dv) }

	// NewNullableDict creates and returns a new NullableDict with the given default value.
	NewNullableDict = func(dv *starlark.Dict) *NullableDict { return NewNullable[*starlark.Dict](dv) }

	// NewNullableIterable creates and returns a new NullableIterable with the given default value.
	NewNullableIterable = func(dv starlark.Iterable) *NullableIterable { return NewNullable[starlark.Iterable](dv) }

	// NewNullableCallable creates and returns a new NullableCallable with the given default value.
	NewNullableCallable = func(dv starlark.Callable) *NullableCallable { return NewNullable[starlark.Callable](dv) }
)
