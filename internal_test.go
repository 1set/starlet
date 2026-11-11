package starlet

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// REPL print-budget sections: separate input chunks share one allowance, the
// error boundary only handles its own budget sentinel, and zero is unlimited.
func TestREPLPrintByteBudget(t *testing.T) {
	for _, limit := range []uint64{0, 2} {
		m := NewDefault()
		m.SetMaxPrintBytes(limit)
		var delivered strings.Builder
		m.SetPrintFunc(func(_ *starlark.Thread, msg string) { delivered.WriteString(msg) })
		var reported error
		completed := false
		m.replInternal(func(opts *syntax.FileOptions, thread *starlark.Thread, globals starlark.StringDict) {
			for _, input := range []string{`print("a")`, `print("b")`, `print("c")`} {
				if _, err := starlark.ExecFileOptions(opts, thread, "repl-input.star", input, globals); err != nil {
					t.Fatal(err)
				}
			}
			completed = true
		}, func(err error) { reported = err })
		if limit == 0 {
			if reported != nil || !completed || delivered.String() != "abc" {
				t.Fatalf("unlimited session: completed=%v output=%q err=%v", completed, delivered.String(), reported)
			}
		} else {
			var limitErr PrintLimitExceededError
			if !errors.As(reported, &limitErr) || limitErr.Limit != limit || completed || delivered.String() != "ab" {
				t.Fatalf("limited session: completed=%v output=%q err=%v", completed, delivered.String(), reported)
			}
		}
		// The boundary must release the Machine lock after either outcome.
		m.SetMaxPrintBytes(0)
	}
}

func TestREPLPrintHandlerPanicIsNotSwallowed(t *testing.T) {
	m := NewDefault()
	m.SetMaxPrintBytes(2)
	sentinel := errors.New("host print callback failed")
	m.SetPrintFunc(func(_ *starlark.Thread, _ string) { panic(sentinel) })
	defer func() {
		if got := recover(); got != sentinel {
			t.Fatalf("host panic changed: %v", got)
		}
	}()
	m.replInternal(func(opts *syntax.FileOptions, thread *starlark.Thread, globals starlark.StringDict) {
		_, _ = starlark.ExecFileOptions(opts, thread, "repl-input.star", `print("a")`, globals)
	}, func(err error) { t.Fatalf("host panic was reported as a budget error: %v", err) })
}

func TestREPLNilPrintHandlerPanicIsNotSwallowed(t *testing.T) {
	m := NewDefault()
	m.SetMaxPrintBytes(2)
	m.SetPrintFunc(func(_ *starlark.Thread, _ string) { panic(nil) })
	returned := false
	func() {
		defer func() { _ = recover() }()
		m.replInternal(func(opts *syntax.FileOptions, thread *starlark.Thread, globals starlark.StringDict) {
			_, _ = starlark.ExecFileOptions(opts, thread, "repl-input.star", `print("a")`, globals)
		}, func(err error) { t.Fatalf("host panic was reported as a budget error: %v", err) })
		returned = true
	}()
	if returned {
		t.Fatal("REPL swallowed the host callback's panic(nil)")
	}
}

// TestWatchLoadThread covers the load-thread cancel watcher's two arms. A load
// thread with no run context (e.g. a load during a REPL, which sets none) must
// get a harmless no-op stop rather than watch a nil context; a thread carrying a
// live, cancellable context gets a real watcher, and cancelling then stopping it
// must be safe. The firing-on-timeout behaviour itself is covered end to end by
// TestMachine_LoadInheritsMachineContext.
func TestWatchLoadThread(t *testing.T) {
	m := NewDefault()

	// no context on the thread -> no-op stop.
	stop := m.watchLoadThread(&starlark.Thread{})
	stop() // must not panic

	// a live context -> a real watcher; cancel then stop must be safe/idempotent.
	ctx, cancel := context.WithCancel(context.Background())
	th := &starlark.Thread{}
	th.SetLocal("context", ctx)
	stop = m.watchLoadThread(th)
	cancel()
	stop()
	stop()
}

func TestCastStringDictToAnyMap(t *testing.T) {
	// Create a starlark.StringDict
	m := starlark.StringDict{
		"key1": starlark.String("value1"),
		"key2": starlark.String("value2"),
		"key3": starlark.NewList([]starlark.Value{starlark.String("value3")}),
	}

	// Convert to StringAnyMap
	anyMap := castStringDictToAnyMap(m)

	// Check that the conversion was successful
	if anyMap["key1"].(starlark.String) != "value1" {
		t.Errorf("Expected 'value1', got '%s'", anyMap["key1"].(starlark.String))
	}
	if anyMap["key2"].(starlark.String) != "value2" {
		t.Errorf("Expected 'value2', got '%s'", anyMap["key2"].(starlark.String))
	}
	if !reflect.DeepEqual(anyMap["key3"], m["key3"]) {
		t.Errorf("Expected '%v', got '%v'", m["key3"], anyMap["key3"])
	}
}

func TestCastStringAnyMapToStringDict(t *testing.T) {
	// Create a StringAnyMap
	m := StringAnyMap{
		"key1": starlark.String("value1"),
		"key2": starlark.String("value2"),
		"key3": starlark.NewList([]starlark.Value{starlark.String("value3")}),
	}

	// Convert to starlark.StringDict
	stringDict, err := castStringAnyMapToStringDict(m)

	// Check that the conversion was successful
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if stringDict["key1"].(starlark.String) != "value1" {
		t.Errorf("Expected 'value1', got '%s'", stringDict["key1"].(starlark.String))
	}
	if stringDict["key2"].(starlark.String) != "value2" {
		t.Errorf("Expected 'value2', got '%s'", stringDict["key2"].(starlark.String))
	}
	if !reflect.DeepEqual(stringDict["key3"], m["key3"]) {
		t.Errorf("Expected '%v', got '%v'", m["key3"], stringDict["key3"])
	}

	// Test with a non-starlark.Value in the map
	m["key4"] = "value4"
	_, err = castStringAnyMapToStringDict(m)

	// Check that an error was returned
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestSetInputConversionEnabled(t *testing.T) {
	m := NewDefault()
	if m.enableInConv != true {
		t.Errorf("Expected input conversion to be enabled by default, but it wasn't")
	}

	m.SetInputConversionEnabled(false)
	if m.enableInConv != false {
		t.Errorf("Expected input conversion to be disabled, but it wasn't")
	}

	m.SetInputConversionEnabled(true)
	if m.enableInConv != true {
		t.Errorf("Expected input conversion to be enabled, but it wasn't")
	}
}

func TestSetOutputConversionEnabled(t *testing.T) {
	m := NewDefault()
	if m.enableOutConv != true {
		t.Errorf("Expected output conversion to be enabled by default, but it wasn't")
	}

	m.SetOutputConversionEnabled(false)
	if m.enableOutConv != false {
		t.Errorf("Expected output conversion to be disabled, but it wasn't")
	}

	m.SetOutputConversionEnabled(true)
	if m.enableOutConv != true {
		t.Errorf("Expected output conversion to be enabled, but it wasn't")
	}
}

func TestGetCacheKey(t *testing.T) {
	m1 := NewDefault()
	k1, ok := m1.getCacheKey("a = 1")
	if !ok || k1 == "" {
		t.Errorf("expected a usable key for a string source, got %q ok=%v", k1, ok)
	}

	// bytes and string of the same content agree
	if k, _ := m1.getCacheKey([]byte("a = 1")); k != k1 {
		t.Errorf("expected the byte and string keys to agree, got %q vs %q", k, k1)
	}

	// reader-like sources cannot be content-keyed and must skip the cache
	if _, ok := m1.getCacheKey(strings.NewReader("a = 1")); ok {
		t.Errorf("expected a reader source to be uncacheable")
	}

	// a different predeclared name set changes the key
	m2 := NewDefault()
	m2.predeclared = starlark.StringDict{"x": starlark.None}
	if k, _ := m2.getCacheKey("a = 1"); k == k1 {
		t.Errorf("expected the predeclared name set to be part of the key")
	}

	// dialect bits change the key
	m3 := NewDefault()
	m3.allowRecursion = true
	if k, _ := m3.getCacheKey("a = 1"); k == k1 {
		t.Errorf("expected the dialect bits to be part of the key")
	}
	m4 := NewDefault()
	m4.allowGlobalReassign = true
	if k, _ := m4.getCacheKey("a = 1"); k == k1 {
		t.Errorf("expected the global-reassign bit to be part of the key")
	}
}

func TestExecStarlarkFileUncacheableSource(t *testing.T) {
	// a source that cannot be content-keyed must execute without touching
	// the cache instead of risking a stale filename-only hit
	m := NewDefault()
	m.SetScriptCacheEnabled(true)
	m.SetScript("seed.star", []byte("a = 1"), nil)
	if _, err := m.Run(); err != nil {
		t.Fatalf("seed run expects no error, got: %v", err)
	}
	res, err := m.execStarlarkFile("reader.star", strings.NewReader("zz = 9"), true)
	if err != nil {
		t.Fatalf("reader source expects no error, got: %v", err)
	}
	if v, ok := res["zz"]; !ok || v.String() != "9" {
		t.Errorf("expected zz = 9 from the uncached reader source, got: %v", res)
	}
}
