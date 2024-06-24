package types

import (
	"fmt"

	"go.starlark.net/starlark"
)

// EitherOrNone is an Unpacker that converts a Starlark None, A, or B to Go's starlark.Value.
type EitherOrNone[A starlark.Value, B starlark.Value] struct {
	value   starlark.Value
	isNone  bool
	isTypeA bool
	isTypeB bool
}

// NewEitherOrNone creates and returns a new EitherOrNone.
func NewEitherOrNone[A starlark.Value, B starlark.Value]() *EitherOrNone[A, B] {
	return &EitherOrNone[A, B]{isNone: true}
}

// Unpack implements the starlark.Unpacker interface.
func (e *EitherOrNone[A, B]) Unpack(v starlark.Value) error {
	if e == nil {
		return fmt.Errorf("nil pointer")
	}
	if _, ok := v.(starlark.NoneType); ok {
		e.value = nil
		e.isNone, e.isTypeA, e.isTypeB = true, false, false
	} else if a, ok := v.(A); ok {
		e.value = a
		e.isNone, e.isTypeA, e.isTypeB = false, true, false
	} else if b, ok := v.(B); ok {
		e.value = b
		e.isNone, e.isTypeA, e.isTypeB = false, false, true
	} else {
		var zeroA A
		var zeroB B
		gt := "nil"
		if v != nil {
			gt = v.Type()
		}
		return fmt.Errorf("expected %T or %T or None, got %s", zeroA, zeroB, gt)
	}
	return nil
}

// IsNone returns true if the value is None.
func (e *EitherOrNone[A, B]) IsNone() bool {
	return e != nil && e.isNone
}

// IsTypeA returns true if the value is of type A.
func (e *EitherOrNone[A, B]) IsTypeA() bool {
	return e != nil && e.isTypeA
}

// IsTypeB returns true if the value is of type B.
func (e *EitherOrNone[A, B]) IsTypeB() bool {
	return e != nil && e.isTypeB
}

// Value returns the underlying value. You can use IsTypeA and IsTypeB to check which type it is.
func (e *EitherOrNone[A, B]) Value() starlark.Value {
	if e == nil {
		var zero starlark.Value
		return zero
	}
	return e.value
}

// ValueA returns the value of type A, if available, and a boolean indicating its presence.
func (e *EitherOrNone[A, B]) ValueA() (A, bool) {
	if e != nil && e.isTypeA {
		return e.value.(A), true
	}
	var zero A
	return zero, false
}

// ValueB returns the value of type B, if available, and a boolean indicating its presence.
func (e *EitherOrNone[A, B]) ValueB() (B, bool) {
	if e != nil && e.isTypeB {
		return e.value.(B), true
	}
	var zero B
	return zero, false
}

// Type returns the type of the underlying value.
func (e *EitherOrNone[A, B]) Type() string {
	if e == nil {
		return "NilReceiver"
	}
	if e.isNone {
		return starlark.None.Type()
	}
	if e.isTypeA {
		var a A
		return a.Type()
	}
	if e.isTypeB {
		var b B
		return b.Type()
	}
	return "Unknown"
}

// Unpacker interface implementation check
var (
	_ starlark.Unpacker = (*EitherOrNone[*starlark.List, *starlark.Dict])(nil)
	_ starlark.Unpacker = (*EitherOrNone[starlark.String, starlark.Int])(nil)
)
