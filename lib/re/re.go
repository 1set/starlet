// Package re defines regular expression functions, it's intended to be a drop-in subset of Python's re module for Starlark.
//
// Migrated from: https://github.com/qri-io/starlib/tree/master/re
// TODO:
// 1) fullmatch, finditer, subn, escape, search
// 2) Match as a type
// 3) Support flags as constants
package re

import (
	"fmt"
	"regexp"
	"sort"
	"sync"

	itn "github.com/1set/starlet/dataconv"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used in Starlark's load() function, eg: load('re', 'match')
const ModuleName = "re"

var (
	once     sync.Once
	reModule starlark.StringDict
)

// LoadModule loads the re module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		reModule = starlark.StringDict{
			"re": &starlarkstruct.Module{
				Name: "re",
				Members: starlark.StringDict{
					"compile": starlark.NewBuiltin("re.compile", compile),
					"search":  starlark.NewBuiltin("re.search", search),
					"match":   starlark.NewBuiltin("re.match", match),
					// "fullmatch": starlark.NewBuiltin("re.fullmatch", fullmatch),
					"split":   starlark.NewBuiltin("re.split", split),
					"findall": starlark.NewBuiltin("re.findall", findall),
					// "finditer":  starlark.NewBuiltin("re.finditer", finditer),
					"sub": starlark.NewBuiltin("re.sub", sub),
					// "subn":      starlark.NewBuiltin("re.subn", subn),
					// "escape":    starlark.NewBuiltin("re.escape", escape),
				},
			},
		}
	})
	return reModule, nil
}

// compile(pattern, flags=0)
// Compile a regular expression pattern into a regular expression object, which
// can be used for matching using its match(), search() and other methods.
//
// flags must be 0: numeric re.* flags are not supported by this module and
// are rejected loudly; use inline pattern flags like (?i), (?m), (?s) instead.
func compile(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		pattern starlark.String
		flags   starlark.Int
	)

	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "pattern", &pattern, "flags?", &flags); err != nil {
		return starlark.None, err
	}
	if err := checkFlags(bn.Name(), flags); err != nil {
		return starlark.None, err
	}

	return newRegex(pattern)
}

// search(pattern,string,flags=0)
// Scan through string looking for the first location where the regular
// expression pattern produces a match, and return the [start, end] byte-index
// pair of that match as a list. Return None if no position in the string
// matches the pattern. flags must be 0 (see compile).
func search(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		pattern, str starlark.String
		flags        starlark.Int
	)
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "pattern", &pattern, "string", &str, "flags?", &flags); err != nil {
		return starlark.None, err
	}
	if err := checkFlags(bn.Name(), flags); err != nil {
		return starlark.None, err
	}
	re, err := newGoRegex(pattern)
	if err != nil {
		return starlark.None, err
	}

	return reSearch(re, str, flags)
}

func reSearch(re *regexp.Regexp, str starlark.String, flags starlark.Int) (starlark.Value, error) {
	matches := re.FindStringIndex(string(str))
	if len(matches) == 0 {
		return starlark.None, nil
	}

	vals := make([]interface{}, len(matches))
	for i, m := range matches {
		vals[i] = m
	}

	return itn.Marshal(vals)
}

// match(pattern, string, flags=0)
// If zero or more characters at the beginning of string match the regular
// expression pattern, return a list with a single tuple holding the full
// match followed by the text of every capture group. Return an empty list
// if the beginning of the string does not match the pattern (a match
// elsewhere in the string does not count; use search() or findall() for
// that). flags must be 0 (see compile).
func match(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		pattern, str starlark.String
		flags        starlark.Int
	)
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "pattern", &pattern, "string", &str, "flags?", &flags); err != nil {
		return starlark.None, err
	}
	if err := checkFlags(bn.Name(), flags); err != nil {
		return starlark.None, err
	}

	re, err := newGoRegex(pattern)
	if err != nil {
		return starlark.None, err
	}

	return reMatch(re, str, flags)
}

func reMatch(re *regexp.Regexp, str starlark.String, flags starlark.Int) (starlark.Value, error) {
	// match() is documented to match only at the beginning of the string
	// (Python semantics), so anchor the pattern. The historical
	// implementation scanned the whole string like findall, which inverted
	// the truthiness of `if match(...)` checks ported from Python.
	are, err := regexp.Compile(`\A(?:` + re.String() + `)`)
	if err != nil {
		return starlark.None, err
	}
	vals := starlark.NewList(nil)
	for _, match := range are.FindAllStringSubmatch(string(str), 1) {
		if err := vals.Append(slStrSlice(match)); err != nil {
			return starlark.None, err
		}
	}
	return vals, nil
}

// fullmatch(pattern, string, flags=0)¶
// If the whole string matches the regular expression pattern, return a corresponding match object.
// Return None if the string does not match the pattern; note that this is different from a zero-length match.
// func fullmatch(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
// 	var pattern starlark.String
// 	if err := starlark.UnpackArgs("fullmatch", args, kwargs, "pattern", &pattern); err != nil {
// 		return starlark.None, err
// 	}

// 	return starlark.None, nil
// }

// split(pattern, string, maxsplit=0, flags=0)
// Split string by the occurrences of pattern. If maxsplit is positive, at
// most maxsplit splits occur, and the remainder of the string is returned
// as the final element; a negative maxsplit means no splits happen at all.
// Note that unlike Python, the text of capture groups in the pattern is
// NOT included in the result. flags must be 0 (see compile).
func split(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		pattern, str    starlark.String
		maxSplit, flags starlark.Int
	)
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "pattern", &pattern, "string", &str, "maxsplit?", &maxSplit, "flags?", &flags); err != nil {
		return starlark.None, err
	}
	if err := checkFlags(bn.Name(), flags); err != nil {
		return starlark.None, err
	}

	re, err := newGoRegex(pattern)
	if err != nil {
		return starlark.None, err
	}

	return reSplit(re, str, maxSplit, flags)
}

func reSplit(re *regexp.Regexp, str starlark.String, maxSplit, flags starlark.Int) (starlark.Value, error) {
	ms, ok := maxSplit.Int64()
	if !ok || ms > 1<<30 {
		ms = 0 // an astronomically large maxsplit behaves like "no limit"
	}
	switch {
	case ms < 0:
		// Python semantics: a negative maxsplit means no splits at all
		return slStrSlice([]string{string(str)}), nil
	case ms == 0:
		// 0 means "split everywhere"; -1 is the sentinel for "all" in Go
		return slStrSlice(re.Split(string(str), -1)), nil
	default:
		// Go's n counts resulting substrings, Python's maxsplit counts
		// splits, so n = maxsplit + 1
		return slStrSlice(re.Split(string(str), int(ms)+1)), nil
	}
}

// findall(pattern, string, flags=0)
// Returns all non-overlapping matches of pattern in string, as a tuple of
// strings. The string is scanned left-to-right, and matches are returned in
// the order found. If one group is present in the pattern, the group text is
// returned instead of the full match; with several groups each element is a
// tuple of the group texts. Empty matches are included in the result.
// flags must be 0 (see compile).
func findall(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		pattern starlark.String
		str     starlark.String
		flags   starlark.Int
	)
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "pattern", &pattern, "string", &str, "flags?", &flags); err != nil {
		return starlark.None, err
	}
	if err := checkFlags(bn.Name(), flags); err != nil {
		return starlark.None, err
	}

	re, err := newGoRegex(pattern)
	if err != nil {
		return starlark.None, err
	}
	return reFindall(re, str, flags)
}

func reFindall(re *regexp.Regexp, str starlark.String, flags starlark.Int) (starlark.Value, error) {
	// Python shaping: no capture groups -> the full match text; one group ->
	// that group's text; several groups -> a tuple of the group texts.
	// The historical implementation always returned the full match and
	// silently dropped the groups its own docstring promised.
	ng := re.NumSubexp()
	var vals starlark.Tuple
	for _, m := range re.FindAllStringSubmatch(string(str), -1) {
		switch ng {
		case 0:
			vals = append(vals, starlark.String(m[0]))
		case 1:
			vals = append(vals, starlark.String(m[1]))
		default:
			vals = append(vals, slStrSlice(m[1:]))
		}
	}
	return vals, nil
}

// func finditer(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
// 	var pattern starlark.String
// 	if err := starlark.UnpackArgs("finditer", args, kwargs, "pattern", &pattern); err != nil {
// 		return starlark.None, err
// 	}

// 	return starlark.None, nil
// }

// sub(pattern, repl, string, count=0, flags=0)
// Return the string obtained by replacing the leftmost non-overlapping
// occurrences of pattern in string by the replacement repl. If the pattern
// isn't found, string is returned unchanged. count limits the number of
// replacements: 0 (the default) replaces all occurrences, and a negative
// count replaces nothing. repl must be a string and uses Go's template
// syntax: $1 or ${name} refers to a capture group and $$ is a literal
// dollar sign — Python backslash references like \1 are NOT interpreted,
// and a function repl is not supported. flags must be 0 (see compile).
func sub(thread *starlark.Thread, bn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		pattern, repl, str starlark.String
		count, flags       starlark.Int
	)
	if err := starlark.UnpackArgs(bn.Name(), args, kwargs, "pattern", &pattern, "repl", &repl, "string", &str, "count?", &count, "flags?", &flags); err != nil {
		return starlark.None, err
	}
	if err := checkFlags(bn.Name(), flags); err != nil {
		return starlark.None, err
	}

	re, err := newGoRegex(pattern)
	if err != nil {
		// the historical code swallowed this error and returned None, which
		// let an invalid dynamic pattern propagate silently as a None value
		return starlark.None, err
	}

	return reSub(re, repl, str, count, flags)
}

func reSub(re *regexp.Regexp, repl, str starlark.String, count, flags starlark.Int) (starlark.Value, error) {
	s := string(str)
	n, ok := count.Int64()
	if !ok {
		n = 0 // an astronomically large count behaves like "replace all"
	}
	switch {
	case n < 0:
		// Python semantics: a negative count means no replacements
		return str, nil
	case n == 0:
		// 0 means "replace all occurrences"
		return starlark.String(re.ReplaceAllString(s, string(repl))), nil
	default:
		// replace only the first n matches; ExpandString shares the same
		// $-template engine as ReplaceAllString, so replaced segments are
		// byte-identical to the replace-all path
		var b []byte
		last := 0
		for _, m := range re.FindAllStringSubmatchIndex(s, int(n)) {
			b = append(b, s[last:m[0]]...)
			b = re.ExpandString(b, string(repl), s, m)
			last = m[1]
		}
		b = append(b, s[last:]...)
		return starlark.String(b), nil
	}
}

// subn(pattern, repl, string, count=0, flags=0)
// Perform the same operation as sub(), but return a tuple (new_string, number_of_subs_made)
// func subn(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
// 	var pattern starlark.String
// 	if err := starlark.UnpackArgs("subn", args, kwargs, "pattern", &pattern); err != nil {
// 		return starlark.None, err
// 	}

// 	return starlark.None, nil
// }

// func escape(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
// 	var pattern starlark.String
// 	if err := starlark.UnpackArgs("escape", args, kwargs, "pattern", &pattern); err != nil {
// 		return starlark.None, err
// 	}

// 	return starlark.None, nil
// }

func newGoRegex(pattern starlark.String) (*regexp.Regexp, error) {
	return regexp.Compile(string(pattern))
}

// checkFlags rejects any non-zero flags value. The flags parameter was
// historically accepted and silently ignored, so Python-style numeric flags
// (e.g. re.IGNORECASE == 2) appeared to work while matching with default
// behaviour — a silent-wrong outcome. The module exports no flag constants,
// so any non-zero value is a porting mistake; fail loudly and point at the
// supported alternative.
func checkFlags(name string, flags starlark.Int) error {
	if flags.Sign() != 0 {
		return fmt.Errorf("%s: flags are not supported by this module, use inline flags like (?i) in the pattern instead", name)
	}
	return nil
}

func slStrSlice(strs []string) starlark.Tuple {
	var vals starlark.Tuple
	for _, s := range strs {
		vals = append(vals, starlark.String(s))
	}
	return vals
}

// hashString computes the FNV hash of s.
func hashString(s string) uint32 {
	var h uint32
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

// Regex is a starlark representation of a compiled regular expression
type Regex struct {
	re *regexp.Regexp
}

func newRegex(pattern starlark.String) (*Regex, error) {
	re, err := newGoRegex(pattern)
	if err != nil {
		return nil, err
	}
	return &Regex{re: re}, nil
}

// String implements the Stringer interface
func (r *Regex) String() string { return r.re.String() }

// Type returns a short string describing the value's type.
func (r *Regex) Type() string { return "regexp" }

// Freeze renders time immutable. required by starlark.Value interface.
// The interface regex presents to the starlark runtime renders it immutable,
// making this a no-op
func (r *Regex) Freeze() {}

// Hash returns a function of x such that Equals(x, y) => Hash(x) == Hash(y)
// required by starlark.Value interface
func (r *Regex) Hash() (uint32, error) { return hashString(r.re.String()), nil }

// Truth returns the truth value of an object required by starlark.Value
// interface. Any non-empty regexp is considered truthy
func (r *Regex) Truth() starlark.Bool { return r.String() != "" }

// Attr gets a value for a string attribute, implementing dot expression support
// in starklark. required by starlark.HasAttrs interface
func (r *Regex) Attr(name string) (starlark.Value, error) {
	return builtinMethods(r, name, regexMethods)
}

var regexMethods = map[string]builtinMethod{
	"search":  compiledSearch,
	"match":   compiledMatch,
	"split":   compiledSplit,
	"findall": compiledFindall,
	"sub":     compiledSub,
}

// AttrNames lists available dot expression strings for time. required by
// starlark.HasAttrs interface
func (r *Regex) AttrNames() []string { return builtinAttrNames(regexMethods) }

func builtinAttrNames(methods map[string]builtinMethod) []string {
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func compiledSearch(fnname string, recV starlark.Value, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		str   starlark.String
		flags starlark.Int
	)
	if err := starlark.UnpackArgs("search", args, kwargs, "string", &str, "flags?", &flags); err != nil {
		return starlark.None, err
	}
	if err := checkFlags(fnname, flags); err != nil {
		return starlark.None, err
	}

	r := recV.(*Regex)
	return reSearch(r.re, str, flags)
}

func compiledMatch(fnname string, recV starlark.Value, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		str   starlark.String
		flags starlark.Int
	)
	if err := starlark.UnpackArgs("match", args, kwargs, "string", &str, "flags?", &flags); err != nil {
		return starlark.None, err
	}
	if err := checkFlags(fnname, flags); err != nil {
		return starlark.None, err
	}

	r := recV.(*Regex)
	return reMatch(r.re, str, flags)
}

func compiledSplit(fnname string, recV starlark.Value, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		str             starlark.String
		maxSplit, flags starlark.Int
	)
	if err := starlark.UnpackArgs("split", args, kwargs, "string", &str, "maxsplit?", &maxSplit, "flags?", &flags); err != nil {
		return starlark.None, err
	}
	if err := checkFlags(fnname, flags); err != nil {
		return starlark.None, err
	}

	r := recV.(*Regex)
	return reSplit(r.re, str, maxSplit, flags)
}

func compiledFindall(fnname string, recV starlark.Value, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		str   starlark.String
		flags starlark.Int
	)
	if err := starlark.UnpackArgs("findall", args, kwargs, "string", &str, "flags?", &flags); err != nil {
		return starlark.None, err
	}
	if err := checkFlags(fnname, flags); err != nil {
		return starlark.None, err
	}

	r := recV.(*Regex)
	return reFindall(r.re, str, flags)
}

func compiledSub(fnname string, recV starlark.Value, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		repl, str    starlark.String
		count, flags starlark.Int
	)
	if err := starlark.UnpackArgs("sub", args, kwargs, "repl", &repl, "string", &str, "count?", &count, "flags?", &flags); err != nil {
		return starlark.None, err
	}
	if err := checkFlags(fnname, flags); err != nil {
		return starlark.None, err
	}

	r := recV.(*Regex)
	return reSub(r.re, repl, str, count, flags)
}

type builtinMethod func(fnname string, recv starlark.Value, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)

func builtinMethods(recv starlark.Value, name string, methods map[string]builtinMethod) (starlark.Value, error) {
	method := methods[name]
	if method == nil {
		return nil, nil // no such method
	}

	// Allocate a closure over 'method'.
	impl := func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return method(b.Name(), b.Receiver(), args, kwargs)
	}
	return starlark.NewBuiltin(name, impl).BindReceiver(recv), nil
}
