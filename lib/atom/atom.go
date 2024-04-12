// Package atom provides atomic operations for integers, floats and strings.
// Inspired by the sync/atomic and go.uber.org/atomic packages from Go.
package atom

import (
	"fmt"
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
	"go.uber.org/atomic"
)

/*
new_int(value: int) -> AtomicInt
AtomicInt.get() -> int
AtomicInt.set(value: int)
AtomicInt.cas(old: int, new: int) -> bool
AtomicInt.add(delta: int) -> int
AtomicInt.sub(delta: int) -> int
AtomicInt.inc() -> int
AtomicInt.dec() -> int

new_float(value: float) -> AtomicFloat
AtomicFloat.get() -> float
AtomicFloat.set(value: float)
AtomicFloat.cas(old: float, new: float) -> bool
AtomicFloat.add(delta: float) -> float
AtomicFloat.sub(delta: float) -> float

new_string(value: string) -> AtomicString
AtomicString.get() -> string
AtomicString.set(value: string)
AtomicString.cas(old: string, new: string) -> bool

*/

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
					"new_int": starlark.NewBuiltin(ModuleName+".new_int", newInt),
					//"new_float":  starlark.NewBuiltin(ModuleName+".new_float", newFloat),
					//"new_string": starlark.NewBuiltin(ModuleName+".new_string", newString),
				},
			},
		}
	})
	return atomModule, nil
}

// for integer

var (
	_ starlark.Value     = (*AtomicInt)(nil)
	_ starlark.HasAttrs  = (*AtomicInt)(nil)
	_ starlark.HasBinary = (*AtomicInt)(nil)
)

func newInt(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var value int64
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "value", &value); err != nil {
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

func (a *AtomicInt) CompareSameType(op syntax.Token, y starlark.Value, depth int) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AtomicInt) Binary(op syntax.Token, y starlark.Value, side starlark.Side) (starlark.Value, error) {
	//TODO implement me
	panic("implement me")
}
