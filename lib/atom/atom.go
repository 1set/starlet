// Package atom provides atomic operations for integers, floats and strings, inspired by the sync/atomic and go.uber.org/atomic Go packages.
package atom

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
