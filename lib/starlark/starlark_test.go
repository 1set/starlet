package starlark_test

// Sections: data isolation; result boundaries; errors; caller execution state.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	mod "github.com/1set/starlet/lib/starlark"
	std "go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func builtin(t *testing.T, name string) std.Value {
	t.Helper()
	m, err := mod.LoadModule()
	if err != nil {
		t.Fatal(err)
	}
	v, err := m[mod.ModuleName].(std.HasAttrs).Attr(name)
	if err != nil || v == nil {
		t.Fatalf("%s: %v", name, err)
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
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			v, err := mod.Eval(&std.Thread{}, `(sorted(tags), [n * 2 for n in items])`, vars)
			if err != nil || v.String() != "([1, 2], [4, 6])" {
				t.Errorf("%v, %v", v, err)
			}
		}()
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
			try, err := module[mod.ModuleName].(std.HasAttrs).Attr("try_eval")
			if err != nil {
				b.Fatal(err)
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
