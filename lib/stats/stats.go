// Package stats provides a Starlark module for comprehensive statistics functions. It's a wrapper around the Go package: https://github.com/montanaflynn/stats
package stats

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
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

// newUnaryFloatBuiltin wraps a unary function accepting []float64 and returning (float64, error) as a Starlark built-in.
func newUnaryFloatBuiltin(name string, fn func([]float64) (float64, error)) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data tps.FloatOrIntList
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 1, &data); err != nil {
			return nil, err
		}
		result, err := fn([]float64(data.GoSlice()))
		if err != nil {
			return nil, err
		}
		return starlark.Float(result), nil
	})
}

// newBinaryFloatBuiltin wraps a binary function accepting two []float64 arguments and returning (float64, error) as a Starlark built-in.
func newBinaryFloatBuiltin(name string, fn func([]float64, []float64) (float64, error)) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data1, data2 tps.FloatOrIntList
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 2, &data1, &data2); err != nil {
			return nil, err
		}
		result, err := fn([]float64(data1.GoSlice()), []float64(data2.GoSlice()))
		if err != nil {
			return nil, err
		}
		return starlark.Float(result), nil
	})
}

// newBinaryFloatToFloatBuiltin wraps a binary function accepting two float64 arguments and returning float64 as a Starlark built-in.
func newBinaryFloatToFloatBuiltin(name string, fn func(float64, float64) float64) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var x, y starlark.Float
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 2, &x, &y); err != nil {
			return nil, err
		}
		result := fn(float64(x), float64(y))
		return starlark.Float(result), nil
	})
}
