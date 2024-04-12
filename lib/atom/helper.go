package atom

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"sort"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func builtinAttr(recv starlark.Value, name string, methods map[string]*starlark.Builtin) (starlark.Value, error) {
	b := methods[name]
	if b == nil {
		return nil, nil // no such method
	}
	return b.BindReceiver(recv), nil
}

func builtinAttrNames(methods map[string]*starlark.Builtin) []string {
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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

// hashFloat64 hashes a float64 value to a uint32 hash value
func hashFloat64(value float64) uint32 {
	// Convert the float64 value into its binary representation as uint64
	bits := math.Float64bits(value)
	// Allocate a byte slice
	bytes := make([]byte, 8)
	// Convert the uint64 bits into bytes using little-endian encoding
	binary.LittleEndian.PutUint64(bytes, bits)
	// Initialize a new 32-bit FNV-1a hash
	h := fnv.New32a()
	// Write the bytes to the hasher, and ignore the error returned by Write, as hashing can't really fail here
	_, _ = h.Write(bytes)
	// Calculate the hash and return it
	return h.Sum32()
}

// hashString hashes a string value to a uint32 hash value
func hashString(value string) uint32 {
	// Initialize a new 32-bit FNV-1a hash
	h := fnv.New32a()
	// Write the string to the hasher, and ignore the error returned by Write, as hashing can't really fail here
	_, _ = h.Write([]byte(value))
	// Calculate the hash and return it
	return h.Sum32()
}

// threewayCompare interprets a three-way comparison value cmp (-1, 0, +1)
// as a boolean comparison (e.g. x < y).
func threewayCompare(op syntax.Token, cmp int) (bool, error) {
	switch op {
	case syntax.EQL:
		return cmp == 0, nil
	case syntax.NEQ:
		return cmp != 0, nil
	case syntax.LE:
		return cmp <= 0, nil
	case syntax.LT:
		return cmp < 0, nil
	case syntax.GE:
		return cmp >= 0, nil
	case syntax.GT:
		return cmp > 0, nil
	default:
		return false, fmt.Errorf("unexpected comparison operator %s", op)
	}
}
