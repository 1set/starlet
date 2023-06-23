package starlet_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/1set/starlet"
	"go.starlark.net/starlark"
)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare to load
			m := starlet.NewDefault()
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
