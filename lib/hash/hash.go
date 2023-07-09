// Package hash defines hash primitives for Starlark.
//
// Migrated from: https://github.com/qri-io/starlib/tree/master/hash
package hash

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

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used
// in starlark's load() function, eg: load('hash', 'md5')
const ModuleName = "hash"

var (
	once       sync.Once
	hashModule starlark.StringDict
	hashError  error
)

// LoadModule loads the time module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		hashModule = starlark.StringDict{
			"hash": &starlarkstruct.Module{
				Name: "hash",
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
		if !(len(args) == 1 && len(kwargs) == 0) {
			return starlark.None, fmt.Errorf("%s takes exactly 1 argument", fn.Name())
		}

		// convert arg to string
		var s string
		switch v := args[0].(type) {
		case starlark.String:
			s = v.GoString()
		case starlark.Bytes:
			s = string(v)
		default:
			return starlark.None, fmt.Errorf("%s takes a string or bytes argument", fn.Name())
		}

		// get hash
		h := algo()
		_, err := io.WriteString(h, s)
		if err != nil {
			return starlark.None, err
		}
		return starlark.String(hex.EncodeToString(h.Sum(nil))), nil
	}
}
