// Package hashlib defines hash primitives for Starlark.
//
// Migrated from: https://github.com/qri-io/starlib/tree/master/hash
package hashlib

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"io"
	"sync"

	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used
// in starlark's load() function, eg: load('hashlib', 'md5')
const ModuleName = "hashlib"

var (
	once       sync.Once
	hashModule starlark.StringDict
	hashError  error
)

// LoadModule loads the hashlib module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		hashModule = starlark.StringDict{
			"hashlib": &starlarkstruct.Module{
				Name: "hashlib",
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
		var sb itn.StringOrBytes
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
