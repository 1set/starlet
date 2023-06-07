package starlet_test

import (
	"context"
	"os"
	"starlet"
	"testing"
)

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
	code := `print("Aloha, Honua!")`
	m.SetScript("aloha.star", []byte(code), nil)
	// run
	_, err := m.Run(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// compare
	cmpFunc("Aloha, Honua!\n")
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
	m.SetScript("test.star", []byte(code), os.DirFS("example"))
	// run
	_, err := m.Run(context.Background())
	// check result
	if isOnWindows {
		expectErr(t, err, `starlet: exec: cannot load nonexist.star:`, `The system cannot find the file specified.`)
	} else {
		expectErr(t, err, `starlet: exec: cannot load nonexist.star:`, `: no such file or directory`)
	}
}

func Test_Machine_Run_Globals(t *testing.T) {
	m := starlet.NewMachine(map[string]interface{}{
		"a": 2,
	}, nil, nil)
	// set code
	code := `b = a * 10`
	m.SetScript("test.star", []byte(code), nil)
	// run
	out, err := m.Run(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["b"].(int64) != int64(20) {
		t.Errorf("unexpected output: %v", out)
	}
}

func Test_Machine_Run_Globals_Miss(t *testing.T) {
	m := starlet.NewMachine(map[string]interface{}{
		"a": 2,
	}, nil, nil)
	// set code
	code := `b = c * 10`
	m.SetScript("test.star", []byte(code), nil)
	// run
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: exec: test.star:1:5: undefined: c`)
}

func Test_Machine_Run_PreloadModules(t *testing.T) {
	m := starlet.NewMachine(nil, []starlet.ModuleName{starlet.ModuleGoIdiomatic}, nil)
	// set code
	code := `a = nil == None`
	m.SetScript("test.star", []byte(code), nil)
	// run
	out, err := m.Run(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["a"].(bool) != true {
		t.Errorf("unexpected output: %v", out)
	}
}

func Test_Machine_Run_PreloadModules_Miss(t *testing.T) {
	m := starlet.NewMachine(nil, []starlet.ModuleName{}, nil)
	// set code
	code := `a = nil == None`
	m.SetScript("test.star", []byte(code), nil)
	// run
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: exec: test.star:1:5: undefined: nil`)
}

func Test_Machine_Run_PreloadModules_NonExist(t *testing.T) {
	m := starlet.NewMachine(nil, []starlet.ModuleName{"nonexist"}, nil)
	// set code
	code := `a = nil == None`
	m.SetScript("test.star", []byte(code), nil)
	// run
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: preload modules: load module "nonexist": module not found`)
}
