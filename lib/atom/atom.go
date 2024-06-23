// Package atom provides atomic operations for integers, floats and strings.
// Inspired by the sync/atomic and go.uber.org/atomic packages from Go.
package atom

import (
	"fmt"
	"strings"
	"sync"

	tps "github.com/1set/starlet/dataconv/types"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
	"go.uber.org/atomic"
)

// ModuleName defines the expected name for this Module when used in starlark's load() function, eg: load('atom', 'new_int')
const ModuleName = "atom"

var (
	once       sync.Once
	atomModule starlark.StringDict
)

// LoadModule loads the atom module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		atomModule = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"new_int":    starlark.NewBuiltin(ModuleName+".new_int", newInt),
					"new_float":  starlark.NewBuiltin(ModuleName+".new_float", newFloat),
					"new_string": starlark.NewBuiltin(ModuleName+".new_string", newString),
				},
			},
		}
	})
	return atomModule, nil
}

// for integer

var (
	_ starlark.Value      = (*AtomicInt)(nil)
	_ starlark.HasAttrs   = (*AtomicInt)(nil)
	_ starlark.Comparable = (*AtomicInt)(nil)
)

func newInt(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var value int64
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "value?", &value); err != nil {
		return nil, err
	}
	return &AtomicInt{val: atomic.NewInt64(value)}, nil
}

type AtomicInt struct {
	val    *atomic.Int64
	frozen bool
}

func (a *AtomicInt) String() string {
	return fmt.Sprintf("<atom_int:%d>", a.val.Load())
}

func (a *AtomicInt) Type() string {
	return "atom_int"
}

func (a *AtomicInt) Freeze() {
	a.frozen = true
}

func (a *AtomicInt) Truth() starlark.Bool {
	return a.val.Load() != 0
}

func (a *AtomicInt) Hash() (uint32, error) {
	//return 0, fmt.Errorf("unhashable: %s", a.Type())
	return hashInt64(a.val.Load()), nil
}

func (a *AtomicInt) Attr(name string) (starlark.Value, error) {
	return builtinAttr(a, name, intMethods)
}

func (a *AtomicInt) AttrNames() []string {
	return builtinAttrNames(intMethods)
}

func (a *AtomicInt) CompareSameType(op syntax.Token, y_ starlark.Value, depth int) (bool, error) {
	vx := a.val.Load()
	y := y_.(*AtomicInt)
	vy := y.val.Load()

	cmp := 0
	if vx < vy {
		cmp = -1
	} else if vx > vy {
		cmp = 1
	} else {
		cmp = 0
	}
	return threewayCompare(op, cmp)
}

// for float

var (
	_ starlark.Value      = (*AtomicFloat)(nil)
	_ starlark.HasAttrs   = (*AtomicFloat)(nil)
	_ starlark.Comparable = (*AtomicFloat)(nil)
)

type AtomicFloat struct {
	val    *atomic.Float64
	frozen bool
}

func newFloat(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var value tps.FloatOrInt
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "value?", &value); err != nil {
		return nil, err
	}
	return &AtomicFloat{val: atomic.NewFloat64(value.GoFloat())}, nil
}

func (a *AtomicFloat) String() string {
	return fmt.Sprintf("<atom_float:%v>", a.val.Load())
}

func (a *AtomicFloat) Type() string {
	return "atom_float"
}

func (a *AtomicFloat) Freeze() {
	a.frozen = true
}

func (a *AtomicFloat) Truth() starlark.Bool {
	return a.val.Load() != 0
}

func (a *AtomicFloat) Hash() (uint32, error) {
	return hashFloat64(a.val.Load()), nil
}

func (a *AtomicFloat) Attr(name string) (starlark.Value, error) {
	return builtinAttr(a, name, floatMethods)
}

func (a *AtomicFloat) AttrNames() []string {
	return builtinAttrNames(floatMethods)
}

func (a *AtomicFloat) CompareSameType(op syntax.Token, y_ starlark.Value, depth int) (bool, error) {
	vx := a.val.Load()
	y := y_.(*AtomicFloat)
	vy := y.val.Load()

	cmp := 0
	if vx < vy {
		cmp = -1
	} else if vx > vy {
		cmp = 1
	} else {
		cmp = 0
	}
	return threewayCompare(op, cmp)
}

// for string

var (
	_ starlark.Value      = (*AtomicString)(nil)
	_ starlark.HasAttrs   = (*AtomicString)(nil)
	_ starlark.Comparable = (*AtomicString)(nil)
)

type AtomicString struct {
	val    *atomic.String
	frozen bool
}

func newString(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var value string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "value?", &value); err != nil {
		return nil, err
	}
	return &AtomicString{val: atomic.NewString(value)}, nil
}

func (a *AtomicString) String() string {
	return fmt.Sprintf("<atom_string:%q>", a.val.Load())
}

func (a *AtomicString) Type() string {
	return "atom_string"
}

func (a *AtomicString) Freeze() {
	a.frozen = true
}

func (a *AtomicString) Truth() starlark.Bool {
	return a.val.Load() != ""
}

func (a *AtomicString) Hash() (uint32, error) {
	return hashString(a.val.Load()), nil
}

func (a *AtomicString) Attr(name string) (starlark.Value, error) {
	return builtinAttr(a, name, stringMethods)
}

func (a *AtomicString) AttrNames() []string {
	return builtinAttrNames(stringMethods)
}

func (a *AtomicString) CompareSameType(op syntax.Token, y_ starlark.Value, depth int) (bool, error) {
	vx := a.val.Load()
	y := y_.(*AtomicString)
	vy := y.val.Load()

	return threewayCompare(op, strings.Compare(vx, vy))
}
