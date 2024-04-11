// Package atom provides atomic operations for integers, floats and strings.
// Inspired by the sync/atomic and go.uber.org/atomic packages from Go.
package atom

import (
	"encoding/binary"
	"fmt"
	"go.uber.org/atomic"
	"hash/fnv"
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

var (
	_ starlark.Value = (*AtomicInt)(nil)
)

type AtomicInt struct {
	val    atomic.Int64
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

// hashInt64 hashes an int64 value to a uint32 hash value using little-endian byte order
func hashInt64(value int64) uint32 {
	// Allocate a byte slice
	bytes := make([]byte, 8)
	// Convert the int64 value into bytes using little-endian encoding
	binary.LittleEndian.PutUint64(bytes, uint64(value))
	// Initialize a new 32-bit FNV-1a hash
	h := fnv.New32a()
	// Write the bytes to the hasher, and ignore the error returned by Write, as hashing can't really fail here
	_, _ = h.Write(bytes)
	// Calculate the hash and return it
	return h.Sum32()
}
