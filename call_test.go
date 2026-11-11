package starlet_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/1set/starlet"
	"go.starlark.net/starlark"
)

// Print budget sections: per-call reset, overflow termination, and updating the
// handler/limit on a reused thread.
func TestMachine_CallPrintByteBudget(t *testing.T) {
	m := starlet.NewDefault()
	if _, err := m.RunScript([]byte("def emit(s):\n    print(s)\n    return 7"), nil); err != nil {
		t.Fatal(err)
	}
	m.SetMaxPrintBytes(2)
	delivered := 0
	m.SetPrintFunc(func(_ *starlark.Thread, msg string) { delivered += len(msg) })
	for i := 0; i < 2; i++ {
		if _, err := m.CallWithTimeout(time.Second, "emit", "ok"); err != nil {
			t.Fatalf("call %d did not get a fresh allowance: %v", i, err)
		}
	}
	var limitErr starlet.PrintLimitExceededError
	if out, err := m.Call("emit", "too long"); !errors.As(err, &limitErr) || out != nil {
		t.Fatalf("want withheld result and typed limit: out=%v err=%v", out, err)
	}
	if delivered != 4 {
		t.Fatalf("delivered %d bytes, want 4", delivered)
	}
	m.SetMaxPrintBytes(0)
	m.SetPrintFunc(starlet.NoopPrintFunc)
	if _, err := m.CallWithContext(context.Background(), "emit", "unlimited again"); err != nil {
		t.Fatalf("clearing the limit after a rejected call: %v", err)
	}
}

// Call print configuration changes take effect on the next execution, including
// replacing handlers on an existing thread and clearing an exhausted allowance.
func TestMachine_CallPrintConfigurationChanges(t *testing.T) {
	const script = "def emit(s):\n    print(s)\n    return 7"
	t.Run("replace unlimited handler after Run", func(t *testing.T) {
		m := starlet.NewDefault()
		var original, replacement strings.Builder
		m.SetPrintFunc(func(_ *starlark.Thread, msg string) { original.WriteString(msg) })
		if _, err := m.RunScript([]byte(script), nil); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Call("emit", "first"); err != nil {
			t.Fatal(err)
		}
		m.SetPrintFunc(func(_ *starlark.Thread, msg string) { replacement.WriteString(msg) })
		if out, err := m.Call("emit", "second"); err != nil || out != int64(7) {
			t.Fatalf("replacement call: out=%v err=%v", out, err)
		}
		if original.String() != "first" || replacement.String() != "second" {
			t.Fatalf("handlers: original=%q replacement=%q", original.String(), replacement.String())
		}
	})
	t.Run("disable limit after replacing handler", func(t *testing.T) {
		m := starlet.NewDefault()
		var original, replacement strings.Builder
		m.SetMaxPrintBytes(2)
		m.SetPrintFunc(func(_ *starlark.Thread, msg string) { original.WriteString(msg) })
		if _, err := m.RunScript([]byte(script), nil); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Call("emit", "ok"); err != nil {
			t.Fatal(err)
		}
		m.SetPrintFunc(func(_ *starlark.Thread, msg string) { replacement.WriteString(msg) })
		if _, err := m.Call("emit", "hi"); err != nil {
			t.Fatalf("limited replacement handler: %v", err)
		}
		var limitErr starlet.PrintLimitExceededError
		if _, err := m.Call("emit", "rejected"); !errors.As(err, &limitErr) {
			t.Fatalf("want rejected message before clearing allowance: %v", err)
		}
		m.SetMaxPrintBytes(0)
		if _, err := m.Call("emit", "unlimited"); err != nil {
			t.Fatalf("clearing limit: %v", err)
		}
		if original.String() != "ok" || replacement.String() != "hiunlimited" {
			t.Fatalf("handlers: original=%q replacement=%q", original.String(), replacement.String())
		}
	})
	t.Run("nil handler restores default stderr", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "call-stderr-")
		if err != nil {
			t.Fatal(err)
		}
		previous := os.Stderr
		os.Stderr = f
		defer func() { os.Stderr = previous; _ = f.Close() }()
		for _, limit := range []uint64{0, 2} {
			m := starlet.NewDefault()
			var original strings.Builder
			m.SetPrintFunc(func(_ *starlark.Thread, msg string) { original.WriteString(msg) })
			m.SetMaxPrintBytes(limit)
			if _, err := m.RunScript([]byte(script), nil); err != nil {
				t.Fatal(err)
			}
			if _, err := m.Call("emit", "ok"); err != nil {
				t.Fatal(err)
			}
			m.SetPrintFunc(nil)
			if limit != 0 {
				m.SetMaxPrintBytes(0)
			}
			if _, err := m.Call("emit", "default handler"); err != nil {
				t.Fatalf("limit=%d default handler: %v", limit, err)
			}
			if original.String() != "ok" {
				t.Fatalf("limit=%d retained previous handler: %q", limit, original.String())
			}
		}
		os.Stderr = previous
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(f.Name())
		if err != nil || string(got) != "default handler\ndefault handler\n" {
			t.Fatalf("default stderr: err=%v output=%q", err, got)
		}
	})
}

func TestMachine_CallLoadedFunctionPrintByteBudget(t *testing.T) {
	for _, cached := range []bool{false, true} {
		t.Run(fmt.Sprintf("compiled cache=%v", cached), func(t *testing.T) {
			m := starlet.NewDefault()
			m.SetScriptCacheEnabled(cached)
			m.SetPrintFunc(starlet.NoopPrintFunc)
			m.SetMaxPrintBytes(3)
			m.SetScript("main.star", []byte("load(\"emit.star\", \"emit\")\nexported = emit"), fstest.MapFS{
				"emit.star": &fstest.MapFile{Data: []byte("print(\"abc\")\ndef emit(s):\n    print(s)")},
			})
			for run := 0; run < 2; run++ {
				if _, err := m.Run(); err != nil {
					t.Fatalf("run %d: %v", run, err)
				}
				if _, err := m.Call("exported", "abc"); err != nil {
					t.Fatalf("loaded function retained the previous allowance: %v", err)
				}
				var limitErr starlet.PrintLimitExceededError
				if _, err := m.CallWithContext(context.Background(), "exported", "abcd"); !errors.As(err, &limitErr) {
					t.Fatalf("loaded function bypassed the current allowance: %v", err)
				}
			}
		})
	}
}

func TestMachine_Call_Preconditions(t *testing.T) {
	m := starlet.NewDefault()

	// test: if name == ""
	_, err := m.Call("")
	expectErr(t, err, "starlet: call: no function name")

	// test: if m.thread == nil
	_, err = m.Call("no_thread")
	expectErr(t, err, "starlet: call: no function loaded")

	// test: if m.predeclared == nil
	m.SetGlobals(map[string]interface{}{"x": 1})
	_, err = m.Call("no_globals")
	expectErr(t, err, "starlet: call: no function loaded")

	// prepare: run a script to load a function if exists
	_, err = m.RunScript([]byte(`y = 2`), map[string]interface{}{
		"println": fmt.Println,
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// test: if no such function
	_, err = m.Call("no_such_function")
	expectErr(t, err, "starlet: call: no such function: no_such_function")

	// test: if mistyped function
	_, err = m.Call("y")
	expectErr(t, err, "starlet: call: mistyped function: y")

	ei := err.(starlet.ExecError).Unwrap()
	expectErr(t, ei, "mistyped function: y")

	// test: if builtin function
	_, err = m.Call("println", "hello")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMachine_Call_Functions(t *testing.T) {
	type mockData struct {
		Apple   string `json:"app"`
		Banana  int64  `json:"bana"`
		Coconut bool   `json:"coco"`
	}
	tests := []struct {
		name    string
		code    string
		args    []interface{}
		want    interface{}
		wantErr string
	}{
		{
			name: "no args nor return",
			code: `
def work():
	pass
`,
			want: nil,
		},
		{
			name: "no args but return",
			code: `
def work():
	return 1
`,
			want: int64(1),
		},
		{
			name: "args but no return",
			code: `
def work(x, y):
	pass
`,
			args: []interface{}{1, 2},
			want: nil,
		},
		{
			name: "args and return",
			code: `
def work(x, y):
	return x + y
`,
			args: []interface{}{1, 2},
			want: int64(3),
		},
		{
			name: "lambda",
			code: `
work = lambda x, y: x * y
`,
			args: []interface{}{2, 3},
			want: int64(6),
		},
		{
			name: "multiple return",
			code: `
def work(x, y):
	return x + 1, y + 2
`,
			args: []interface{}{1, 2},
			want: []interface{}{int64(2), int64(4)},
		},
		{
			name: "multiple return with tuple",
			code: `
def work(x, y):
	return (x + 1, y + 2)
`,
			args: []interface{}{1, 2},
			want: []interface{}{int64(2), int64(4)},
		},
		{
			name: "multiple return with list",
			code: `
def work(x, y):
	return [x + 1, y + 2]
`,
			args: []interface{}{1, 2},
			want: []interface{}{int64(2), int64(4)},
		},
		{
			name: "convert args fail",
			code: `
def work(x, y):
	return x + y
`,
			args:    []interface{}{1, make(chan int64)},
			wantErr: `starlight: convert args: type chan int64 is not a supported starlark type`,
		},
		{
			name: "invalid args",
			code: `
def work(x, y):
	return x + y
`,
			args:    []interface{}{1, "two"},
			wantErr: `starlark: call: unknown binary op: int + string`,
		},
		{
			name: "func runtime error",
			code: `
def work(x, y):
	fail("oops")
`,
			args:    []interface{}{1, 2},
			wantErr: `starlark: call: fail: oops`,
		},
		{
			name: "func runtime panic",
			code: `
def work(x, y):
	panic("outside starlark")
`,
			args:    []interface{}{1, 2},
			wantErr: `starlark: call: panic: as expected`,
		},
		{
			name: "custom tag for arg",
			args: []interface{}{
				mockData{Apple: "red", Banana: 2, Coconut: true},
			},
			code: `
def work(x):
	return "apple is " + x.app + ", banana is " + str(x.bana) + ", coconut is " + str(x.coco)
`,
			want: "apple is red, banana is 2, coconut is True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare to load
			m := starlet.NewDefault()
			m.SetCustomTag("json")
			_, err := m.RunScript([]byte(tt.code), map[string]interface{}{
				"panic": starlark.NewBuiltin("panic", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
					panic(errors.New("as expected"))
				}),
			})
			if err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			// call and check
			got, err := m.Call("work", tt.args...)
			if err != nil {
				if tt.wantErr == "" {
					t.Errorf("expected no error, got %v", err)
				} else {
					expectErr(t, err, tt.wantErr)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected %v (%T), got %v (%T)", tt.want, tt.want, got, got)
			}
		})
	}
}

func TestMachine_Call_Convert(t *testing.T) {
	m := starlet.NewDefault()
	m.SetCustomTag("json")
	_, err := m.RunScript([]byte(`
def work(x, y):
	return x * y
`), nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
		return
	}

	// call and check --- conversion is enabled by default
	got, err := m.Call("work", 10, 20)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
		return
	}
	if got != int64(200) {
		t.Errorf("expected 200, got %v", got)
	}

	// not convert
	m.SetOutputConversionEnabled(false)
	got, err = m.Call("work", 5, 6)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
		return
	}
	if v, ok := got.(starlark.Value); !ok {
		t.Errorf("expected starlark Value, got %T", got)
		return
	} else if v != starlark.MakeInt(30) {
		t.Errorf("got unexpected value: %v", v)
	}
}

func TestMachine_CallWithContext(t *testing.T) {
	prep := func() *starlet.Machine {
		m := starlet.NewWithNames(nil, []string{"go_idiomatic"}, nil)
		m.SetScript("test.star", []byte("def slow():\n    sleep(2)\n    return 1\n\ndef quick(x):\n    return x + 1\n"), nil)
		if _, err := m.Run(); err != nil {
			t.Fatalf("prep run expects no error, got: %v", err)
		}
		return m
	}

	// the timeout aborts the call long before the function finishes
	{
		m := prep()
		ts := time.Now()
		_, err := m.CallWithTimeout(200*time.Millisecond, "slow")
		if err == nil || !strings.Contains(err.Error(), "context") {
			t.Errorf("expected a context error from the timed-out call, got: %v", err)
		}
		if elapsed := time.Since(ts); elapsed > time.Second {
			t.Errorf("call was not aborted in time: %v", elapsed)
		}
	}

	// an already-cancelled context fails immediately
	{
		m := prep()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := m.CallWithContext(ctx, "quick", 1)
		if err == nil || !strings.Contains(err.Error(), "starlet: call: context canceled") {
			t.Errorf("expected fail-fast on a cancelled context, got: %v", err)
		}
	}

	// a nil context behaves like plain Call
	{
		m := prep()
		out, err := m.CallWithContext(nil, "quick", 41)
		if err != nil || out != int64(42) {
			t.Errorf("expected 42 with no error, got: %v / %v", out, err)
		}
	}

	// a live context lets the call complete
	{
		m := prep()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		out, err := m.CallWithContext(ctx, "quick", 1)
		if err != nil || out != int64(2) {
			t.Errorf("expected 2 with no error, got: %v / %v", out, err)
		}
	}
}
