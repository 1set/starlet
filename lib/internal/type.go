// Package internal contains types and utilities that are not part of the public API, and may change without notice.
// It should be only imported by the custom Starlark modules under starlet/lib folders, and not by the Starlet main package to avoid cyclic import.
package internal

import (
	"fmt"

	"go.starlark.net/starlark"
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

// StarNumber is a custom type that implements the Starlark Value interface.
// It can be used to represent a number in Starlark.
type StarNumber struct {
	numInt   starlark.Int
	numFloat starlark.Float
	cntInt   int
	cntFloat int
}

// NewStarNumber creates a new StarNumber.
func NewStarNumber() *StarNumber {
	return &StarNumber{numInt: starlark.MakeInt(0), numFloat: starlark.Float(0)}
}

// Add adds the given value to this StarNumber.
func (n *StarNumber) Add(v starlark.Value) error {
	switch vv := v.(type) {
	case starlark.Int:
		n.numInt = n.numInt.Add(vv)
		n.cntInt++
	case starlark.Float:
		n.numFloat += vv
		n.cntFloat++
	default:
		return fmt.Errorf("got %s, want float or int", vv.Type())
	}
	return nil
}

// AsFloat returns the float value of this StarNumber.
func (n *StarNumber) AsFloat() float64 {
	return float64(n.numFloat + n.numInt.Float())
}

// Value returns the Starlark value of this StarNumber.
func (n *StarNumber) Value() starlark.Value {
	if n.cntFloat > 0 {
		return n.numFloat + n.numInt.Float()
	}
	return n.numInt
}
