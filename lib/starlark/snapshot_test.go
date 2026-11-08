package starlark_test

// Sections: unsupported host values; cycles and aliases; data quotas.

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	mod "github.com/1set/starlet/lib/starlark"
	std "go.starlark.net/starlark"
)

type hostileValue struct{ called *bool }

func (v hostileValue) String() string        { *v.called = true; panic("String called") }
func (v hostileValue) Type() string          { *v.called = true; panic("Type called") }
func (v hostileValue) Freeze()               { *v.called = true; panic("Freeze called") }
func (v hostileValue) Truth() std.Bool       { *v.called = true; panic("Truth called") }
func (v hostileValue) Hash() (uint32, error) { *v.called = true; panic("Hash called") }

func TestEvalRejectsUnsupportedInput(t *testing.T) {
	var list *std.List
	var dict *std.Dict
	var set *std.Set
	called := false
	values := []std.Value{nil, list, dict, set, hostileValue{&called}, std.NewBuiltin("callback", nil)}
	for _, v := range values {
		_, err := mod.Eval(&std.Thread{}, "x", std.StringDict{"x": std.NewList([]std.Value{v})})
		var e *mod.Error
		if !errors.As(err, &e) || e.Phase != "input" || !strings.Contains(err.Error(), "x") {
			t.Fatalf("%T: %v", v, err)
		}
	}
	if called {
		t.Fatal("input inspection invoked host behavior")
	}
	for _, key := range []string{"", "1x", "try", "x.y"} {
		if _, err := mod.Eval(&std.Thread{}, "1", std.StringDict{key: std.None}); err == nil {
			t.Fatal("accepted bad name", key)
		}
	}
	vars := std.NewDict(1)
	_ = vars.SetKey(std.MakeInt(1), std.None)
	if _, err := call(t, "eval", "1", vars); err == nil {
		t.Fatal("accepted nonstring variable name")
	}
}

func TestEvalCyclesAliasesAndQuota(t *testing.T) {
	list := std.NewList(nil)
	_ = list.Append(list)
	if _, err := mod.Eval(&std.Thread{}, "x", std.StringDict{"x": list}); err == nil {
		t.Fatal("accepted cycle")
	}
	if err := list.Append(std.None); err != nil {
		t.Fatal("rejected input was frozen")
	}
	shared := std.NewList([]std.Value{std.MakeInt(1)})
	v, err := mod.Eval(&std.Thread{}, "(a, b)", std.StringDict{"a": shared, "b": shared})
	if err != nil {
		t.Fatal(err)
	}
	items := v.(std.Tuple)
	if items[0] != items[1] || items[0] == shared {
		t.Fatal("aliases not privately preserved")
	}
	bigValue := std.MakeBigInt(new(big.Int).Lsh(big.NewInt(1), 4096))
	for _, value := range []std.Value{std.String(strings.Repeat("x", (1<<20)+1)), bigValue, std.NewList(make([]std.Value, 4097))} {
		if _, err := mod.Eval(&std.Thread{}, "x", std.StringDict{"x": value}); err == nil {
			t.Fatalf("accepted oversized %T", value)
		}
	}
	var deep std.Value = std.None
	for i := 0; i < 65; i++ {
		deep = std.NewList([]std.Value{deep})
	}
	if _, err := mod.Eval(&std.Thread{}, "x", std.StringDict{"x": deep}); err == nil {
		t.Fatal("accepted deep graph")
	}
	if _, err := mod.Eval(&std.Thread{}, `"x" * 1048577`, nil); err == nil {
		t.Fatal("accepted oversized output")
	}
}

func TestEvalContainerGraphBoundaries(t *testing.T) {
	dict := std.NewDict(1)
	_ = dict.SetKey(std.String("self"), dict)
	tuple := make(std.Tuple, 1)
	tuple[0] = tuple
	for _, value := range []std.Value{dict, tuple} {
		if _, err := mod.Eval(&std.Thread{}, "x", std.StringDict{"x": value}); err == nil {
			t.Fatal("accepted cycle")
		}
	}
	var shared std.Value = std.None
	for i := 0; i < 40; i++ {
		shared = std.NewList([]std.Value{shared})
	}
	longer := shared
	for i := 0; i < 40; i++ {
		longer = std.NewList([]std.Value{longer})
	}
	if _, err := mod.Eval(&std.Thread{}, "1", std.StringDict{"a": shared, "b": longer}); err == nil {
		t.Fatal("memo bypassed path depth")
	}
	containers := []std.Value{std.Tuple(make([]std.Value, 4097)), std.Bytes(strings.Repeat("x", (1<<20)+1))}
	largeDict := std.NewDict(2050)
	largeSet := std.NewSet(4097)
	for i := 0; i < 4097; i++ {
		_ = largeSet.Insert(std.MakeInt(i))
		if i < 2050 {
			_ = largeDict.SetKey(std.MakeInt(i), std.None)
		}
	}
	containers = append(containers, largeDict, largeSet)
	for _, value := range containers {
		if _, err := mod.Eval(&std.Thread{}, "x", std.StringDict{"x": value}); err == nil {
			t.Fatalf("accepted oversized %T", value)
		}
	}
	for _, src := range []string{`[0] * 5000`, `int("f" * 1025, 16)`, `(lambda x: (x.append(x), x)[1])([])`} {
		_, err := mod.Eval(&std.Thread{}, src, nil)
		var e *mod.Error
		if !errors.As(err, &e) || e.Phase != "output" {
			t.Fatalf("%s: %v", src, err)
		}
	}
	variables := make(std.StringDict, 4097)
	for i := 0; i < 4097; i++ {
		variables[strings.Repeat("a", i+1)] = std.None
	}
	if _, err := mod.Eval(&std.Thread{}, "1", variables); err == nil {
		t.Fatal("accepted too many variables")
	}
	if _, err := mod.Eval(&std.Thread{}, "1", std.StringDict{strings.Repeat("a", (1<<20)+1): std.None}); err == nil || len(err.Error()) > 5000 {
		t.Fatal("unbounded variable name rejection")
	}
	if _, err := mod.Eval(&std.Thread{}, "1", std.StringDict{strings.Repeat("a", 513): std.None}); err == nil || len(err.Error()) > 5000 {
		t.Fatal("unbounded invalid identifier")
	}
	// Host mutation of a tuple after insertion can invalidate a native key.
	// Reject the malformed graph rather than panic or silently lose the item.
	for _, kind := range []string{"dict", "set"} {
		key := std.Tuple{std.MakeInt(1)}
		var value std.Value
		if kind == "dict" {
			d := std.NewDict(1)
			_ = d.SetKey(key, std.None)
			value = d
		} else {
			s := std.NewSet(1)
			_ = s.Insert(key)
			value = s
		}
		key[0] = std.NewList(nil)
		if _, err := mod.Eval(&std.Thread{}, "x", std.StringDict{"x": value}); err == nil {
			t.Fatal("accepted malformed", kind)
		}
	}
	called := false
	for _, kind := range []string{"dict", "set"} {
		key := std.Tuple{std.MakeInt(1)}
		var value std.Value
		if kind == "dict" {
			d := std.NewDict(1)
			_ = d.SetKey(key, std.None)
			value = d
		} else {
			s := std.NewSet(1)
			_ = s.Insert(key)
			value = s
		}
		key[0] = hostileValue{&called}
		if _, err := mod.Eval(&std.Thread{}, "x", std.StringDict{"x": value}); err == nil {
			t.Fatal("accepted invalid native key", kind)
		}
	}
	if called {
		t.Fatal("key validation executed host behavior")
	}
	inside := std.NewDict(1)
	_ = inside.SetKey(std.String("tuple"), std.Tuple{std.MakeInt(1)})
	if _, err := mod.Eval(&std.Thread{}, "x", std.StringDict{"x": std.NewList([]std.Value{inside})}); err != nil {
		t.Fatal(err)
	}
	set := std.NewSet(1)
	_ = set.Insert(std.MakeInt(1))
	if _, err := mod.Eval(&std.Thread{}, "(a,b)", std.StringDict{"a": set, "b": set}); err != nil {
		t.Fatal(err)
	}
	if _, err := mod.Eval(&std.Thread{}, "1", std.StringDict{"a": std.String(strings.Repeat("x", (1<<20)-1)), "b": std.None}); err == nil {
		t.Fatal("ignored names in total bytes")
	}
	full := make([]std.Value, 4094)
	for i := range full {
		full[i] = std.None
	}
	if _, err := mod.Eval(&std.Thread{}, "x", std.StringDict{"x": std.Tuple{std.NewList(full), std.None}}); err == nil {
		t.Fatal("ignored later graph edge")
	}
}
