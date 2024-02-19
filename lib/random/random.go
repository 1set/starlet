// Package random defines functions that generate random values for various distributions, it's intended to be a drop-in subset of Python's random module for Starlark.
package random

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"

	itn "github.com/1set/starlet/internal"
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
					"randbytes": starlark.NewBuiltin("random.randbytes", randbytes),
					"randint":   starlark.NewBuiltin("random.randint", randint),
					"randstr":   starlark.NewBuiltin("random.randstr", randstr),
					"choice":    starlark.NewBuiltin("random.choice", choice),
					"shuffle":   starlark.NewBuiltin("random.shuffle", shuffle),
					"random":    starlark.NewBuiltin("random.random", random),
					"uniform":   starlark.NewBuiltin("random.uniform", uniform),
				},
			},
		}
	})
	return module, nil
}

// for convenience
var (
	emptyStr string
	none     = starlark.None
)

// randbytes(n) returns a random byte string of length n.
func randbytes(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// precondition checks
	var n starlark.Int
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "n?", &n); err != nil {
		return nil, err
	}
	// set default value if n is not provided correctly
	nInt := n.BigInt()
	if nInt.Sign() <= 0 {
		nInt = big.NewInt(10)
	}
	// get random bytes
	buf := make([]byte, nInt.Int64())
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return starlark.Bytes(buf), nil
}

// randint(a, b) returns a random integer N such that a <= N <= b. Alias for randrange(a, b+1).
func randint(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// precondition checks
	var a, b starlark.Int
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "a", &a, "b", &b); err != nil {
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
func choice(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// precondition checks
	var seq starlark.Indexable
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "seq", &seq); err != nil {
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
func shuffle(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// precondition checks
	var seq starlark.HasSetIndex
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "seq", &seq); err != nil {
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

// random() returns a random floating point number in the range 0.0 <= X < 1.0.
func random(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	f, err := getRandomFloat(1 << 53)
	if err != nil {
		return starlark.None, err
	}
	return starlark.Float(f), nil
}

// uniform(a, b) returns a random floating point number N such that a <= N <= b for a <= b and b <= N <= a for b < a. The end-point value b may or may not be included in the range depending on floating-point rounding in the equation a + (b-a) * random().
func uniform(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// precondition checks
	var a, b itn.FloatOrInt
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "a", &a, "b", &b); err != nil {
		return starlark.None, err
	}
	// get random float
	f, err := getRandomFloat(1 << 53)
	if err != nil {
		return starlark.None, err
	}
	// a + (b-a) * random()
	diff := float64(b - a)
	return starlark.Float(float64(a) + diff*f), nil
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

// getRandomFloat returns a random floating point number in the range [0.0, 1.0).
func getRandomFloat(prec int64) (n float64, err error) {
	if prec <= 0 {
		return 0, errors.New(`prec must be > 0`)
	}
	maxBig := new(big.Int).SetUint64(uint64(prec))
	nBig, err := rand.Int(rand.Reader, maxBig)
	if err != nil {
		return 0, err
	}
	return float64(nBig.Int64()) / float64(prec), nil
}

// getRandStr returns a random string of given length from given characters.
func getRandStr(alphabet string, length int64) (string, error) {
	// basic checks
	if length <= 0 {
		return emptyStr, errors.New(`length must be > 0`)
	}
	if alphabet == emptyStr {
		return emptyStr, errors.New(`alphabet must not be empty`)
	}

	// split alphabet into runes
	runes := []rune(alphabet)
	rc := len(runes)

	// get random runes
	buf := make([]rune, length)
	for i := range buf {
		idx, err := getRandomInt(rc)
		if err != nil {
			return emptyStr, err
		}
		buf[i] = runes[idx]
	}

	// convert to string
	return string(buf), nil
}

// randstr(a, l) returns a random string of given length from given characters.
func randstr(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// precondition checks
	var (
		ab starlark.String
		l  starlark.Int
	)
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "alphabet", &ab, "len", &l); err != nil {
		return nil, err
	}
	// get random strings
	li := l.BigInt()
	s, err := getRandStr(ab.GoString(), li.Int64())
	if err != nil {
		return none, err
	}
	return starlark.String(s), nil
}
