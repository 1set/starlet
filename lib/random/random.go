// Package random defines functions that generate random values for various distributions, it's intended to be a drop-in subset of Python's random module for Starlark.
package random

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used in Starlark's load() function, eg: load('random', 'choice')
const ModuleName = "random"

var (
	once   sync.Once
	module starlark.StringDict
)

// LoadModule loads the random module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		module = starlark.StringDict{
			"random": &starlarkstruct.Module{
				Name: "random",
				Members: starlark.StringDict{
					"randint": starlark.NewBuiltin("randint", randint),
					"choice":  starlark.NewBuiltin("choice", choice),
					"shuffle": starlark.NewBuiltin("shuffle", shuffle),
				},
			},
		}
	})
	return module, nil
}

// randint(a, b) returns a random integer N such that a <= N <= b. Alias for randrange(a, b+1).
func randint(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// precondition checks
	var a, b starlark.Int
	if err := starlark.UnpackArgs("randint", args, kwargs, "a", &a, "b", &b); err != nil {
		return nil, err
	}
	// a <= b, then a - b <= 0
	if cmp := a.Sub(b).BigInt(); cmp.Sign() > 0 {
		return nil, errors.New(`a must be less than or equal to b`)
	}
	// get random diff
	var (
		aInt = a.BigInt()
		bInt = b.BigInt()
	)
	diff := new(big.Int).Sub(bInt, aInt)
	diff.Add(diff, big.NewInt(1)) // make it inclusive
	n, err := rand.Int(rand.Reader, diff)
	if err != nil {
		return nil, err
	}
	// rand big int is low + diff
	n.Add(n, aInt)
	return starlark.MakeBigInt(n), nil
}

// choice returns a random element from the non-empty sequence seq. If seq is empty, raises a ValueError.
func choice(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// precondition checks
	var seq starlark.Indexable
	if err := starlark.UnpackArgs("choice", args, kwargs, "seq", &seq); err != nil {
		return starlark.None, err
	}
	l := seq.Len()
	if l == 0 {
		return starlark.None, errors.New(`cannot choose from an empty sequence`)
	}
	// get random index
	i, err := getRandomInt(l)
	if err != nil {
		return starlark.None, err
	}
	// return element at index
	return seq.Index(i), nil
}

// shuffle(x) shuffles the sequence x in place.
func shuffle(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// precondition checks
	var seq starlark.HasSetIndex
	if err := starlark.UnpackArgs("shuffle", args, kwargs, "seq", &seq); err != nil {
		return starlark.None, err
	}
	// nothing to do if seq is empty or has only one element
	l := seq.Len()
	if l <= 1 {
		return starlark.None, nil
	}
	// The shuffle algorithm is the Fisher-Yates Shuffle and its complexity is O(n).
	var (
		randBig   = new(big.Int)
		randBytes = make([]byte, 8)
		swap      = func(i, j int) error {
			x := seq.Index(i)
			y := seq.Index(j)
			if e := seq.SetIndex(i, y); e != nil {
				return e
			}
			if e := seq.SetIndex(j, x); e != nil {
				return e
			}
			return nil
		}
	)
	for i := uint64(l - 1); i > 0; {
		if _, err := rand.Read(randBytes); err != nil {
			return starlark.None, err
		}
		randBig.SetBytes(randBytes)
		for num := randBig.Uint64(); num > i && i > 0; i-- {
			max := i + 1
			j := int(num % max)
			num /= max
			if e := swap(int(i), j); e != nil {
				return starlark.None, e
			}
		}
	}
	// done
	return starlark.None, nil
}

// getRandomInt returns a random integer in the range [0, max).
func getRandomInt(max int) (int, error) {
	if max <= 0 {
		return 0, errors.New(`max must be > 0`)
	}
	maxBig := new(big.Int).SetUint64(uint64(max))
	n, err := rand.Int(rand.Reader, maxBig)
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}