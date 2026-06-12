// Package regex provides regular expression functions for Starlark, intended
// as a subset of Python's re module backed by Go's RE2 engine.
//
// Python re semantics subset (RE2 engine, no lookaround / no backreferences).
// RE2 guarantees linear-time matching — no catastrophic backtracking / ReDoS —
// which suits running untrusted or LLM-generated scripts in a sandbox. Where
// RE2 genuinely differs from Python (lookahead/lookbehind (?=...)/(?<=...) and
// in-pattern backreferences \1), the pattern fails to compile with a clear
// error instead of silently misbehaving. Shapes that DO align with Python —
// Match objects, findall group tuples, sub's \1 / \g<name> replacement, the
// IGNORECASE/MULTILINE/DOTALL flags — are matched faithfully.
//
// The legacy `re` module is frozen; new code should use this `regex` module.
package regex

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName is the name under which the module is loaded, e.g. load('regex', 'compile').
const ModuleName = "regex"

// Flag values mirror Python's re.I / re.M / re.S so scripts can OR them; they
// are translated to RE2 inline flags ((?i)(?m)(?s)) at compile time.
const (
	flagIgnoreCase = 2
	flagMultiline  = 8
	flagDotAll     = 16
	flagsAll       = flagIgnoreCase | flagMultiline | flagDotAll
)

var (
	once     sync.Once
	regexMod starlark.StringDict
)

// LoadModule loads the regex module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		regexMod = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"compile":     starlark.NewBuiltin(ModuleName+".compile", compile),
					"try_compile": starlark.NewBuiltin(ModuleName+".try_compile", wrapTry(compile)),
					"search":      starlark.NewBuiltin(ModuleName+".search", search),
					"try_search":  starlark.NewBuiltin(ModuleName+".try_search", wrapTry(search)),
					"match":       starlark.NewBuiltin(ModuleName+".match", match),
					"fullmatch":   starlark.NewBuiltin(ModuleName+".fullmatch", fullmatch),
					"findall":     starlark.NewBuiltin(ModuleName+".findall", findall),
					"finditer":    starlark.NewBuiltin(ModuleName+".finditer", finditer),
					"sub":         starlark.NewBuiltin(ModuleName+".sub", sub),
					"subn":        starlark.NewBuiltin(ModuleName+".subn", subn),
					"split":       starlark.NewBuiltin(ModuleName+".split", split),
					"escape":      starlark.NewBuiltin(ModuleName+".escape", escape),
					// flag constants (Python re values, OR-able)
					"I":          starlark.MakeInt(flagIgnoreCase),
					"IGNORECASE": starlark.MakeInt(flagIgnoreCase),
					"M":          starlark.MakeInt(flagMultiline),
					"MULTILINE":  starlark.MakeInt(flagMultiline),
					"S":          starlark.MakeInt(flagDotAll),
					"DOTALL":     starlark.MakeInt(flagDotAll),
				},
			},
		}
	})
	return regexMod, nil
}

// builtinFn is the signature of a module-level builtin.
type builtinFn func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)

// wrapTry converts a builtin into its try_ variant: a (value, error-string)
// pair with a nil Go error, the shape shared by lib/json/csv/http.
func wrapTry(fn builtinFn) builtinFn {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		res, err := fn(thread, b, args, kwargs)
		if err != nil {
			return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
		}
		if res == nil {
			res = starlark.None
		}
		return starlark.Tuple{res, starlark.None}, nil
	}
}

// flagsPrefix builds the RE2 inline-flag prefix for the given Python flags.
func flagsPrefix(flags int) (string, error) {
	if flags&^flagsAll != 0 {
		return "", fmt.Errorf("unsupported flags value %d (only I/IGNORECASE, M/MULTILINE, S/DOTALL)", flags)
	}
	var f string
	if flags&flagIgnoreCase != 0 {
		f += "i"
	}
	if flags&flagMultiline != 0 {
		f += "m"
	}
	if flags&flagDotAll != 0 {
		f += "s"
	}
	if f == "" {
		return "", nil
	}
	return "(?" + f + ")", nil
}

// compileFor compiles a pattern for the given match kind. kind controls
// anchoring: "search" none, "match" anchors at the start (\A), "full" anchors
// both ends (\A...\z). The flags become an inline prefix.
func compileFor(pattern string, flags int, kind string) (*regexp.Regexp, error) {
	prefix, err := flagsPrefix(flags)
	if err != nil {
		return nil, err
	}
	body := pattern
	switch kind {
	case "match":
		body = `\A(?:` + pattern + `)`
	case "full":
		body = `\A(?:` + pattern + `)\z`
	}
	re, err := regexp.Compile(prefix + body)
	if err != nil {
		return nil, fmt.Errorf("cannot compile pattern: %w", err)
	}
	return re, nil
}

// ---- module-level functions ----

func compile(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		pattern string
		flags   int
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "pattern", &pattern, "flags?", &flags); err != nil {
		return nil, err
	}
	return newPattern(pattern, flags)
}

func search(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	p, str, err := unpackPatternCall(b, args, kwargs)
	if err != nil {
		return nil, err
	}
	return p.doFind(str, "search"), nil
}

func match(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	p, str, err := unpackPatternCall(b, args, kwargs)
	if err != nil {
		return nil, err
	}
	return p.doFind(str, "match"), nil
}

func fullmatch(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	p, str, err := unpackPatternCall(b, args, kwargs)
	if err != nil {
		return nil, err
	}
	return p.doFind(str, "full"), nil
}

func findall(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	p, str, err := unpackPatternCall(b, args, kwargs)
	if err != nil {
		return nil, err
	}
	return p.doFindall(str), nil
}

func finditer(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	p, str, err := unpackPatternCall(b, args, kwargs)
	if err != nil {
		return nil, err
	}
	return p.doFinditer(str), nil
}

func split(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		pattern, str string
		maxsplit     int
		flags        int
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "pattern", &pattern, "string", &str, "maxsplit?", &maxsplit, "flags?", &flags); err != nil {
		return nil, err
	}
	p, err := newPattern(pattern, flags)
	if err != nil {
		return nil, err
	}
	return p.doSplit(str, maxsplit), nil
}

func sub(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	pat, repl, str, count, flags, err := unpackSubCall(b, args, kwargs)
	if err != nil {
		return nil, err
	}
	p, err := newPattern(pat, flags)
	if err != nil {
		return nil, err
	}
	out, _, err := p.doSub(thread, repl, str, count)
	if err != nil {
		return nil, err
	}
	return starlark.String(out), nil
}

func subn(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	pat, repl, str, count, flags, err := unpackSubCall(b, args, kwargs)
	if err != nil {
		return nil, err
	}
	p, err := newPattern(pat, flags)
	if err != nil {
		return nil, err
	}
	out, n, err := p.doSub(thread, repl, str, count)
	if err != nil {
		return nil, err
	}
	return starlark.Tuple{starlark.String(out), starlark.MakeInt(n)}, nil
}

func escape(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var s string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "pattern", &s); err != nil {
		return nil, err
	}
	return starlark.String(regexp.QuoteMeta(s)), nil
}

// unpackPatternCall handles the (pattern, string, flags=0) argument form.
func unpackPatternCall(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (*Pattern, string, error) {
	var (
		pattern, str string
		flags        int
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "pattern", &pattern, "string", &str, "flags?", &flags); err != nil {
		return nil, "", err
	}
	p, err := newPattern(pattern, flags)
	if err != nil {
		return nil, "", err
	}
	return p, str, nil
}

// unpackSubCall handles the (pattern, repl, string, count=0, flags=0) form.
func unpackSubCall(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (pat string, repl starlark.Value, str string, count, flags int, err error) {
	err = starlark.UnpackArgs(b.Name(), args, kwargs, "pattern", &pat, "repl", &repl, "string", &str, "count?", &count, "flags?", &flags)
	return
}

// ---- Pattern type ----

// Pattern is a compiled regular expression (the value returned by compile()).
type Pattern struct {
	pattern string
	flags   int
	search  *regexp.Regexp // unanchored
	mat     *regexp.Regexp // \A-anchored
	full    *regexp.Regexp // \A...\z-anchored
}

func newPattern(pattern string, flags int) (*Pattern, error) {
	s, err := compileFor(pattern, flags, "search")
	if err != nil {
		return nil, err
	}
	m, err := compileFor(pattern, flags, "match")
	if err != nil {
		return nil, err
	}
	f, err := compileFor(pattern, flags, "full")
	if err != nil {
		return nil, err
	}
	return &Pattern{pattern: pattern, flags: flags, search: s, mat: m, full: f}, nil
}

func (p *Pattern) reFor(kind string) *regexp.Regexp {
	switch kind {
	case "match":
		return p.mat
	case "full":
		return p.full
	default:
		return p.search
	}
}

// doFind runs a single match and returns a Match value or None.
func (p *Pattern) doFind(str, kind string) starlark.Value {
	re := p.reFor(kind)
	loc := re.FindStringSubmatchIndex(str)
	if loc == nil {
		return starlark.None
	}
	return &Match{p: p, str: str, loc: loc}
}

// doFindall implements Python findall shaping: no group -> the full matches;
// one group -> that group's text; several groups -> a tuple per match.
func (p *Pattern) doFindall(str string) starlark.Value {
	ng := p.search.NumSubexp()
	var out starlark.Tuple
	for _, m := range p.search.FindAllStringSubmatchIndex(str, -1) {
		switch ng {
		case 0:
			out = append(out, starlark.String(str[m[0]:m[1]]))
		case 1:
			out = append(out, groupText(str, m, 1))
		default:
			var grp starlark.Tuple
			for g := 1; g <= ng; g++ {
				grp = append(grp, groupText(str, m, g))
			}
			out = append(out, grp)
		}
	}
	return out
}

func (p *Pattern) doFinditer(str string) starlark.Value {
	var out starlark.Tuple
	for _, m := range p.search.FindAllStringSubmatchIndex(str, -1) {
		loc := make([]int, len(m))
		copy(loc, m)
		out = append(out, &Match{p: p, str: str, loc: loc})
	}
	return out
}

// doSplit implements Python split: the text of capture groups is included in
// the result. A non-positive maxsplit means no limit.
func (p *Pattern) doSplit(str string, maxsplit int) starlark.Value {
	ng := p.search.NumSubexp()
	var out starlark.Tuple
	last := 0
	n := 0
	for _, m := range p.search.FindAllStringSubmatchIndex(str, -1) {
		if maxsplit > 0 && n >= maxsplit {
			break
		}
		out = append(out, starlark.String(str[last:m[0]]))
		for g := 1; g <= ng; g++ {
			out = append(out, groupText(str, m, g))
		}
		last = m[1]
		n++
	}
	out = append(out, starlark.String(str[last:]))
	return out
}

// doSub replaces matches. repl is either a string template (Python \1 /
// \g<name> translated to Go's $-syntax) or a callable invoked per Match.
func (p *Pattern) doSub(thread *starlark.Thread, repl starlark.Value, str string, count int) (string, int, error) {
	var (
		tmpl     string
		callable starlark.Callable
	)
	switch r := repl.(type) {
	case starlark.String:
		tmpl = translateRepl(r.GoString())
	case starlark.Callable:
		callable = r
	default:
		return "", 0, fmt.Errorf("repl must be a string or a function, got %s", repl.Type())
	}

	if count < 0 {
		// Python: a negative count replaces nothing
		return str, 0, nil
	}
	limit := -1
	if count > 0 {
		limit = count
	}

	var b []byte
	last := 0
	n := 0
	for _, m := range p.search.FindAllStringSubmatchIndex(str, limit) {
		b = append(b, str[last:m[0]]...)
		if callable != nil {
			rv, err := starlark.Call(thread, callable, starlark.Tuple{&Match{p: p, str: str, loc: append([]int(nil), m...)}}, nil)
			if err != nil {
				return "", 0, err
			}
			rs, ok := starlark.AsString(rv)
			if !ok {
				return "", 0, fmt.Errorf("repl function must return a string, got %s", rv.Type())
			}
			b = append(b, rs...)
		} else {
			b = p.search.ExpandString(b, tmpl, str, m)
		}
		last = m[1]
		n++
	}
	b = append(b, str[last:]...)
	return string(b), n, nil
}

// translateRepl converts a Python replacement template to Go's ExpandString
// syntax: \1 -> ${1}, \g<name> -> ${name}, and a literal $ is escaped to $$.
func translateRepl(r string) string {
	var b strings.Builder
	rs := []rune(r)
	for i := 0; i < len(rs); i++ {
		switch {
		case rs[i] == '$':
			b.WriteString("$$")
		case rs[i] == '\\' && i+1 < len(rs):
			next := rs[i+1]
			switch {
			case next >= '0' && next <= '9':
				b.WriteString("${" + string(next) + "}")
				i++
			case next == 'g' && i+2 < len(rs) && rs[i+2] == '<':
				j := i + 3
				for j < len(rs) && rs[j] != '>' {
					j++
				}
				if j < len(rs) {
					b.WriteString("${" + string(rs[i+3:j]) + "}")
					i = j
				} else {
					b.WriteRune(rs[i])
				}
			case next == '\\':
				b.WriteByte('\\')
				i++
			case next == 'n':
				b.WriteByte('\n')
				i++
			case next == 't':
				b.WriteByte('\t')
				i++
			case next == 'r':
				b.WriteByte('\r')
				i++
			default:
				b.WriteRune(rs[i])
			}
		default:
			b.WriteRune(rs[i])
		}
	}
	return b.String()
}

// groupText returns the text of capture group g for a submatch index slice,
// or None if the group did not participate.
func groupText(str string, loc []int, g int) starlark.Value {
	if 2*g+1 >= len(loc) || loc[2*g] < 0 {
		return starlark.None
	}
	return starlark.String(str[loc[2*g]:loc[2*g+1]])
}

// starlark.Value implementation for Pattern.

func (p *Pattern) String() string        { return fmt.Sprintf("regex.compile(%q)", p.pattern) }
func (p *Pattern) Type() string          { return "regex.Pattern" }
func (p *Pattern) Freeze()               {}
func (p *Pattern) Truth() starlark.Bool  { return starlark.True }
func (p *Pattern) Hash() (uint32, error) { return hashString(p.pattern), nil }

var patternMethods = map[string]builtinMethod{
	"search": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Pattern).methFind(a, k, "search")
	},
	"match": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Pattern).methFind(a, k, "match")
	},
	"fullmatch": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Pattern).methFind(a, k, "full")
	},
	"findall": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Pattern).methFindall(a, k)
	},
	"finditer": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Pattern).methFinditer(a, k)
	},
	"split": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Pattern).methSplit(a, k)
	},
	"sub": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Pattern).methSub(t, a, k, false)
	},
	"subn": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Pattern).methSub(t, a, k, true)
	},
}

func (p *Pattern) Attr(name string) (starlark.Value, error) {
	switch name {
	case "pattern":
		return starlark.String(p.pattern), nil
	case "flags":
		return starlark.MakeInt(p.flags), nil
	case "groups":
		return starlark.MakeInt(p.search.NumSubexp()), nil
	}
	return builtinMethods(p, name, patternMethods)
}

func (p *Pattern) AttrNames() []string {
	names := builtinAttrNames(patternMethods)
	names = append(names, "pattern", "flags", "groups")
	sort.Strings(names)
	return names
}

func (p *Pattern) methFind(args starlark.Tuple, kwargs []starlark.Tuple, kind string) (starlark.Value, error) {
	var str string
	if err := starlark.UnpackArgs("("+kind+")", args, kwargs, "string", &str); err != nil {
		return nil, err
	}
	return p.doFind(str, kind), nil
}

func (p *Pattern) methFindall(args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var str string
	if err := starlark.UnpackArgs("findall", args, kwargs, "string", &str); err != nil {
		return nil, err
	}
	return p.doFindall(str), nil
}

func (p *Pattern) methFinditer(args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var str string
	if err := starlark.UnpackArgs("finditer", args, kwargs, "string", &str); err != nil {
		return nil, err
	}
	return p.doFinditer(str), nil
}

func (p *Pattern) methSplit(args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		str      string
		maxsplit int
	)
	if err := starlark.UnpackArgs("split", args, kwargs, "string", &str, "maxsplit?", &maxsplit); err != nil {
		return nil, err
	}
	return p.doSplit(str, maxsplit), nil
}

func (p *Pattern) methSub(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple, withN bool) (starlark.Value, error) {
	var (
		repl  starlark.Value
		str   string
		count int
	)
	if err := starlark.UnpackArgs("sub", args, kwargs, "repl", &repl, "string", &str, "count?", &count); err != nil {
		return nil, err
	}
	out, n, err := p.doSub(thread, repl, str, count)
	if err != nil {
		return nil, err
	}
	if withN {
		return starlark.Tuple{starlark.String(out), starlark.MakeInt(n)}, nil
	}
	return starlark.String(out), nil
}

// ---- Match type ----

// Match is the result of a successful search/match/fullmatch.
type Match struct {
	p   *Pattern
	str string
	loc []int // submatch index pairs, as from FindStringSubmatchIndex
}

func (m *Match) String() string        { return fmt.Sprintf("regex.Match(%q)", m.str[m.loc[0]:m.loc[1]]) }
func (m *Match) Type() string          { return "regex.Match" }
func (m *Match) Freeze()               {}
func (m *Match) Truth() starlark.Bool  { return starlark.True }
func (m *Match) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: regex.Match") }

// groupIndex resolves a group selector (int index or name) to a numeric index.
func (m *Match) groupIndex(v starlark.Value) (int, error) {
	switch g := v.(type) {
	case starlark.Int:
		i, _ := g.Int64()
		if i < 0 || int(i) > m.p.search.NumSubexp() {
			return 0, fmt.Errorf("no such group: %d", i)
		}
		return int(i), nil
	case starlark.String:
		for i, name := range m.p.search.SubexpNames() {
			if name == g.GoString() {
				return i, nil
			}
		}
		return 0, fmt.Errorf("no such group: %q", g.GoString())
	default:
		return 0, fmt.Errorf("group index must be an int or string, got %s", v.Type())
	}
}

var matchMethods = map[string]builtinMethod{
	"group": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Match).methGroup(a, k)
	},
	"groups": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Match).methGroups(a, k)
	},
	"groupdict": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Match).methGroupdict(a, k)
	},
	"start": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Match).methPos(a, k, true)
	},
	"end": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Match).methPos(a, k, false)
	},
	"span": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Match).methSpan(a, k)
	},
	"expand": func(t *starlark.Thread, recv starlark.Value, a starlark.Tuple, k []starlark.Tuple) (starlark.Value, error) {
		return recv.(*Match).methExpand(a, k)
	},
}

func (m *Match) Attr(name string) (starlark.Value, error) {
	switch name {
	case "string":
		return starlark.String(m.str), nil
	case "re":
		return m.p, nil
	}
	return builtinMethods(m, name, matchMethods)
}

func (m *Match) AttrNames() []string {
	names := builtinAttrNames(matchMethods)
	names = append(names, "string", "re")
	sort.Strings(names)
	return names
}

// methGroup implements Match.group(*indices): no arg -> group(0); one arg ->
// that group; several -> a tuple of them. Index may be int or name.
func (m *Match) methGroup(args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(kwargs) > 0 {
		return nil, fmt.Errorf("group: unexpected keyword argument")
	}
	if len(args) == 0 {
		return groupText(m.str, m.loc, 0), nil
	}
	if len(args) == 1 {
		gi, err := m.groupIndex(args[0])
		if err != nil {
			return nil, err
		}
		return groupText(m.str, m.loc, gi), nil
	}
	var out starlark.Tuple
	for _, a := range args {
		gi, err := m.groupIndex(a)
		if err != nil {
			return nil, err
		}
		out = append(out, groupText(m.str, m.loc, gi))
	}
	return out, nil
}

// methGroups returns a tuple of all capture groups, with a default for
// non-participating groups (default is None unless given).
func (m *Match) methGroups(args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var def starlark.Value = starlark.None
	if err := starlark.UnpackArgs("groups", args, kwargs, "default?", &def); err != nil {
		return nil, err
	}
	var out starlark.Tuple
	for g := 1; g <= m.p.search.NumSubexp(); g++ {
		v := groupText(m.str, m.loc, g)
		if v == starlark.None {
			v = def
		}
		out = append(out, v)
	}
	return out, nil
}

// methGroupdict returns a dict of named groups to their text.
func (m *Match) methGroupdict(args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var def starlark.Value = starlark.None
	if err := starlark.UnpackArgs("groupdict", args, kwargs, "default?", &def); err != nil {
		return nil, err
	}
	d := starlark.NewDict(0)
	for i, name := range m.p.search.SubexpNames() {
		if name == "" {
			continue
		}
		v := groupText(m.str, m.loc, i)
		if v == starlark.None {
			v = def
		}
		if err := d.SetKey(starlark.String(name), v); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func (m *Match) methPos(args starlark.Tuple, kwargs []starlark.Tuple, start bool) (starlark.Value, error) {
	var gv starlark.Value = starlark.MakeInt(0)
	if err := starlark.UnpackArgs("pos", args, kwargs, "group?", &gv); err != nil {
		return nil, err
	}
	gi, err := m.groupIndex(gv)
	if err != nil {
		return nil, err
	}
	idx := m.loc[2*gi]
	if !start {
		idx = m.loc[2*gi+1]
	}
	return starlark.MakeInt(idx), nil
}

func (m *Match) methSpan(args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var gv starlark.Value = starlark.MakeInt(0)
	if err := starlark.UnpackArgs("span", args, kwargs, "group?", &gv); err != nil {
		return nil, err
	}
	gi, err := m.groupIndex(gv)
	if err != nil {
		return nil, err
	}
	return starlark.Tuple{starlark.MakeInt(m.loc[2*gi]), starlark.MakeInt(m.loc[2*gi+1])}, nil
}

func (m *Match) methExpand(args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var tmpl string
	if err := starlark.UnpackArgs("expand", args, kwargs, "template", &tmpl); err != nil {
		return nil, err
	}
	out := m.p.search.ExpandString(nil, translateRepl(tmpl), m.str, m.loc)
	return starlark.String(out), nil
}

// ---- shared method-dispatch helpers ----

type builtinMethod func(thread *starlark.Thread, recv starlark.Value, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)

func builtinMethods(recv starlark.Value, name string, methods map[string]builtinMethod) (starlark.Value, error) {
	method := methods[name]
	if method == nil {
		return nil, nil // no such method
	}
	impl := func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return method(thread, b.Receiver(), args, kwargs)
	}
	return starlark.NewBuiltin(name, impl).BindReceiver(recv), nil
}

func builtinAttrNames(methods map[string]builtinMethod) []string {
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	return names
}

func hashString(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}
