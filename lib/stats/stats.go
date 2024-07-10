// Package stats provides a Starlark module for comprehensive statistics functions. It's a wrapper around the Go package: https://github.com/montanaflynn/stats
package stats

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"sync"

	tps "github.com/1set/starlet/dataconv/types"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used
// in starlark's load() function, eg: load('stats', 'md5')
const ModuleName = "stats"

var (
	once       sync.Once
	hashModule starlark.StringDict
	hashError  error
)

// LoadModule loads the hashlib module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		hashModule = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"md5":    starlark.NewBuiltin("hash.md5", fnHash(md5.New)),
					"sha1":   starlark.NewBuiltin("hash.sha1", fnHash(sha1.New)),
					"sha256": starlark.NewBuiltin("hash.sha256", fnHash(sha256.New)),
					"sha512": starlark.NewBuiltin("hash.sha512", fnHash(sha512.New)),
				},
			},
		}
	})
	return hashModule, hashError
}

func fnHash(algo func() hash.Hash) func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
	return func(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		// check args
		var sb tps.StringOrBytes
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "data", &sb); err != nil {
			return starlark.None, err
		}

		// get hash
		h := algo()
		_, err := io.WriteString(h, sb.GoString())
		if err != nil {
			return starlark.None, err
		}
		return starlark.String(hex.EncodeToString(h.Sum(nil))), nil
	}
}

// floatOrIntList is an Unpacker that converts a Starlark list of int or float to Go's []float64.
type floatOrIntList []float64

func (p *floatOrIntList) Unpack(v starlark.Value) error {
	if list, ok := v.(*starlark.List); ok {
		for i := 0; i < list.Len(); i++ {
			elem := list.Index(i)
			switch elem := elem.(type) {
			case starlark.Int:
				*p = append(*p, float64(elem.Float()))
			case starlark.Float:
				*p = append(*p, float64(elem))
			default:
				return fmt.Errorf("list element %d: got %s, want float or int", i, elem.Type())
			}
		}
		return nil
	}
	return fmt.Errorf("got %s, want list", v.Type())
}

// floatOrInt is an Unpacker that converts a Starlark int or float to Go's float64.
type floatOrInt float64

func (p *floatOrInt) Unpack(v starlark.Value) error {
	switch v := v.(type) {
	case starlark.Int:
		*p = floatOrInt(v.Float())
		return nil
	case starlark.Float:
		*p = floatOrInt(v)
		return nil
	}
	return fmt.Errorf("got %s, want float or int", v.Type())
}

// newUnaryBuiltin wraps a unary function accepting []float64 and returning (float64, error) as a Starlark built-in.
func newUnaryBuiltin(name string, fn func([]float64) (float64, error)) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data floatOrIntList
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 1, &data); err != nil {
			return nil, err
		}
		result, err := fn(data)
		if err != nil {
			return nil, err
		}
		return starlark.Float(result), nil
	})
}

// newBinaryBuiltin wraps a binary function accepting []float64 and []float64, returning (float64, error) as a Starlark built-in.
func newBinaryBuiltin(name string, fn func([]float64, []float64) (float64, error)) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data1, data2 floatOrIntList
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 2, &data1, &data2); err != nil {
			return nil, err
		}
		result, err := fn(data1, data2)
		if err != nil {
			return nil, err
		}
		return starlark.Float(result), nil
	})
}

// newTernaryBuiltin wraps a ternary function accepting []float64, []float64, float64, returning (float64, error) as a Starlark built-in.
func newTernaryBuiltin(name string, fn func([]float64, []float64, float64) (float64, error)) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data1, data2 floatOrIntList
		var param floatOrInt
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 3, &data1, &data2, &param); err != nil {
			return nil, err
		}
		result, err := fn(data1, data2, float64(param))
		if err != nil {
			return nil, err
		}
		return starlark.Float(result), nil
	})
}
