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
