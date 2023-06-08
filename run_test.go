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

func Test_EmptyMachine_Run_LoadFunc(t *testing.T) {
	m := starlet.NewEmptyMachine()
	// set code
	code := `load("fibonacci.star", "fibonacci"); val = fibonacci(10)[-1]`
	m.SetScript("test.star", []byte(code), os.DirFS("example"))
	// run
	out, err := m.Run(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// check result
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["val"] != int64(55) {
		t.Errorf("unexpected output: %v", out)
	}
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

func Test_EmptyMachine_Run_LoadNonExist(t *testing.T) {
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

func Test_Machine_Run_AllowedModules(t *testing.T) {
	m := starlet.NewMachine(nil, nil, []starlet.ModuleName{starlet.ModuleGoIdiomatic})
	// set code
	code := `
load("go_idiomatic", "nil")
a = nil == None
`
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

func Test_Machine_Run_LoadErrors(t *testing.T) {
	testCases := []struct {
		name        string
		globals     map[string]interface{}
		preloadMods []starlet.ModuleName
		allowMods   []starlet.ModuleName
		code        string
		expectedErr string
	}{
		// for globals
		{
			name:        "Missed Globals Variable",
			globals:     map[string]interface{}{"a": 2},
			code:        `b = c * 10`,
			expectedErr: `starlet: exec: test.star:1:5: undefined: c`,
		},
		{
			name:        "Wrong Type Globals Variable",
			globals:     map[string]interface{}{"a": 2},
			code:        `b = a + "10"`,
			expectedErr: `starlet: exec: unknown binary op: int + string`,
		},
		// for preload modules
		{
			name:        "Missed Preload Modules",
			preloadMods: []starlet.ModuleName{},
			code:        `a = nil == None`,
			expectedErr: `starlet: exec: test.star:1:5: undefined: nil`,
		},
		{
			name:        "NonExist Preload Modules",
			preloadMods: []starlet.ModuleName{"nonexist"},
			code:        `a = nil == None`,
			expectedErr: `starlet: preload modules: load module "nonexist": module not found`,
		},
		// for allow modules
		{
			name:        "Missed load() for Builtin Modules",
			allowMods:   []starlet.ModuleName{starlet.ModuleGoIdiomatic},
			code:        `a = nil == None`,
			expectedErr: `starlet: exec: test.star:1:5: undefined: nil`,
		},
		{
			name:        "NonExist Builtin Modules",
			allowMods:   []starlet.ModuleName{"nonexist"},
			code:        `load("nonexist", "nil"); a = nil == None`,
			expectedErr: `starlet: exec: cannot load nonexist: no file system given`,
		},
		{
			name:        "NonExist Function Builtin Modules",
			allowMods:   []starlet.ModuleName{starlet.ModuleGoIdiomatic},
			code:        `load("go_idiomatic", "fake"); a = fake == None`,
			expectedErr: `starlet: exec: load: name fake not found in module go_idiomatic`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := starlet.NewMachine(tc.globals, tc.preloadMods, tc.allowMods)
			m.SetScript("test.star", []byte(tc.code), nil)
			_, err := m.Run(context.Background())
			expectErr(t, err, tc.expectedErr)
		})
	}
}
