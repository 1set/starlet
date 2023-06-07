package starlet_test

import (
	"context"
	"os"
	"starlet"
	"strings"
	"testing"

	"go.starlark.net/starlark"
)

func expectErr(t *testing.T, err error, expected string) {
	if err == nil {
		t.Errorf("unexpected nil error")
	}
	if err.Error() != expected {
		t.Errorf(`expected error %q, got: %v`, expected, err)
	}
}

// getPrintCompareFunc returns a print function and a compare function.
func getPrintCompareFunc(t *testing.T) (starlet.PrintFunc, func(s string)) {
	var sb strings.Builder
	return func(thread *starlark.Thread, msg string) {
			sb.WriteString(msg)
			sb.WriteString("\n")
		}, func(exp string) {
			act := sb.String()
			if act != exp {
				t.Errorf("expected print %q, got %q", exp, act)
			}
		}
}

func Test_EmptyMachine_Run_NoCode(t *testing.T) {
	m := starlet.NewEmptyMachine()
	// run with empty script
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: run: no code to run`)
}

func Test_EmptyMachine_Run_NoSpecificFile(t *testing.T) {
	m := starlet.NewEmptyMachine()
	m.SetScript("", nil, os.DirFS("example"))
	// run with no specific file name
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: run: no code to run`)
}

func Test_EmptyMachine_Run_APlusB(t *testing.T) {
	m := starlet.NewEmptyMachine()
	code := `a = 1 + 2`
	m.SetScript("a_plus_b.star", []byte(code), nil)
	out, err := m.Run(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["a"] != int64(3) {
		t.Errorf("unexpected output: %v", out)
	}
}

func Test_EmptyMachine_Run_HelloWorld(t *testing.T) {
	m := starlet.NewEmptyMachine()
	// set print function
	printFunc, cmpFunc := getPrintCompareFunc(t)
	m.SetPrintFunc(printFunc)
	// set code
	code := `print("Aloha kāua!")`
	m.SetScript("aloha.star", []byte(code), nil)
	// run
	_, err := m.Run(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// compare
	cmpFunc("Aloha kāua!\n")
}

func Test_EmptyMachine_Run_LoadNoFS(t *testing.T) {
	m := starlet.NewEmptyMachine()
	// set code
	code := `load("fibonacci.star", "fibonacci")`
	m.SetScript("test.star", []byte(code), nil)
	// run
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: exec: cannot load fibonacci.star: no file system given`)
}

func Test_EmptyMachine_Run_LoadNonexist(t *testing.T) {
	m := starlet.NewEmptyMachine()
	// set code
	code := `load("nonexist.star", "a")`
	m.SetScript("load_nonexist.star", []byte(code), os.DirFS("example"))
	// run
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: exec: cannot load nonexist.star: open example/nonexist.star: no such file or directory`)
}
