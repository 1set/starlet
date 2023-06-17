package starlet_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/1set/starlet"
)

func TestMachine_Call_Preconditions(t *testing.T) {
	m := starlet.NewDefault()

	// test: if name == ""
	_, err := m.Call("")
	expectErr(t, err, "no function name")

	// test: if m.thread == nil
	_, err = m.Call("no_thread")
	expectErr(t, err, "no function loaded")

	// test: if m.predeclared == nil
	m.SetGlobals(map[string]interface{}{"x": 1})
	_, err = m.Call("no_globals")
	expectErr(t, err, "no function loaded")

	// prepare: run a script to load a function if exists
	_, err = m.RunScript([]byte(`y = 2`), map[string]interface{}{
		"println": fmt.Println,
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// test: if no such function
	_, err = m.Call("no_such_function")
	expectErr(t, err, "no such function: no_such_function")

	// test: if mistyped function
	_, err = m.Call("y")
	expectErr(t, err, "mistyped function: y")

	// test: if builtin function
	_, err = m.Call("println")
	expectErr(t, err, "mistyped function: println")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare to load
			m := starlet.NewDefault()
			_, err := m.RunScript([]byte(tt.code), nil)
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
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}