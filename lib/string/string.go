// Package string defines functions that manipulate strings, it's intended to be a drop-in subset of Python's string module for Starlark.
// See https://docs.python.org/3/library/string.html and https://github.com/python/cpython/blob/main/Lib/string.py for reference.
package string

import (
	"fmt"
	"sync"
	"unicode/utf8"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used
// in starlark's load() function, eg: load('string', 'length')
const ModuleName = "string"

var (
	once      sync.Once
	strModule starlark.StringDict
)

const (
	whitespace     = ` \t\n\r\v\f`
	asciiLowerCase = `abcdefghijklmnopqrstuvwxyz`
	asciiUpperCase = `ABCDEFGHIJKLMNOPQRSTUVWXYZ`
	asciiLetters   = asciiLowerCase + asciiUpperCase
	decDigits      = `0123456789`
	hexDigits      = decDigits + `abcdefABCDEF`
	octDigits      = `01234567`
	punctuation    = `!"#$%&'()*+,-./:;<=>?@[\]^_` + "{|}~`"
	printable      = decDigits + asciiLetters + punctuation + whitespace
)

// LoadModule loads the time module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		strModule = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					// constants
					"ascii_lowercase": starlark.String(asciiLowerCase),
					"ascii_uppercase": starlark.String(asciiUpperCase),
					"ascii_letters":   starlark.String(asciiLetters),
					"digits":          starlark.String(decDigits),
					"hexdigits":       starlark.String(hexDigits),
					"octdigits":       starlark.String(octDigits),
					"punctuation":     starlark.String(punctuation),
					"whitespace":      starlark.String(whitespace),
					"printable":       starlark.String(printable),

					// functions
					"length": starlark.NewBuiltin(ModuleName+".length", length),
				},
			},
		}
	})
	return strModule, nil
}

// for convenience
var (
	emptyStr string
	none     = starlark.None
)

// length returns the length of the given value.
func length(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if l := len(args); l != 1 {
		return none, fmt.Errorf(`length() takes exactly one argument (%d given)`, l)
	}

	switch r := args[0]; v := r.(type) {
	case starlark.String:
		return starlark.MakeInt(utf8.RuneCountInString(v.GoString())), nil
	case starlark.Bytes:
		return starlark.MakeInt(len(v)), nil
	default:
		if sv, ok := v.(starlark.Sequence); ok {
			return starlark.MakeInt(sv.Len()), nil
		}
		return none, fmt.Errorf(`object of type '%s' has no length()`, v.Type())
	}
}
