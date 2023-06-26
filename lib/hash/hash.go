/*Package hash defines hash primitives for starlark.

Migrated from: https://github.com/qri-io/starlib/tree/master/hash

  outline: hash
    hash defines hash primitives for starlark.
    path: hash
    functions:
      md5(string) string
        returns an md5 hash for a string
        examples:
          basic
            calculate an md5 checksum for "hello world"
            code:
              load("hash.star", "hash")
              sum = hash.md5("hello world!")
              print(sum)
              # Output: fc3ff98e8c6a0d3087d515c0473f8677
      sha1(string) string
        returns a SHA1 hash for a string
        examples:
          basic
            calculate an SHA1 checksum for "hello world"
            code:
              load("hash.star", "hash")
              sum = hash.sha1("hello world!")
              print(sum)
              # Output: 430ce34d020724ed75a196dfc2ad67c77772d169
      sha256(string) string
        returns an SHA2-256 hash for a string
        examples:
          basic
            calculate an SHA2-256 checksum for "hello world"
            code:
              load("hash.star", "hash")
              sum = hash.sha256("hello world!")
              print(sum)
              # Output: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9
*/
package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
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
		var s starlark.String
		if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &s); err != nil {
			return nil, err
		}

		h := algo()
		_, err := io.WriteString(h, s.GoString())
		if err != nil {
			return starlark.None, err
		}
		return starlark.String(hex.EncodeToString(h.Sum(nil))), nil
	}
}
