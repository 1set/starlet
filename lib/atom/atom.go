// Package atom provides atomic operations for integers, floats and strings.
// Inspired by the sync/atomic and go.uber.org/atomic packages from Go.
package atom

import (
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
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
				Name:    ModuleName,
				Members: starlark.StringDict{
					//"new_int":    starlark.NewBuiltin(ModuleName+".new_int", newInt),
					//"new_float":  starlark.NewBuiltin(ModuleName+".new_float", newFloat),
					//"new_string": starlark.NewBuiltin(ModuleName+".new_string", newString),
				},
			},
		}
	})
	return atomModule, nil
}
