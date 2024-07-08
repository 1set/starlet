package types

import (
	"fmt"

	"go.starlark.net/starlark"
)

// OneOrMany is a struct that can hold either a single value or multiple values of a specific type, and can be unpacked from a Starlark value.
type OneOrMany[T starlark.Value] struct {
	values       []T
	defaultValue T
	hasDefault   bool
}

// NewOneOrMany creates and returns a new OneOrMany with the given default value.
func NewOneOrMany[T starlark.Value](defaultValue T) *OneOrMany[T] {
	return &OneOrMany[T]{values: nil, defaultValue: defaultValue, hasDefault: true}
}

// NewOneOrManyNoDefault creates and returns a new OneOrMany without a default value.
func NewOneOrManyNoDefault[T starlark.Value]() *OneOrMany[T] {
	return &OneOrMany[T]{values: nil, hasDefault: false}
}

// Unpack implements the starlark.Unpacker interface, allowing the struct to unpack from a starlark.Value.
func (o *OneOrMany[T]) Unpack(v starlark.Value) error {
	if o == nil {
		return errNilReceiver
	}
	if _, ok := v.(starlark.NoneType); ok {
		// None
		o.values = nil
	} else if t, ok := v.(T); ok {
		// Single value
		o.values = []T{t}
	} else if l, ok := v.(starlark.Iterable); ok {
		// List or Tuple or Set of values
		sl := make([]T, 0, 1)
		iter := l.Iterate()
		defer iter.Done()
		// Iterate over the iterable
		var x starlark.Value
		for iter.Next(&x) {
			if t, ok := x.(T); ok {
				sl = append(sl, t)
			} else {
				return fmt.Errorf("expected %T, got %s", o.defaultValue, gotStarType(x))
			}
		}
		o.values = sl
	} else {
		return fmt.Errorf("expected %T or Iterable or None, got %s", o.defaultValue, gotStarType(v))
	}
	return nil
}

// IsNull checks if the struct is nil or has no underlying slice. It returns true if the struct is nil or has no underlying slice, no matter if a default value is set.
func (o *OneOrMany[T]) IsNull() bool {
	return o == nil || o.values == nil
}

// Len returns the length of the underlying slice or default value.
func (o *OneOrMany[T]) Len() int {
	if o == nil {
		return 0
	}
	if o.values == nil {
		if o.hasDefault {
			return 1
		}
		return 0
	}
	return len(o.values)
}

// Slice returns the underlying slice, or a slice containing the default value if the slice is nil and a default is set.
// It returns an empty slice if the underlying slice is nil and has no default value.
func (o *OneOrMany[T]) Slice() []T {
	if o == nil {
		return []T{}
	}
	if o.values == nil {
		if o.hasDefault {
			return []T{o.defaultValue}
		}
		return []T{}
	}
	return o.values
}

// First returns the first element in the slice, or the default value if the slice is nil or empty and a default is set.
// It returns the zero value of the type if the underlying slice is nil and has no default value.
func (o *OneOrMany[T]) First() T {
	if o == nil || o.values == nil || len(o.values) == 0 {
		if o != nil && o.hasDefault {
			return o.defaultValue
		}
		var zero T
		return zero
	}
	return o.values[0]
}

// Unpacker is an interface for types that can be unpacked from Starlark values.
var (
	_ starlark.Unpacker = (*OneOrMany[starlark.Int])(nil)
	_ starlark.Unpacker = (*OneOrMany[starlark.String])(nil)
	_ starlark.Unpacker = (*OneOrMany[*starlark.Dict])(nil)
)
