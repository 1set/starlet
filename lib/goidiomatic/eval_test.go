package goidiomatic_test

// Sections: data isolation; result boundaries; errors; caller execution state;
// evaluation benchmarks.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	mod "github.com/1set/starlet/lib/goidiomatic"
	std "go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func builtin(t *testing.T, name string) std.Value {
	t.Helper()
	m, err := mod.LoadModule()
	if err != nil {
		t.Fatal(err)
	}
	v := m[name]
	if v == nil {
		t.Fatalf("missing flat builtin %q", name)
	}
	return v
}

func call(t *testing.T, name, source string, vars std.Value) (std.Value, error) {
	t.Helper()
	return std.Call(&std.Thread{}, builtin(t, name), std.Tuple{std.String(source), vars}, nil)
}

func TestEvalReadOnlyData(t *testing.T) {
	items := std.NewList([]std.Value{std.MakeInt(1), std.MakeInt(2)})
	vars := std.StringDict{"items": items, "name": std.String("Ada")}
	v, err := mod.Eval(&std.Thread{}, `{"name": name.upper(), "values": [n * 2 for n in items]}`, vars)
	if err != nil || v.String() != `{"name": "ADA", "values": [2, 4]}` {
		t.Fatalf("%v, %v", v, err)
	}
	if err := items.Append(std.MakeInt(3)); err != nil {
		t.Fatalf("parent was frozen: %v", err)
	}
	if v.String() != `{"name": "ADA", "values": [2, 4]}` {
		t.Fatalf("result shares parent: %v", v)
	}
	if err := v.(*std.Dict).SetKey(std.String("bad"), std.None); err == nil {
		t.Fatal("result must be frozen")
	}
	for _, src := range []string{`items.append(9)`, `items.clear()`} {
		if _, err := mod.Eval(&std.Thread{}, src, vars); err == nil {
			t.Fatal("accepted input mutation", src)
		}
	}
	if items.String() != "[1, 2, 3]" {
		t.Fatal(items)
	}
	v, err = mod.Eval(&std.Thread{}, `(lambda xs: (xs.append(8), xs))([7])`, nil)
	if err != nil || v.String() != "(None, [7, 8])" {
		t.Fatalf("private intermediate: %v, %v", v, err)
	}
}

func TestEvalResultBoundary(t *testing.T) {
	for _, src := range []string{`lambda: 1`, `"abc".upper`, `{"nested": [len]}`} {
		_, err := mod.Eval(&std.Thread{}, src, nil)
		var e *mod.Error
		if !errors.As(err, &e) || e.Phase != "output" {
			t.Fatalf("%q: %v", src, err)
		}
	}
	v, err := mod.Eval(&std.Thread{}, `(123456789012345678901234567890, b"abc", set([1, 2]), None)`, nil)
	if err != nil {
		t.Fatal(err)
	}
	tuple, ok := v.(std.Tuple)
	if !ok || len(tuple) != 4 || tuple[0].Type() != "int" || tuple[1].Type() != "bytes" || tuple[2].Type() != "set" || tuple[3] != std.None {
		t.Fatal(v)
	}
}

func TestEvalErrors(t *testing.T) {
	for _, tc := range []struct{ source, phase string }{{"1 +", "parse"}, {"missing + 1", "resolve"}, {"1 // 0", "runtime"}} {
		_, err := mod.Eval(&std.Thread{}, tc.source, nil)
		var e *mod.Error
		if !errors.As(err, &e) || e.Phase != tc.phase {
			t.Fatalf("%q: %v", tc.source, err)
		}
		v, err := call(t, "try_eval", tc.source, std.NewDict(0))
		if err != nil {
			t.Fatal(err)
		}
		pair, ok := v.(std.Tuple)
		if !ok || len(pair) != 2 || pair[0] != std.None {
			t.Fatalf("%v", v)
		}
		info, ok := pair[1].(*std.Dict)
		if !ok {
			t.Fatalf("not error data: %v", v)
		}
		phase, found, err := info.Get(std.String("phase"))
		if err != nil || !found || phase != std.String(tc.phase) {
			t.Fatal(info, err)
		}
		if err := info.SetKey(std.String("x"), std.None); err == nil {
			t.Fatal("mutable error")
		}
	}
	v, err := call(t, "try_eval", "None", std.NewDict(0))
	if err != nil || v.String() != "(None, None)" {
		t.Fatalf("%v, %v", v, err)
	}
	if _, err := call(t, "eval", "1 // 0", std.NewDict(0)); err == nil {
		t.Fatal("eval swallowed error")
	}
	if _, err := mod.Eval(nil, "1", nil); err == nil {
		t.Fatal("accepted nil thread")
	}
	if _, err := call(t, "eval", "1", std.NewList(nil)); err == nil {
		t.Fatal("accepted non-dict vars")
	}
}

func TestEvalBuiltinPolicy(t *testing.T) {
	prints, touches := 0, 0
	thread := &std.Thread{Print: func(*std.Thread, string) { prints++ }}
	for _, src := range []string{`print("bad")`, `getattr("abc", "upper")() if False else print("bad")`} {
		if _, err := mod.Eval(thread, src, nil); err == nil {
			t.Fatal("allowed print")
		}
	}
	if prints != 0 {
		t.Fatal("called parent Print")
	}
	std.Universe["dangerous_eval_test"] = std.NewBuiltin("dangerous_eval_test", func(*std.Thread, *std.Builtin, std.Tuple, []std.Tuple) (std.Value, error) {
		touches++
		return std.None, nil
	})
	defer delete(std.Universe, "dangerous_eval_test")
	if _, err := mod.Eval(&std.Thread{}, "dangerous_eval_test()", nil); err == nil || touches != 0 {
		t.Fatalf("dangerous builtin: %v/%d", err, touches)
	}
	v, err := mod.Eval(&std.Thread{}, `getattr("abc", "upper")()`, nil)
	if err != nil || v != std.String("ABC") {
		t.Fatalf("native getattr: %v, %v", v, err)
	}
	v, err = mod.Eval(&std.Thread{}, "len", std.StringDict{"len": std.MakeInt(7)})
	if err != nil || v.String() != "7" {
		t.Fatalf("data shadow: %v, %v", v, err)
	}
	original := std.Universe["len"]
	std.Universe["len"] = std.NewBuiltin("len", func(*std.Thread, *std.Builtin, std.Tuple, []std.Tuple) (std.Value, error) {
		touches++
		return std.None, nil
	})
	v, err = mod.Eval(&std.Thread{}, "len([1])", nil)
	std.Universe["len"] = original
	if err != nil || v.String() != "1" || touches != 0 {
		t.Fatalf("replaced builtin: %v, %v, %d", v, err, touches)
	}
	delete(std.Universe, "len")
	v, err = mod.Eval(&std.Thread{}, "len([1])", nil)
	std.Universe["len"] = original
	if err != nil || v.String() != "1" {
		t.Fatalf("removed builtin: %v, %v", v, err)
	}
}

func TestEvalOuterHelpersAndDataShadow(t *testing.T) {
	vars := std.StringDict{"eval": std.MakeInt(7), "try_eval": std.NewList([]std.Value{std.MakeInt(8)})}
	result, err := mod.Eval(&std.Thread{}, "eval + try_eval[0]", vars)
	if err != nil || result != std.MakeInt(15) {
		t.Fatalf("data shadows builtin names: %v, %v", result, err)
	}
	for _, source := range []string{
		"eval('1')", "try_eval('1')", "sleep(0)", "exit(0)", "quit(0)",
		"eprint('hidden')", "pprint('hidden')", "getenv('PATH')", "runtime.getenv('PATH')",
		"outer_value", "callback()",
	} {
		if _, err := mod.Eval(&std.Thread{}, source, nil); err == nil {
			t.Fatalf("inherited outer helper: %s", source)
		}
	}
	first, err := mod.LoadModule()
	if err != nil {
		t.Fatal(err)
	}
	delete(first, "eval")
	second, err := mod.LoadModule()
	if err != nil || second["eval"] == nil {
		t.Fatalf("module dictionaries share mutations: %v", err)
	}
	second["try_eval"].Freeze()
	value, err := std.Call(&std.Thread{}, second["try_eval"], std.Tuple{std.String("1 + 1")}, nil)
	if err != nil || value.String() != "(2, None)" {
		t.Fatalf("shared immutable builtin: %v, %v", value, err)
	}
}

func TestEvalPreservesCallerState(t *testing.T) {
	thread := &std.Thread{}
	if _, err := mod.Eval(thread, "1 + 2", nil); err != nil {
		t.Fatal(err)
	}
	steps := thread.ExecutionSteps()
	if _, err := mod.Eval(thread, "3 + 4", nil); err != nil {
		t.Fatal(err)
	}
	if thread.ExecutionSteps() <= steps {
		t.Fatal("steps reset")
	}
	thread.Cancel("caller cancelled")
	_, err := std.Call(thread, builtin(t, "try_eval"), std.Tuple{std.String("1 +")}, nil)
	if err == nil || !strings.Contains(err.Error(), "caller cancelled") {
		t.Fatalf("try swallowed cancellation: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ctxThread := &std.Thread{}
	ctxThread.SetLocal("context", ctx)
	_, err = std.Call(ctxThread, builtin(t, "try_eval"), std.Tuple{std.String("1")}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("context cancellation: %v", err)
	}
	limitThread := &std.Thread{}
	limitThread.SetMaxExecutionSteps(1)
	_, err = std.Call(limitThread, builtin(t, "try_eval"), std.Tuple{std.String("1")}, nil)
	if err == nil || !strings.Contains(err.Error(), "too many steps") {
		t.Fatalf("try swallowed limit: %v", err)
	}
	panics := &std.Thread{}
	panics.SetMaxExecutionSteps(1)
	panics.OnMaxSteps = func(*std.Thread) { panic("parent limit") }
	func() {
		defer func() {
			if recover() != "parent limit" {
				t.Error("parent panic swallowed")
			}
		}()
		_, _ = mod.Eval(panics, "1", nil)
	}()
}

func TestEvalArgumentsAndDiagnostics(t *testing.T) {
	called := false
	var list *std.List
	for _, name := range []string{"eval", "try_eval"} {
		for _, source := range []std.Value{nil, list, hostileValue{&called}, std.MakeInt(1)} {
			v, err := std.Call(&std.Thread{}, builtin(t, name), std.Tuple{source}, nil)
			if name == "eval" && err == nil || name == "try_eval" && (err != nil || v.(std.Tuple)[0] != std.None) {
				t.Fatalf("%s(%T): %v, %v", name, source, v, err)
			}
		}
	}
	if called {
		t.Fatal("source rejection executed host behavior")
	}
	for _, args := range []std.Tuple{nil, {std.String("1"), std.NewDict(0), std.None}} {
		v, err := std.Call(&std.Thread{}, builtin(t, "try_eval"), args, nil)
		if err != nil || v.(std.Tuple)[0] != std.None {
			t.Fatalf("argument error: %v, %v", v, err)
		}
	}
	v, err := std.Call(&std.Thread{}, builtin(t, "eval"), std.Tuple{std.String("2 + 3")}, nil)
	if err != nil || v.String() != "5" {
		t.Fatalf("omitted vars: %v, %v", v, err)
	}
	vars := std.NewDict(1)
	_ = vars.SetKey(std.String("n"), std.MakeInt(5))
	v, err = std.Call(&std.Thread{}, builtin(t, "eval"), std.Tuple{std.String("n * 2")}, []std.Tuple{{std.String("vars"), vars}})
	if err != nil || v.String() != "10" {
		t.Fatalf("keyword vars: %v, %v", v, err)
	}
	tooMany := std.NewDict(4097)
	for i := 0; i < 4097; i++ {
		_ = tooMany.SetKey(std.String(strings.Repeat("x", i+1)), std.None)
	}
	if _, err := call(t, "eval", "1", tooMany); err == nil {
		t.Fatal("accepted too many dict variables")
	}
	v, err = call(t, "try_eval", `fail("é" * 10000)`, std.NewDict(0))
	if err != nil {
		t.Fatal(err)
	}
	info := v.(std.Tuple)[1].(*std.Dict)
	message, _, _ := info.Get(std.String("message"))
	truncated, _, _ := info.Get(std.String("truncated"))
	if len(string(message.(std.String))) > 4096 || truncated != std.True || !strings.Contains(string(message.(std.String)), "truncated") {
		t.Fatal("unbounded diagnostic")
	}
	depth := 0
	var outer *std.Builtin
	try := builtin(t, "try_eval")
	outer = std.NewBuiltin("outer", func(th *std.Thread, _ *std.Builtin, _ std.Tuple, _ []std.Tuple) (std.Value, error) {
		if depth < 64 {
			depth++
			defer func() { depth-- }()
			return std.Call(th, outer, nil, nil)
		}
		return std.Call(th, try, std.Tuple{std.String("1 // 0")}, nil)
	})
	v, err = std.Call(&std.Thread{}, outer, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	info = v.(std.Tuple)[1].(*std.Dict)
	frames, _, _ := info.Get(std.String("frames"))
	truncated, _, _ = info.Get(std.String("truncated"))
	if len(frames.(std.Tuple)) != 32 || truncated != std.True {
		t.Fatal("unbounded call frames")
	}
	_, err = std.ExecFileOptions(&syntax.FileOptions{Set: true}, &std.Thread{}, strings.Repeat("p", 1024), `result = evaluator("1 // 0")`, std.StringDict{"evaluator": try})
	if err != nil {
		t.Fatal(err)
	}
	// Inspect through a capturing wrapper so the outer filename is part of
	// the diagnostic while the error itself remains inside try_eval.
	var captured std.Value
	capture := std.NewBuiltin("capture", func(_ *std.Thread, _ *std.Builtin, a std.Tuple, _ []std.Tuple) (std.Value, error) {
		captured = a[0]
		return std.None, nil
	})
	_, err = std.ExecFileOptions(&syntax.FileOptions{Set: true}, &std.Thread{}, strings.Repeat("p", 1024), `capture(evaluator("1 // 0"))`, std.StringDict{"evaluator": try, "capture": capture})
	if err != nil {
		t.Fatal(err)
	}
	info = captured.(std.Tuple)[1].(*std.Dict)
	truncated, _, _ = info.Get(std.String("truncated"))
	if truncated != std.True {
		t.Fatal("long filename truncation not marked")
	}
}

func TestEvalConcurrentFrozenData(t *testing.T) {
	set := std.NewSet(2)
	_ = set.Insert(std.MakeInt(1))
	_ = set.Insert(std.MakeInt(2))
	vars := std.StringDict{"items": std.NewList([]std.Value{std.MakeInt(2), std.MakeInt(3)}), "tags": set}
	vars.Freeze()
	exports, err := mod.LoadModule()
	if err != nil {
		t.Fatal(err)
	}
	arguments := std.NewDict(len(vars))
	for name, value := range vars {
		if err := arguments.SetKey(std.String(name), value); err != nil {
			t.Fatal(err)
		}
	}
	arguments.Freeze()
	cancelled := &std.Thread{}
	cancelled.Cancel("one caller only")
	if _, err := std.Call(cancelled, exports["try_eval"], std.Tuple{std.String("1")}, nil); err == nil {
		t.Fatal("cancelled caller succeeded")
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			v, err := mod.Eval(&std.Thread{}, `(sorted(tags), [n * 2 for n in items])`, vars)
			if err != nil || v.String() != "([1, 2], [4, 6])" {
				t.Errorf("%v, %v", v, err)
			}
			v, err = std.Call(&std.Thread{}, exports["eval"], std.Tuple{std.String(`(sorted(tags), [n * 2 for n in items])`), arguments}, nil)
			if err != nil || v.String() != "([1, 2], [4, 6])" {
				t.Errorf("shared builtin: %v, %v", v, err)
			}
			local := std.NewDict(1)
			if err := local.SetKey(std.String("n"), std.MakeInt(index)); err != nil {
				t.Error(err)
				return
			}
			v, err = std.Call(&std.Thread{}, exports["try_eval"], std.Tuple{std.String("n + 1"), local}, nil)
			if err != nil {
				t.Errorf("independent caller: %v", err)
				return
			}
			pair, ok := v.(std.Tuple)
			if !ok || len(pair) != 2 || pair[0] != std.MakeInt(index+1) || pair[1] != std.None {
				t.Errorf("crossed caller data/state: %v", v)
			}
		}(i)
	}
	group.Wait()
}

func TestEvalBudgetBoundaries(t *testing.T) {
	// Across entry, copy, execution and result publication checkpoints, a
	// depleted caller budget must never turn into a successful try tuple.
	for limit := uint64(1); limit <= 40; limit++ {
		for _, source := range []string{"1 // 0", "list(range(100))"} {
			thread := &std.Thread{}
			thread.SetMaxExecutionSteps(limit)
			value, err := std.Call(thread, builtin(t, "try_eval"), std.Tuple{std.String(source)}, nil)
			if err != nil && !strings.Contains(err.Error(), "too many steps") {
				t.Fatal(err)
			}
			if err == nil && value == nil {
				t.Fatal("nil success")
			}
		}
	}
	thread := &std.Thread{}
	thread.SetMaxExecutionSteps(3)
	items := make([]std.Value, 100)
	for i := range items {
		items[i] = std.None
	}
	_, err := mod.Eval(thread, "x", std.StringDict{"x": std.NewList(items)})
	if err == nil || !strings.Contains(err.Error(), "too many steps") {
		t.Fatalf("copy budget: %v", err)
	}
	t.Run("owner_clears_cancellation_after_failure", func(t *testing.T) {
		caller := &std.Thread{}
		caller.SetMaxExecutionSteps(1)
		checkpoints := 0
		cleared := false
		caller.OnMaxSteps = func(thread *std.Thread) {
			checkpoints++
			// The fixed VM's closed integer probe has two instructions.
			// Let the entry probe finish, cancel during Eval's next probe,
			// then let the owner clear that state at the final checkpoint.
			// try_eval must retain the error that has already occurred.
			if checkpoints == 3 {
				thread.Cancel("owner checkpoint cancellation")
			} else if checkpoints > 3 {
				thread.Uncancel()
				cleared = true
			}
		}
		value, err := std.Call(caller, builtin(t, "try_eval"), std.Tuple{std.String("1")}, nil)
		var execution *mod.Error
		if !cleared || value != std.None || !errors.As(err, &execution) || execution.Phase != "execution" || !strings.Contains(err.Error(), "owner checkpoint cancellation") {
			t.Fatalf("lost prior execution error after owner reset: value=%v error=%v cleared=%v checkpoints=%d", value, err, cleared, checkpoints)
		}
		// The owner's reset is real: a later evaluation may succeed. It
		// does not retroactively turn the failed try_eval into a tuple.
		value, err = mod.Eval(caller, "7", nil)
		if err != nil || value.String() != "7" {
			t.Fatalf("owner did not clear cancellation: %v, %v", value, err)
		}
	})
}

func BenchmarkEval(b *testing.B) {
	for _, name := range []string{"Scalar", "Nested", "TryError"} {
		b.Run(name, func(b *testing.B) {
			thread := &std.Thread{}
			vars := std.StringDict{"qty": std.MakeInt(3), "items": std.NewList([]std.Value{std.MakeInt(1), std.MakeInt(2), std.MakeInt(3)})}
			if name == "Scalar" {
				vars = std.StringDict{"qty": std.MakeInt(3)}
			}
			module, err := mod.LoadModule()
			if err != nil {
				b.Fatal(err)
			}
			try := module["try_eval"]
			if try == nil {
				b.Fatal("missing try_eval builtin")
			}
			if _, err := mod.Eval(thread, "0", nil); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var err error
				switch name {
				case "Scalar":
					_, err = mod.Eval(thread, "qty * 2", vars)
				case "Nested":
					_, err = mod.Eval(thread, `{"values": [x * qty for x in items]}`, vars)
				case "TryError":
					_, err = std.Call(thread, try, std.Tuple{std.String("1 // 0")}, nil)
				}
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkEvalComparison compares ordinary compiled Starlark with the same
// expression passed to eval. Each path reuses one caller Thread with identical
// argument values and a three-argument Starlark call shape. Normal uses its first
// two parameters; EvalReusedVars and TryEvalReusedVars use the equivalent prebuilt
// vars dict; EvalNewVars builds that dict from the parameters on each call.
//
// Input creation, wrapper compilation, initialization and result checks are not
// timed. The eval paths include source parsing/compilation and data isolation on
// every call; there is no expression cache. This is a feature-cost comparison,
// not a before/after regression benchmark.
func BenchmarkEvalComparison(b *testing.B) {
	items := make([]std.Value, 32)
	products := make([]std.Value, len(items))
	for i := range items {
		items[i] = std.MakeInt(i)
		products[i] = std.MakeInt(i * 3)
	}
	nestedResult := std.NewDict(1)
	if err := nestedResult.SetKey(std.String("values"), std.NewList(products)); err != nil {
		b.Fatal(err)
	}
	cases := []struct {
		name, source, newVars string
		names                 []string
		values                std.Tuple
		want                  std.Value
	}{
		{
			name: "Scalar", source: "price * qty",
			newVars: `{"price": price, "qty": qty}`,
			names:   []string{"price", "qty"},
			values:  std.Tuple{std.MakeInt(12), std.MakeInt(3)},
			want:    std.MakeInt(36),
		},
		{
			name: "Nested32", source: `{"values": [x * qty for x in items]}`,
			newVars: `{"items": items, "qty": qty}`,
			names:   []string{"items", "qty"},
			values:  std.Tuple{std.NewList(items), std.MakeInt(3)},
			want:    nestedResult,
		},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			vars := std.NewDict(len(tc.names))
			for i, name := range tc.names {
				if err := vars.SetKey(std.String(name), tc.values[i]); err != nil {
					b.Fatal(err)
				}
			}
			vars.Freeze()
			args := append(append(std.Tuple(nil), tc.values...), vars)
			module, err := mod.LoadModule()
			if err != nil {
				b.Fatal(err)
			}
			parameters := strings.Join(tc.names, ", ") + ", vars"
			quotedSource := std.String(tc.source).String()
			script := "def normal(" + parameters + "):\n    return " + tc.source + "\n\n" +
				"def eval_reused(" + parameters + "):\n    return eval(" + quotedSource + ", vars=vars)\n\n" +
				"def eval_new(" + parameters + "):\n    return eval(" + quotedSource + ", vars=" + tc.newVars + ")\n\n" +
				"def try_reused(" + parameters + "):\n    value, error = try_eval(" + quotedSource + ", vars=vars)\n    return value\n"
			functions, err := std.ExecFileOptions(&syntax.FileOptions{Set: true}, &std.Thread{}, "eval_comparison.star", script, module)
			if err != nil {
				b.Fatal(err)
			}
			for _, path := range []struct{ name, function string }{
				{"Normal", "normal"},
				{"EvalReusedVars", "eval_reused"},
				{"EvalNewVars", "eval_new"},
				{"TryEvalReusedVars", "try_reused"},
			} {
				b.Run(path.name, func(b *testing.B) {
					thread := &std.Thread{Name: "eval-comparison"}
					function := functions[path.function]
					if path.function == "try_reused" {
						value, err := std.Call(thread, module["try_eval"], std.Tuple{std.String(tc.source), vars}, nil)
						if err != nil {
							b.Fatal(err)
						}
						pair, ok := value.(std.Tuple)
						if !ok || len(pair) != 2 || pair[1] != std.None {
							b.Fatalf("try_eval did not succeed: %v", value)
						}
					}
					result, err := std.Call(thread, function, args, nil)
					if err != nil {
						b.Fatal(err)
					}
					if equal, err := std.Equal(result, tc.want); err != nil || !equal {
						b.Fatalf("wrong result: %v, want %v (error %v)", result, tc.want, err)
					}
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						result, err = std.Call(thread, function, args, nil)
						if err != nil {
							b.Fatal(err)
						}
					}
					b.StopTimer()
					if equal, err := std.Equal(result, tc.want); err != nil || !equal {
						b.Fatalf("wrong final result: %v, want %v (error %v)", result, tc.want, err)
					}
				})
			}
		})
	}
}

// BenchmarkEvalWorkload keeps source, input shape and scalar output shape fixed
// while increasing work inside the expression. Normal and EvalReusedVars use
// identical arguments and the same two-argument Starlark call shape. Every eval
// includes parsing, compilation and data isolation; setup and result checks are
// outside timing. Differences estimate the incremental cost of this expression
// on the measured system, not a universal constant or a before/after regression.
func BenchmarkEvalWorkload(b *testing.B) {
	const source = "len([(i * 3 + 7) % 97 for i in range(n)])"
	module, err := mod.LoadModule()
	if err != nil {
		b.Fatal(err)
	}
	script := "def normal(n, vars):\n    return " + source + "\n\n" +
		"def eval_reused(n, vars):\n    return eval(" + std.String(source).String() + ", vars=vars)\n"
	functions, err := std.ExecFileOptions(&syntax.FileOptions{Set: true}, &std.Thread{}, "eval_workload.star", script, module)
	if err != nil {
		b.Fatal(err)
	}
	for _, workload := range []struct {
		name string
		n    int
	}{
		{"N0", 0},
		{"N10", 10},
		{"N100", 100},
		{"N1000", 1000},
		{"N10000", 10000},
		{"N100000", 100000},
	} {
		b.Run(workload.name, func(b *testing.B) {
			want := std.MakeInt(workload.n)
			vars := std.NewDict(1)
			if err := vars.SetKey(std.String("n"), want); err != nil {
				b.Fatal(err)
			}
			vars.Freeze()
			args := std.Tuple{want, vars}
			for _, path := range []struct{ name, function string }{
				{"Normal", "normal"},
				{"EvalReusedVars", "eval_reused"},
			} {
				b.Run(path.name, func(b *testing.B) {
					thread := &std.Thread{Name: "eval-workload"}
					function := functions[path.function]
					result, err := std.Call(thread, function, args, nil)
					if err != nil {
						b.Fatal(err)
					}
					if equal, err := std.Equal(result, want); err != nil || !equal {
						b.Fatalf("wrong result: %v, want %v (error %v)", result, want, err)
					}
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						result, err = std.Call(thread, function, args, nil)
						if err != nil {
							b.Fatal(err)
						}
					}
					b.StopTimer()
					if equal, err := std.Equal(result, want); err != nil || !equal {
						b.Fatalf("wrong final result: %v, want %v (error %v)", result, want, err)
					}
				})
			}
		})
	}
}
