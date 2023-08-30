package starlet_test

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/1set/starlet"
	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
)

func Test_DefaultMachine_Run_NoCode(t *testing.T) {
	m := starlet.NewDefault()
	// run with empty script
	_, err := m.Run()
	expectErr(t, err, `starlet: run: no script to execute`)
}

func Test_DefaultMachine_Run_NoSpecificFile(t *testing.T) {
	m := starlet.NewDefault()
	m.SetScript("", nil, os.DirFS("testdata"))
	// run with no specific file name
	_, err := m.Run()
	expectErr(t, err, `starlet: run: no script name`)
}

func Test_DefaultMachine_Run_APlusB(t *testing.T) {
	m := starlet.NewDefault()
	code := `a = 1 + 2`
	m.SetScript("", []byte(code), nil)
	out, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["a"] != int64(3) {
		t.Errorf("unexpected output: %v", out)
	}
}

func Test_DefaultMachine_Run_HelloWorld(t *testing.T) {
	m := starlet.NewDefault()
	// set print function
	printFunc, cmpFunc := getPrintCompareFunc(t)
	m.SetPrintFunc(printFunc)
	// set code
	code := `print("Aloha, Honua!")`
	m.SetScript("aloha.star", []byte(code), nil)
	// run
	_, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// compare
	cmpFunc("Aloha, Honua!\n")
}

func Test_DefaultMachine_Run_File(t *testing.T) {
	m := starlet.NewDefault()
	// set print function
	printFunc, cmpFunc := getPrintCompareFunc(t)
	m.SetPrintFunc(printFunc)
	// run
	_, err := m.RunFile("aloha.star", os.DirFS("testdata"), nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// compare
	cmpFunc("Aloha, Honua!\n")
}

func Test_DefaultMachine_Run_Script(t *testing.T) {
	m := starlet.NewDefault()
	// set print function
	printFunc, cmpFunc := getPrintCompareFunc(t)
	m.SetPrintFunc(printFunc)
	// run
	code := `
print(text)
`
	_, err := m.RunScript([]byte(code), map[string]interface{}{"text": "Hello, World!"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// compare
	cmpFunc("Hello, World!\n")
}

func Test_DefaultMachine_Run_LocalFile(t *testing.T) {
	m := starlet.NewDefault()
	// set print function
	printFunc, cmpFunc := getPrintCompareFunc(t)
	m.SetPrintFunc(printFunc)
	// set code
	m.SetScript("aloha.star", nil, os.DirFS("testdata"))
	// run
	_, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// compare
	cmpFunc("Aloha, Honua!\n")
}

func Test_DefaultMachine_Run_LocalFileNonExist(t *testing.T) {
	m := starlet.NewDefault()
	// set code
	m.SetScript("notfound.star", nil, os.DirFS("testdata"))
	// run
	_, err := m.Run()
	if isOnWindows {
		expectErr(t, err, `starlet: run: open`, `The system cannot find the file specified.`)
	} else {
		expectErr(t, err, `starlet: run: open`, `: no such file or directory`)
	}
}

func Test_DefaultMachine_Run_FSNonExist(t *testing.T) {
	m := starlet.NewDefault()
	// set code
	m.SetScript("aloha.star", nil, os.DirFS("not-found-dir"))
	// run
	_, err := m.Run()
	if isOnWindows {
		expectErr(t, err, `starlet: run: open`, `The system cannot find the path specified.`)
	} else {
		expectErr(t, err, `starlet: run: open`, `: no such file or directory`)
	}
}

func Test_DefaultMachine_Run_InvalidGlobals(t *testing.T) {
	m := starlet.NewDefault()
	// set invalid globals
	m.SetGlobals(map[string]interface{}{
		"a": make(chan int),
	})
	// set code
	m.SetScript("test.star", []byte(`a = 1`), nil)
	// run
	_, err := m.Run()
	expectErr(t, err, `starlight: convert globals: type chan int is not a supported starlark type`)
}

func Test_DefaultMachine_Run_InvalidExtras(t *testing.T) {
	m := starlet.NewDefault()
	_, err := m.RunScript([]byte(`a = 1`), map[string]interface{}{
		"a": make(chan int),
	})
	expectErr(t, err, `starlight: convert extras: type chan int is not a supported starlark type`)
}

func Test_DefaultMachine_Run_LoadFunc(t *testing.T) {
	m := starlet.NewDefault()
	// set code
	code := `load("fibonacci.star", "fibonacci"); val = fibonacci(10)[-1]`
	m.SetScript("test.star", []byte(code), os.DirFS("testdata"))
	// run
	out, err := m.Run()
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

func Test_DefaultMachine_Run_LoadNonExist(t *testing.T) {
	m := starlet.NewDefault()
	// set code
	code := `load("nonexist.star", "a")`
	m.SetScript("test.star", []byte(code), os.DirFS("testdata"))
	// run
	_, err := m.Run()
	// check result
	if isOnWindows {
		expectErr(t, err, `starlark: exec: cannot load nonexist.star: open`, `The system cannot find the file specified.`)
	} else {
		expectErr(t, err, `starlark: exec: cannot load nonexist.star: open`, `: no such file or directory`)
	}
}

func Test_Machine_Run_Globals(t *testing.T) {
	sm := starlark.NewDict(1)
	_ = sm.SetKey(starlark.String("bee"), starlark.MakeInt(2))
	m := starlet.NewWithNames(map[string]interface{}{
		"a": 2,
		"c": sm,
	}, nil, nil)
	// set code
	code := `
c = 100
b = a * 10 + c
`
	m.SetScript("test.star", []byte(code), nil)
	// run
	out, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["b"].(int64) != int64(120) {
		t.Errorf("unexpected output: %v", out)
	}
}

func Test_Machine_Run_Override(t *testing.T) {
	getValLoadList := func(x int64) starlet.ModuleLoaderList {
		return starlet.ModuleLoaderList{
			starlet.MakeModuleLoaderFromMap(map[string]interface{}{
				"x": x,
			}),
		}
	}
	getValLoadMap := func(x int64) starlet.ModuleLoaderMap {
		return starlet.ModuleLoaderMap{
			"number": starlet.MakeModuleLoaderFromMap(map[string]interface{}{
				"x": x,
			}),
		}
	}
	testCases := []struct {
		name      string
		setFunc   func(m *starlet.Machine)
		extras    starlet.StringAnyMap
		code      string
		expectVal int64
	}{
		{
			name:      "Runtime",
			code:      `val = 100`,
			expectVal: 100,
		},
		{
			name: "Globals",
			setFunc: func(m *starlet.Machine) {
				m.SetGlobals(map[string]interface{}{
					"x": 200,
				})
			},
			code:      `val = x`,
			expectVal: 200,
		},
		{
			name: "Preload",
			setFunc: func(m *starlet.Machine) {
				m.SetPreloadModules(getValLoadList(300))
			},
			code:      `val = x`,
			expectVal: 300,
		},
		{
			name: "Extras",
			extras: starlet.StringAnyMap{
				"x": 400,
			},
			code:      `val = x`,
			expectVal: 400,
		},
		{
			name: "LazyLoad",
			setFunc: func(m *starlet.Machine) {
				m.SetLazyloadModules(getValLoadMap(500))
			},
			code:      `load("number", "x"); val = x`,
			expectVal: 500,
		},
		{
			name: "Globals and Preload",
			setFunc: func(m *starlet.Machine) {
				m.SetGlobals(map[string]interface{}{
					"x": 200,
				})
				m.SetPreloadModules(getValLoadList(300))
			},
			code:      `val = x`,
			expectVal: 300,
		},
		{
			name: "Globals and Preload and Extras",
			setFunc: func(m *starlet.Machine) {
				m.SetGlobals(map[string]interface{}{
					"x": 200,
				})
				m.SetPreloadModules(getValLoadList(300))
			},
			extras:    map[string]interface{}{"x": 400},
			code:      `val = x`,
			expectVal: 400,
		},
		{
			name: "Globals and Preload and Extras and LazyLoad",
			setFunc: func(m *starlet.Machine) {
				m.SetGlobals(map[string]interface{}{
					"x": 200,
				})
				m.SetPreloadModules(getValLoadList(300))
				m.SetLazyloadModules(getValLoadMap(500))
			},
			extras:    map[string]interface{}{"x": 400},
			code:      `load("number", "x"); val = x`,
			expectVal: 500,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// prepare machine
			m := starlet.NewDefault()
			m.SetPrintFunc(getLogPrintFunc(t))
			if tc.setFunc != nil {
				tc.setFunc(m)
			}

			// run script
			out, err := m.RunScript([]byte(tc.code), tc.extras)

			// check result
			if err != nil {
				t.Errorf("unexpected error for %s: %v", tc.name, err)
				return
			}
			if out == nil {
				t.Errorf("unexpected nil output for %s", tc.name)
			} else if out["val"].(int64) != tc.expectVal {
				t.Errorf("unexpected output for %s: %v, want: %d", tc.name, out, tc.expectVal)
			}
		})
	}
}

func Test_Machine_Run_Exit_Quit(t *testing.T) {
	m := starlet.NewDefault()
	m.SetPreloadModules(starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic")})

	// test exit()
	m.SetScript("test.star", []byte(`
a = 1
exit()
b = 2
`), nil)
	out, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["a"].(int64) != int64(1) {
		t.Errorf("unexpected output: %v", out)
	} else if _, ok := out["b"]; ok {
		t.Errorf("unexpected output: %v", out)
	}

	// test quit()
	m.SetScript("test.star", []byte(`
c = 3
quit()
d = 4
`), nil)
	out, err = m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["c"].(int64) != int64(3) {
		t.Errorf("unexpected output: %v", out)
	} else if _, ok := out["d"]; ok {
		t.Errorf("unexpected output: %v", out)
	}

	// test exit(1)
	m.SetScript("test.star", []byte(`
e = 5
exit(1)
f = 6
`), nil)
	out, err = m.Run()
	expectErr(t, err, `starlet: run: exit code: 1`)
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["e"].(int64) != int64(5) {
		t.Errorf("unexpected output: %v", out)
	} else if _, ok := out["f"]; ok {
		t.Errorf("unexpected output: %v", out)
	}

	// test quit(2)
	m.SetScript("test.star", []byte(`
g = 7
quit(2)
h = 8
`), nil)
	out, err = m.Run()
	expectErr(t, err, `starlet: run: exit code: 2`)
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["g"].(int64) != int64(7) {
		t.Errorf("unexpected output: %v", out)
	} else if _, ok := out["h"]; ok {
		t.Errorf("unexpected output: %v", out)
	}
}

func Test_Machine_Run_File_Globals(t *testing.T) {
	m := starlet.NewWithNames(map[string]interface{}{
		"magic_number": 30,
	}, nil, nil)
	// set code
	m.SetScript("magic.star", nil, os.DirFS("testdata"))
	// run
	out, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if f, ok := out["custom"]; !ok {
		t.Errorf("got no func, unexpected output: %v", out)
	} else if fn, ok := f.(*starlark.Function); !ok {
		t.Errorf("unexpected output: %v", out)
	} else {
		res, err := starlark.Call(&starlark.Thread{Name: "afterparty"}, fn, nil, nil)
		if err != nil {
			t.Errorf("unexpected call error: %v", err)
		} else if r, ok := res.(starlark.String); !ok {
			t.Errorf("unexpected call result: %v", res)
		} else if r != "Custom[30]" {
			t.Errorf("unexpected call string result: %q", res)
		}
	}
}

func Test_Machine_Run_Load_Use_Globals(t *testing.T) {
	m := starlet.NewWithNames(map[string]interface{}{
		"magic_number": 50,
	}, nil, nil)
	// set code
	code := `load("magic.star", "custom"); val = custom()`
	m.SetScript("dummy.star", []byte(code), os.DirFS("testdata"))
	// run
	out, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if f, ok := out["val"]; !ok {
		t.Errorf("got no value, unexpected output: %v", out)
	} else if f != "Custom[50]" {
		t.Errorf("unexpected output: %v", out)
	}
}

func Test_Machine_Run_File_Missing_Globals(t *testing.T) {
	m := starlet.NewWithNames(map[string]interface{}{
		"other_number": 30,
	}, nil, nil)
	// set code
	m.SetScript("magic.star", nil, os.DirFS("testdata"))
	// run
	_, err := m.Run()
	expectErr(t, err, `starlark: exec: magic.star:5:32: undefined: magic_number`)
}

func Test_Machine_Run_PreloadModules(t *testing.T) {
	m := starlet.NewWithNames(nil, []string{"go_idiomatic"}, nil)
	// set code
	code := `a = nil == None`
	m.SetScript("test.star", []byte(code), nil)
	// run
	out, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["a"].(bool) != true {
		t.Errorf("unexpected output: %v", out)
	}
}

func Test_Machine_Run_LazyloadModules(t *testing.T) {
	m := starlet.NewWithNames(nil, nil, []string{"go_idiomatic", "json"})
	// set code
	code := `
load("go_idiomatic", "nil")
load("json", "encode")
a = nil == None
`
	m.SetScript("test.star", []byte(code), nil)
	// run
	out, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["a"].(bool) != true {
		t.Errorf("unexpected output: %v", out)
	}
}

func Test_Machine_Run_LazyLoad_Override_Globals(t *testing.T) {
	// create machine
	m := starlet.NewWithNames(map[string]interface{}{"fibonacci": 123}, nil, nil)
	m.EnableGlobalReassign() // enable global reassign only for this test, if it's not enabled, it will fail for: local variable fibonacci referenced before assignment
	// set code
	code := `
x = fibonacci * 2
load("fibonacci.star", "fibonacci")
load("fibonacci.star", fib="fibonacci")
y = fibonacci(10)[-1]
z = fib(10)[-1]
`
	m.SetScript("test.star", []byte(code), os.DirFS("testdata"))
	// run
	out, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// check result
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["x"] != int64(246) || out["y"] != int64(55) || out["z"] != int64(55) {
		t.Errorf("unexpected output: %v", out)
	}
}

func Test_Machine_Run_Override_Globals(t *testing.T) {
	// create machine
	m := starlet.NewWithNames(map[string]interface{}{"num": 10}, nil, nil)
	// set code
	code := `
x = num * 2
load("fibonacci.star", "fibonacci")
load("fibonacci.star", fib="fibonacci")
y = fibonacci(num)[-1]
z = fib(num)[-1]
`
	m.SetScript("test.star", []byte(code), os.DirFS("testdata"))
	// run
	out, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// check result
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["x"] != int64(20) || out["y"] != int64(55) || out["z"] != int64(55) {
		t.Errorf("unexpected output: %v", out)
	}
}

func Test_Machine_Run_PreLoad_Override_Globals(t *testing.T) {
	// create machine
	m := starlet.NewWithLoaders(map[string]interface{}{"num": 10}, starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("coins.star", os.DirFS("testdata"), nil)}, nil)
	m.EnableGlobalReassign() // enable global reassign only for this test, if it's not enabled, it will fail for: local variable coins referenced before assignment
	// set code
	code := `
num = 100
x = num * 5 + coins['quarter']
coins = 50
`
	m.SetScript("test.star", []byte(code), os.DirFS("testdata"))
	// run
	out, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// check result
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if len(out) != 3 {
		t.Errorf("unexpected output: %v", out)
	} else if out["x"] != int64(525) || out["num"] != int64(100) || out["coins"] != int64(50) {
		t.Errorf("unexpected output: %v", out)
	}

	// disable global reassign to test that it fails
	m.DisableGlobalReassign()
	_, err = m.Run()
	expectErr(t, err, `starlark: exec: global variable coins referenced before assignment`)
}

func Test_Machine_Run_LoadCycle(t *testing.T) {
	m := starlet.NewDefault()
	// set code1
	m.SetScript("circle1.star", nil, os.DirFS("testdata"))
	_, err := m.Run()
	expectErr(t, err, `starlark: exec: cannot load circle2.star: cannot load circle1.star: cannot load circle2.star: cycle in load graph`)
	// set code2
	m.SetScript("circle2.star", nil, os.DirFS("testdata"))
	_, err = m.Run()
	expectErr(t, err, `starlark: exec: cannot load circle1.star: cannot load circle2.star: cycle in load graph`)
}

func Test_Machine_Run_Recursion(t *testing.T) {
	// create machine
	m := starlet.NewDefault()
	m.EnableRecursionSupport() // enable recursion support only for this test, if it's not enabled, it will fail for: function fib called recursively
	// set code
	code := `
def fib(n):
	if n < 2:
		return n
	return fib(n-1) + fib(n-2)
x = fib(10)
`
	m.SetScript("test.star", []byte(code), nil)
	// run
	out, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// check result
	if out == nil {
		t.Errorf("unexpected nil output")
	} else if out["x"] != int64(55) {
		t.Errorf("unexpected output: %v", out)
	}

	// disable to test recursion error
	m.DisableRecursionSupport()
	_, err = m.Run()
	expectErr(t, err, `starlark: exec: function fib called recursively`)
}

func Test_Machine_Run_LoadErrors(t *testing.T) {
	mm := starlark.NewDict(1)
	_ = mm.SetKey(starlark.String("quarter"), starlark.MakeInt(100))
	_ = mm.SetKey(starlark.String("dime"), starlark.MakeInt(10))
	testFS := os.DirFS("testdata")
	nonExistFS := os.DirFS("nonexist")
	testCases := []struct {
		name          string
		globals       map[string]interface{}
		preloadMods   []string
		lazyMods      []string
		code          string
		modFS         fs.FS
		expectedErr   string
		expectedPanic bool
	}{
		// for globals
		{
			name:        "Unsupported Globals Type",
			globals:     map[string]interface{}{"a": make(chan int)},
			code:        `b = a`,
			expectedErr: `starlight: convert globals: type chan int is not a supported starlark type`,
		},
		{
			name:        "Missed Globals Variable",
			globals:     map[string]interface{}{"a": 2},
			code:        `b = c * 10`,
			expectedErr: `starlark: exec: test.star:1:5: undefined: c`,
		},
		{
			name:        "Wrong Type Globals Variable",
			globals:     map[string]interface{}{"a": 2},
			code:        `b = a + "10"`,
			expectedErr: `starlark: exec: unknown binary op: int + string`,
		},
		{
			name:    "Fails to Override Globals Variable",
			globals: map[string]interface{}{"coins": mm},
			code: `
num = 100
x = num * 5 + coins['quarter']
coins = 50
`,
			expectedErr: `starlark: exec: global variable coins referenced before assignment`,
		},
		// for preload modules
		{
			name:        "Missed Preload Modules",
			preloadMods: []string{},
			code:        `a = nil == None`,
			expectedErr: `starlark: exec: test.star:1:5: undefined: nil`,
		},
		{
			name:          "NonExist Preload Modules",
			preloadMods:   []string{"nonexist"},
			code:          `a = nil == None`,
			expectedErr:   `starlet: make: module not found: nonexist`,
			expectedPanic: true,
		},
		// for lazyload modules
		{
			name:        "Missed load() for LazyLoad Modules",
			lazyMods:    []string{"go_idiomatic"},
			code:        `a = nil == None`,
			expectedErr: `starlark: exec: test.star:1:5: undefined: nil`,
		},
		{
			name:          "NonExist LazyLoad Modules",
			lazyMods:      []string{"nonexist"},
			code:          `load("nonexist", "nil"); a = nil == None`,
			expectedErr:   `starlet: make: module not found: nonexist`,
			expectedPanic: true,
		},
		{
			name:        "NonExist Function in LazyLoad Modules",
			lazyMods:    []string{"go_idiomatic"},
			code:        `load("go_idiomatic", "fake"); a = fake == None`,
			expectedErr: `starlark: exec: load: name fake not found in module go_idiomatic`,
		},
		// for load fs --- user modules
		{
			name:        "No FS for User Modules",
			code:        `load("fibonacci.star", "fibonacci"); val = fibonacci(10)[-1]`,
			expectedErr: `starlark: exec: cannot load fibonacci.star: no file system given`,
		},
		{
			name:        "Missed load() for User Modules",
			code:        `val = fibonacci(10)[-1]`,
			modFS:       testFS,
			expectedErr: `starlark: exec: test.star:1:7: undefined: fibonacci`,
		},
		{
			name:        "Duplicate load() for User Modules",
			code:        `load("fibonacci.star", "fibonacci"); load("fibonacci.star", "fibonacci"); val = fibonacci(10)[-1]`,
			modFS:       testFS,
			expectedErr: `starlark: exec: test.star:1:62: cannot reassign top-level fibonacci`,
		},
		{
			name:        "NonExist User Modules",
			code:        `load("nonexist.star", "fibonacci"); val = fibonacci(10)[-1]`,
			modFS:       testFS,
			expectedErr: `starlark: exec: cannot load nonexist.star: open`,
		},
		{
			name:        "NonExist File System",
			code:        `load("fibonacci.star", "fibonacci"); val = fibonacci(10)[-1]`,
			modFS:       nonExistFS,
			expectedErr: `starlark: exec: cannot load fibonacci.star: open`,
		},
		{
			name:        "NonExist Function in User Modules",
			code:        `load("fibonacci.star", "fake"); val = fake(10)[-1]`,
			modFS:       testFS,
			expectedErr: `starlark: exec: load: name fake not found in module fibonacci.star`,
		},
		{
			name:        "Existing and NonExist Functions in User Modules",
			code:        `load("fibonacci.star", "fibonacci", "fake"); val = fibonacci(10)[-1]`,
			modFS:       testFS,
			expectedErr: `starlark: exec: load: name fake not found in module fibonacci.star`,
		},
		// for globals + user modules
		{
			name:    "User Modules Override Globals",
			globals: map[string]interface{}{"fibonacci": 2},
			code:    `load("fibonacci.star", "fibonacci"); val = fibonacci(10)[-1]; print(x, val)`,
			modFS:   testFS,
		},
		{
			name:        "User Modules Fail to Override Globals",
			globals:     map[string]interface{}{"fibonacci": 2},
			code:        `x = fibonacci * 10; load("fibonacci.star", "fibonacci"); val = fibonacci(10)[-1]; print(x, val)`,
			modFS:       testFS,
			expectedErr: `starlark: exec: local variable fibonacci referenced before assignment`,
			// NOTE: for this behavior: read the comments before `r.useToplevel(use)` in `func (r *resolver) use(id *syntax.Ident)` in file: go.starlark.net@v0.0.0-20230525235612-a134d8f9ddca/resolve/resolve.go
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					err := r.(error)
					if !tc.expectedPanic {
						t.Errorf("Unexpected panic: %v", err)
					} else {
						expectErr(t, err, tc.expectedErr)
					}
				} else if tc.expectedPanic {
					t.Errorf("Expected panic, but got none")
				}
			}()

			m := starlet.NewWithNames(tc.globals, tc.preloadMods, tc.lazyMods)
			m.SetPrintFunc(getLogPrintFunc(t))
			m.SetScript("test.star", []byte(tc.code), tc.modFS)
			_, err := m.Run()
			expectErr(t, err, tc.expectedErr)
		})
	}
}

func Test_Machine_Run_FileLoaders(t *testing.T) {
	// the following tests tests the combination of preload and lazyload modules loaded via file.
	var (
		testFS     = os.DirFS("testdata")
		nonExistFS = os.DirFS("nonexist")
	)
	testCases := []struct {
		name        string
		globals     map[string]interface{}
		preList     starlet.ModuleLoaderList
		lazyMap     starlet.ModuleLoaderMap
		code        string
		modFS       fs.FS
		expectedErr string
		cmpResult   func(val interface{}) bool
	}{
		// for preload with fs
		{
			name:        "No FS for Preload Modules",
			preList:     starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("fibonacci.star", nil, nil)},
			code:        `val = fibonacci(10)[-1]`,
			expectedErr: "starlet: load: no file system given",
		},
		{
			name:        "NonExist file system for Preload Modules",
			preList:     starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("fibonacci.star", nonExistFS, nil)},
			code:        `val = fibonacci(10)[-1]`,
			expectedErr: "starlet: load: open ",
		},
		{
			name:        "NonExist file for Preload Modules",
			preList:     starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("nonexist.star", testFS, nil)},
			code:        `val = fibonacci(10)[-1]`,
			expectedErr: "starlet: load: open ",
		},
		{
			name:    "Single File for Preload Modules",
			preList: starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("fibonacci.star", testFS, nil)},
			code:    `val = fibonacci(10)[-1]`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(55)
			},
		},
		{
			name:    "Duplicate Files for Preload Modules",
			preList: starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("fibonacci.star", testFS, nil), starlet.MakeModuleLoaderFromFile("fibonacci.star", testFS, nil)},
			code:    `val = fibonacci(10)[-1]`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(55)
			},
		},
		{
			name:    "Multiple Files for Preload Modules",
			preList: starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("fibonacci.star", testFS, nil), starlet.MakeModuleLoaderFromFile("factorial.star", testFS, nil)},
			code:    `val = fibonacci(10)[-1] + factorial(10)`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(3628855)
			},
		},
		{
			name:        "Preload Modules Requires External Value",
			globals:     map[string]interface{}{"input": 10},
			preList:     starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("one.star", testFS, nil)},
			code:        `val = number`,
			expectedErr: `starlet: load: one.star:1:10: undefined: input`,
		},
		{
			name:      "Preload Modules With External Value",
			preList:   starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("one.star", testFS, map[string]starlark.Value{"input": starlark.MakeInt(5)})},
			code:      `val = number`,
			cmpResult: func(val interface{}) bool { return val.(int64) == int64(500) },
		},
		{
			name:    "Override Global Variables",
			globals: map[string]interface{}{"num": 10},
			preList: starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("coins.star", testFS, nil)},
			code: `
num = 100
val = num * 5 + coins['quarter']
`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(525)
			},
		},
		{
			name:    "Fails to Override Preload Modules",
			globals: map[string]interface{}{"num": 10},
			preList: starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("coins.star", testFS, nil)},
			code: `
num = 100
x = num * 5 + coins['quarter']
coins = 50
`,
			expectedErr: `starlark: exec: global variable coins referenced before assignment`,
		},
		// for lazyload with fs
		{
			name:        "Missing Lazyload Modules",
			lazyMap:     starlet.ModuleLoaderMap{},
			code:        `load("fib", "fibonacci"); val = fibonacci(10)[-1]`,
			expectedErr: `starlark: exec: cannot load fib: no file system given`,
		},
		{
			name:        "Missing Lazyload Modules with Invalid FS",
			lazyMap:     starlet.ModuleLoaderMap{},
			modFS:       nonExistFS,
			code:        `load("fib", "fibonacci"); val = fibonacci(10)[-1]`,
			expectedErr: `starlark: exec: cannot load fib: open `,
		},
		{
			name:        "Missing Lazyload Modules with Valid FS",
			lazyMap:     starlet.ModuleLoaderMap{},
			modFS:       testFS,
			code:        `load("fib", "fibonacci"); val = fibonacci(10)[-1]`,
			expectedErr: `starlark: exec: cannot load fib: open `,
		},
		{
			name:        "No FS for Lazyload Modules",
			lazyMap:     starlet.ModuleLoaderMap{"fib": starlet.MakeModuleLoaderFromFile("fibonacci.star", nil, nil)},
			code:        `load("fib", "fibonacci"); val = fibonacci(10)[-1]`,
			expectedErr: `starlark: exec: cannot load fib: no file system given`,
		},
		{
			name:        "NonExist file system for Lazyload Modules",
			lazyMap:     starlet.ModuleLoaderMap{"fib": starlet.MakeModuleLoaderFromFile("fibonacci.star", nonExistFS, nil)},
			code:        `load("fib", "fibonacci"); val = fibonacci(10)[-1]`,
			expectedErr: `starlark: exec: cannot load fib: open `,
		},
		{
			name:        "NonExist file for Lazyload Modules",
			lazyMap:     starlet.ModuleLoaderMap{"fib": starlet.MakeModuleLoaderFromFile("nonexist.star", testFS, nil)},
			code:        `load("fib", "fibonacci"); val = fibonacci(10)[-1]`,
			expectedErr: `starlark: exec: cannot load fib: open `,
		},
		{
			name:    "Single File for Lazyload Modules",
			lazyMap: starlet.ModuleLoaderMap{"fib": starlet.MakeModuleLoaderFromFile("fibonacci.star", testFS, nil)},
			code:    `load("fib", "fibonacci"); val = fibonacci(10)[-1]`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(55)
			},
		},
		{
			name:        "Duplicate Files for Lazyload Modules",
			lazyMap:     starlet.ModuleLoaderMap{"fib": starlet.MakeModuleLoaderFromFile("fibonacci.star", testFS, nil), "fib2": starlet.MakeModuleLoaderFromFile("fibonacci.star", testFS, nil)},
			code:        `load("fib", "fibonacci"); load("fib2", "fibonacci"); val = fibonacci(10)[-1]`,
			expectedErr: `starlark: exec: test.star:1:41: cannot reassign top-level fibonacci`,
		},
		{
			name:    "Multiple Files for Lazyload Modules",
			lazyMap: starlet.ModuleLoaderMap{"fib": starlet.MakeModuleLoaderFromFile("fibonacci.star", testFS, nil), "fac": starlet.MakeModuleLoaderFromFile("factorial.star", testFS, nil)},
			code:    `load("fib", "fibonacci"); load("fac", "factorial"); val = fibonacci(10)[-1] + factorial(10)`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(3628855)
			},
		},
		{
			name:        "Lazyload Modules Requires External Value",
			globals:     map[string]interface{}{"input": 10},
			lazyMap:     starlet.ModuleLoaderMap{"one": starlet.MakeModuleLoaderFromFile("one.star", testFS, nil)},
			code:        `load("one", "number"); val = number`,
			expectedErr: `starlark: exec: cannot load one: one.star:1:10: undefined: input`,
		},
		{
			name:        "Lazyload Modules Misplaced External Value",
			lazyMap:     starlet.ModuleLoaderMap{"one": starlet.MakeModuleLoaderFromFile("one.star", testFS, nil)},
			code:        `input = 10; load("one", "number"); val = number`,
			expectedErr: `starlark: exec: cannot load one: one.star:1:10: undefined: input`,
		},
		{
			name:    "Lazyload Modules With External Value",
			lazyMap: starlet.ModuleLoaderMap{"one": starlet.MakeModuleLoaderFromFile("one.star", testFS, map[string]starlark.Value{"input": starlark.MakeInt(10)})},
			code:    `load("one", "number"); val = number`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(1000)
			},
		},
		// both preload and lazyload
		{
			name:    "Duplicate Files for Preload and Lazyload Modules",
			preList: starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("fibonacci.star", testFS, nil)},
			lazyMap: starlet.ModuleLoaderMap{"fib": starlet.MakeModuleLoaderFromFile("fibonacci.star", testFS, nil)},
			code:    `load("fib", "fibonacci"); val = fibonacci(10)[-1]`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(55)
			},
		},
		{
			name:    "Multiple Files for Preload and Lazyload Modules",
			preList: starlet.ModuleLoaderList{starlet.MakeModuleLoaderFromFile("fibonacci.star", testFS, nil)},
			lazyMap: starlet.ModuleLoaderMap{"fac": starlet.MakeModuleLoaderFromFile("factorial.star", testFS, nil)},
			code:    `load("fac", "factorial"); val = fibonacci(10)[-1] + factorial(10)`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(3628855)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := starlet.NewWithLoaders(tc.globals, tc.preList, tc.lazyMap)
			m.SetPrintFunc(getLogPrintFunc(t))
			m.SetScript("test.star", []byte(tc.code), tc.modFS)
			out, err := m.Run()

			// check result
			if tc.expectedErr != "" {
				expectErr(t, err, tc.expectedErr)
				return
			} else if err != nil {
				t.Errorf("Expected no errors, got error: %v", err)
			}

			if tc.cmpResult != nil {
				if out == nil {
					t.Errorf("Unexpected empty result: %v", out)
				} else if v, ok := out["val"]; !ok {
					t.Errorf("Unexpected missing result: %v", out)
				} else if !tc.cmpResult(v) {
					t.Errorf("Unexpected result: %v", out)
				}
			}
		})
	}
}

func Test_Machine_Run_CodeLoaders(t *testing.T) {
	// the following tests tests the combination of direct preload and lazyload modules in code, and validate the override behavior.
	var (
		testFS                 = os.DirFS("testdata")
		failName, failLoader   = getErrorModuleLoader()
		appleName, appleLoader = getAppleModuleLoader()
		berryName, berryLoader = getBlueberryModuleLoader()
		cocoName, cocoLoader   = getCoconutModuleLoader()
	)
	testCases := []struct {
		name        string
		globals     map[string]interface{}
		preList     starlet.ModuleLoaderList
		lazyMap     starlet.ModuleLoaderMap
		code        string
		modFS       fs.FS
		expectedErr string
		cmpResult   func(val interface{}) bool
	}{
		// no loaders
		{
			name:    "Nil Loaders",
			globals: nil,
			preList: nil,
			lazyMap: nil,
			code:    `val = 1 + 2`,
			modFS:   testFS,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 3
			},
		},
		{
			name:    "Empty Loaders",
			globals: map[string]interface{}{},
			preList: starlet.ModuleLoaderList{},
			lazyMap: starlet.ModuleLoaderMap{},
			code:    `val = 3 + 4`,
			modFS:   testFS,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 7
			},
		},
		{
			name:        "Nil Loader List Element",
			preList:     starlet.ModuleLoaderList{nil},
			code:        `val = 4 + 5`,
			modFS:       testFS,
			expectedErr: `starlet: load: nil module loader`,
		},
		{
			name:        "Nil Loader Map Element",
			lazyMap:     starlet.ModuleLoaderMap{"nil_loader": nil},
			code:        `load("nil_loader", "num"); val = 5 + 6`,
			modFS:       testFS,
			expectedErr: `starlark: exec: cannot load nil_loader: nil module loader`,
		},
		// only pre loaders
		{
			name:    "Preload Module: Go",
			preList: starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic")},
			code:    `val = nil != true`,
			cmpResult: func(val interface{}) bool {
				return val.(bool) == true
			},
		},
		{
			name:        "Preload Module Fails",
			preList:     starlet.ModuleLoaderList{failLoader},
			code:        `val = 1 + 2`,
			modFS:       testFS,
			expectedErr: `starlet: load: invalid module loader`,
		},
		{
			name:    "Preload Module Untouched",
			preList: starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic")},
			code:    `val = 1 + 2`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 3
			},
		},
		{
			name:    "Multiple Preload Modules",
			preList: starlet.ModuleLoaderList{appleLoader, berryLoader, cocoLoader},
			code:    `val = apple + blueberry + coconut`,
			cmpResult: func(val interface{}) bool {
				return val.(string) == `🍎🫐🥥`
			},
		},
		{
			name:    "Duplicate Preload Modules",
			preList: starlet.ModuleLoaderList{appleLoader, appleLoader},
			code:    `val = apple + apple`,
			cmpResult: func(val interface{}) bool {
				return val.(string) == `🍎🍎`
			},
		},
		{
			name:    "Shadowed Preload Modules",
			preList: starlet.ModuleLoaderList{appleLoader, berryLoader},
			code:    `val = number`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 20
			},
		},
		{
			name:    "More Shadowed Preload Modules",
			preList: starlet.ModuleLoaderList{appleLoader, berryLoader, cocoLoader},
			code:    `val = number`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 40
			},
		},
		{
			name:    "Override Global Variables With Preload Modules",
			globals: map[string]interface{}{"num": 10},
			preList: starlet.ModuleLoaderList{appleLoader, berryLoader, cocoLoader},
			code: `
num = 100
val = num * 5 + number
`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(540)
			},
		},
		{
			name:    "Fails to Override Preload Modules",
			globals: map[string]interface{}{"num": 10},
			preList: starlet.ModuleLoaderList{appleLoader, berryLoader, cocoLoader},
			code: `
num = 100
val = num * 5 + number
number = 500
`,
			expectedErr: `starlark: exec: global variable number referenced before assignment`,
		},
		// only lazy loaders
		{
			name:    "LazyLoad Module: Go",
			lazyMap: starlet.ModuleLoaderMap{"gogo": starlet.GetBuiltinModule("go_idiomatic")},
			code:    `load("gogo", "nil", "true"); val = nil != true`,
			cmpResult: func(val interface{}) bool {
				return val.(bool) == true
			},
		},
		{
			name:        "Invalid LazyLoad Module Fails",
			lazyMap:     starlet.ModuleLoaderMap{failName: failLoader},
			code:        fmt.Sprintf(`load(%q, "nil", "true"); val = nil != true`, failName),
			expectedErr: fmt.Sprintf(`starlark: exec: cannot load %s: invalid module loader`, failName),
		},
		{
			name:    "Invalid LazyLoad Module Untouched",
			lazyMap: starlet.ModuleLoaderMap{failName: failLoader},
			code:    `val = 2 * 10`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 20
			},
		},
		{
			name:    "Multiple LazyLoad Modules",
			lazyMap: starlet.ModuleLoaderMap{appleName: appleLoader, berryName: berryLoader, cocoName: cocoLoader},
			code: `
load("mock_apple", "apple")
load("mock_blueberry", berry="blueberry")
load("mock_coconut", coco="coconut")
val = apple + berry + coco
`,
			cmpResult: func(val interface{}) bool {
				return val.(string) == `🍎🫐🥥`
			},
		},
		{
			name:    "Shadowed LazyLoad Modules",
			lazyMap: starlet.ModuleLoaderMap{appleName: appleLoader, berryName: berryLoader},
			code: `
load("mock_apple", "number")
load("mock_blueberry", "number")
val = number
`,
			expectedErr: `starlark: exec: test.star:3:25: cannot reassign top-level number`,
		},
		{
			name:    "Override LazyLoad Modules",
			lazyMap: starlet.ModuleLoaderMap{appleName: appleLoader, berryName: berryLoader},
			code: `
load("mock_apple", "number")
number = 10
val = number * 10
`,
			expectedErr: `starlark: exec: test.star:3:1: cannot reassign local number declared at test.star:2:21`,
		},
		{
			name:    "Override Global Variables With Lazyload Modules",
			globals: map[string]interface{}{"num": 10},
			lazyMap: starlet.ModuleLoaderMap{appleName: appleLoader, berryName: berryLoader},
			code: `
num = 100
load("mock_apple", "number")
val = num * 5 + number
`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(510)
			},
		},
		{
			name:    "Override Global Variables And Lazyload Modules",
			globals: map[string]interface{}{"num": 10},
			lazyMap: starlet.ModuleLoaderMap{appleName: appleLoader, berryName: berryLoader},
			code: `
num = 100
number = 500
load("mock_apple", "number")
val = num * 5 + number
`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == int64(510)
			},
		},
		// both pre and lazy loaders
		{
			name:    "Preload and LazyLoad Same Modules for Same Variable",
			preList: starlet.ModuleLoaderList{appleLoader},
			lazyMap: starlet.ModuleLoaderMap{appleName: appleLoader},
			code:    `load("mock_apple", "number"); val = number`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 10
			},
		},
		{
			name:    "Preload and LazyLoad Same Modules with Different Names",
			preList: starlet.ModuleLoaderList{appleLoader},
			lazyMap: starlet.ModuleLoaderMap{appleName: appleLoader},
			code:    `load("mock_apple", n1="number"); val = number + n1`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 20
			},
		},
		{
			name:    "Preload and LazyLoad Different Modules with Different Names",
			preList: starlet.ModuleLoaderList{appleLoader},
			lazyMap: starlet.ModuleLoaderMap{berryName: berryLoader},
			code:    `load("mock_blueberry", n2="number"); val = number + n2`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 30
			},
		},
		{
			name:    "Preload and LazyLoad Different Modules for Same Variable",
			preList: starlet.ModuleLoaderList{appleLoader},
			lazyMap: starlet.ModuleLoaderMap{berryName: berryLoader},
			code:    `load("mock_blueberry", "number"); val = number`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 20
			},
		},
		{
			name:    "Preload and LazyLoad Same Modules for Same Function",
			preList: starlet.ModuleLoaderList{appleLoader},
			lazyMap: starlet.ModuleLoaderMap{berryName: berryLoader},
			code:    `load("mock_blueberry", "process"); val = process(10)`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 200
			},
		},
		{
			name:    "Preload and LazyLoad Different Modules",
			preList: starlet.ModuleLoaderList{appleLoader},
			lazyMap: starlet.ModuleLoaderMap{berryName: berryLoader, cocoName: cocoLoader},
			code: `
load("mock_blueberry", n2="number")
load("mock_coconut", n3="number")
val = number + n2 + n3
`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 70
			},
		},
		{
			name:    "Preload and LazyLoad Different Modules with Same Name",
			preList: starlet.ModuleLoaderList{appleLoader, berryLoader},
			lazyMap: starlet.ModuleLoaderMap{cocoName: cocoLoader},
			code: `
load("mock_coconut", n3="number")
val = number + n3
`,
			cmpResult: func(val interface{}) bool {
				return val.(int64) == 60
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := starlet.NewWithLoaders(tc.globals, tc.preList, tc.lazyMap)
			m.SetPrintFunc(getLogPrintFunc(t))
			m.SetScript("test.star", []byte(tc.code), tc.modFS)
			out, err := m.Run()

			// check result
			if tc.expectedErr != "" {
				expectErr(t, err, tc.expectedErr)
				return
			} else if err != nil {
				t.Errorf("Expected no errors, got error: %v", err)
			}

			if tc.cmpResult != nil {
				if out == nil {
					t.Errorf("Unexpected empty result: %v", out)
				} else if v, ok := out["val"]; !ok {
					t.Errorf("Unexpected missing result: %v", out)
				} else if !tc.cmpResult(v) {
					t.Errorf("Unexpected result: %v", out)
				}
			}
		})
	}
}

func Test_Machine_Run_Twice(t *testing.T) {
	code1 := `
x = 2
y = 10

load("math", "pow")
z = int(pow(x, y))
print("z =", z)
`
	code2 := `
pow = 10
load("math", p2="pow")
t = p2(2, 5) + pow
print("t =", t, "{}-{}-{}".format(x,y,z))

load("math", "mod")
m = mod(11, 3)
print("m =", m)

load("go", "sleep")
print(type(sleep))
`
	// prepare machine
	m := starlet.NewDefault()
	m.SetPrintFunc(getLogPrintFunc(t))
	m.SetPreloadModules(starlet.ModuleLoaderList{starlet.GetBuiltinModule("json")})
	m.SetLazyloadModules(starlet.ModuleLoaderMap{"math": starlet.GetBuiltinModule("math")})

	// run first time
	m.SetScript("test.star", []byte(code1), nil)
	out, err := m.Run()
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
	}
	if out == nil {
		t.Errorf("Unexpected empty result: %v", out)
	} else if len(out) != 3 {
		t.Errorf("Unexpected result: %v", out)
	} else {
		t.Logf("Result for the frist run: %v", out)
	}

	// run second time
	m.SetLazyloadModules(starlet.ModuleLoaderMap{"math": starlet.GetBuiltinModule("math"), "go": starlet.GetBuiltinModule("go_idiomatic")})
	m.SetScript("test.star", []byte(code2), os.DirFS("testdata"))
	out, err = m.Run()
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	if len(out) != 3 {
		t.Errorf("Unexpected result: %v", out)
	} else {
		t.Logf("Result for the second run: %v", out)
	}
}

func Test_Machine_Run_With_Timeout(t *testing.T) {
	interval := 1000 * time.Millisecond

	// prepare machine
	m := starlet.NewDefault()
	m.SetPrintFunc(getLogPrintFunc(t))
	m.SetGlobals(map[string]interface{}{
		"sleep_go": time.Sleep,
		"itn":      interval,
	})
	m.SetLazyloadModules(starlet.ModuleLoaderMap{"go": starlet.GetBuiltinModule("go_idiomatic")})

	// first run with no timeout
	m.SetScript("time.star", []byte(`a = 1; sleep_go(itn); b = 2`), nil)
	ts := time.Now()
	out, err := m.RunWithContext(nil, nil)
	expectSameDuration(t, time.Since(ts), interval)
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	t.Logf("got result after run #1: %v", out)

	// second run with timeout, but context is not handled in builtin sleep
	m.SetScript("time.star", []byte(`c = 3; sleep_go(itn); d = 4`), nil)
	ts = time.Now()
	ctx, _ := context.WithTimeout(context.Background(), interval/2)
	out, err = m.RunWithContext(ctx, nil)
	expectSameDuration(t, time.Since(ts), interval)
	expectErr(t, err, "starlark: exec: Starlark computation cancelled: context cancelled")
	t.Logf("got result after run #2: %v", out)

	// third run with timeout helper -- it's not timeout
	m.SetScript("time.star", []byte(`load("go", "sleep"); e = 5; sleep(0.5); f = 6`), nil)
	ts = time.Now()
	out, err = m.RunWithTimeout(interval, nil)
	expectSameDuration(t, time.Since(ts), interval/2)
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	t.Logf("got result after run #3: %v", out)

	// fourth run with timeout helper -- it's timeout indeed
	m.SetScript("time.star", []byte(`load("go", "sleep"); g = 7; sleep(1.5); h = 8`), nil)
	ts = time.Now()
	out, err = m.RunWithTimeout(interval, nil)
	expectSameDuration(t, time.Since(ts), interval)
	expectErr(t, err, "starlark: exec: context deadline exceeded")
	t.Logf("got result after run #4: %v", out)
}

func Test_Machine_Run_With_CancelledContext(t *testing.T) {
	// prepare machine
	m := starlet.NewDefault()
	m.SetPrintFunc(getLogPrintFunc(t))
	m.SetPreloadModules(starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic")})

	// prepare context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// run script
	m.SetScript("timer.star", []byte(`a = 1; b = 2`), nil)
	out, err := m.RunWithContext(ctx, nil)
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	if out == nil {
		t.Errorf("Unexpected empty result: %v", out)
	} else if len(out) != 2 {
		t.Errorf("Unexpected result: %v", out)
	} else {
		t.Logf("got result after run: %v", out)
	}
}

func Test_Machine_Run_With_Context(t *testing.T) {
	// prepare machine
	m := starlet.NewDefault()
	m.SetPrintFunc(getLogPrintFunc(t))
	m.SetPreloadModules(starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic")})

	// first run with no context
	m.SetScript("timer.star", []byte(`
x = 1
sleep(1)
y = 2
`), nil)
	ts := time.Now()
	out, err := m.RunWithContext(nil, nil)
	expectSameDuration(t, time.Since(ts), 1*time.Second)
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	t.Logf("got result after run #1: %v", out)

	// second run with timeout
	m.SetScript("timer.star", []byte(`
z = y << 5
sleep(1)
t = 4
`), nil)
	ts = time.Now()
	ctx, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
	out, err = m.RunWithContext(ctx, nil)
	expectSameDuration(t, time.Since(ts), 500*time.Millisecond)
	expectErr(t, err, "starlark: exec: context deadline exceeded")
	t.Logf("got result after run #2: %v", out)

	// third run without timeout
	m.SetScript("timer.star", []byte(`
z = y << 5
sleep(0.5)
t = 4
`), nil)
	ts = time.Now()
	ctx, _ = context.WithTimeout(context.Background(), 800*time.Millisecond) // TODO! occasionally, this test fails with 500ms timeout
	out, err = m.RunWithContext(ctx, nil)
	expectSameDuration(t, time.Since(ts), 500*time.Millisecond)
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	t.Logf("got result after run #3: %v", out)

	// fourth run with old values
	m.SetScript("timer.star", []byte(`
z = y << 5
fail("oops")
`), nil)
	ts = time.Now()
	out, err = m.Run()
	expectErr(t, err, "starlark: exec: fail: oops")
	t.Logf("got result after run #4: %v", out)

	// fifth run to fail in def
	m.SetScript("timer.star", []byte(`
def foo():
	fail("a bar")

zz = z * 2
foo()
`), nil)
	out, err = m.Run()
	expectErr(t, err, `starlark: exec: fail: a bar`)
	t.Logf("got result after run #5: %v", out)

	// sixth run with old values
	m.SetScript("timer.star", []byte(`
zoo = z + 100
`), nil)
	ts = time.Now()
	out, err = m.Run()
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	if out == nil {
		t.Errorf("Unexpected empty result: %v", out)
	} else if n := out["zoo"]; n != int64(164) {
		t.Errorf("Unexpected result: %v", out)
	} else {
		t.Logf("got result after run #6: %v", out)
	}
}

func Test_Machine_Run_With_Reset(t *testing.T) {
	// prepare machine
	m := starlet.NewDefault()
	m.SetPrintFunc(getLogPrintFunc(t))
	m.SetPreloadModules(starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic")})

	// first run with no context
	m.SetScript("run.star", []byte(`x = 100`), nil)
	out, err := m.Run()
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	if out == nil {
		t.Errorf("Unexpected empty result: %v", out)
	} else if n := out["x"]; n != int64(100) {
		t.Errorf("Unexpected result: %v", out)
	} else {
		t.Logf("got result after run #1: %v", out)
	}

	// without reset, the value can be reused
	m.SetScript("run.star", []byte(`y = x * 10`), nil)
	out, err = m.Run()
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	if out == nil {
		t.Errorf("Unexpected empty result: %v", out)
	} else if n := out["y"]; n != int64(1000) {
		t.Errorf("Unexpected result: %v", out)
	} else {
		t.Logf("got result after run #2: %v", out)
	}

	// without reset, all the old values are still there
	m.SetScript("run.star", []byte(`z = x + y`), nil)
	out, err = m.Run()
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	if out == nil {
		t.Errorf("Unexpected empty result: %v", out)
	} else if n := out["z"]; n != int64(1100) {
		t.Errorf("Unexpected result: %v", out)
	} else {
		t.Logf("got result after run #3: %v", out)
	}

	// with reset, the value is cleared
	m.Reset()
	m.SetScript("run.star", []byte(`w = x + z`), nil)
	_, err = m.Run()
	expectErr(t, err, `starlark: exec: run.star:1:5: undefined: x`)
}

func Test_Machine_Run_Panic(t *testing.T) {
	// prepare machine
	m := starlet.NewDefault()
	m.SetPrintFunc(getLogPrintFunc(t))
	m.SetPreloadModules(starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic")})

	// first run with error
	m.SetScript("panic.star", []byte(`
def foo():
	fail("oops")
foo = 123
`), nil)
	out, err := m.Run()
	expectErr(t, err, `starlark: exec: panic.star:4:1: cannot reassign global foo declared at panic.star:2:5`)
	t.Logf("got result after run #1: %v", out)

	// second run smoothly
	m.SetScript("panic.star", []byte(`
a = 123
b = str(a)
`), nil)
	out, err = m.Run()
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	if out == nil {
		t.Errorf("Unexpected empty result: %v", out)
	} else if n := out["b"]; n != "123" {
		t.Errorf("Unexpected result: %v", out)
	}
	t.Logf("got result after run #2: %v", out)

	// third run with panic and recovered in starlight
	lazyMap := func() (starlark.StringDict, error) {
		return convert.MakeStringDict(map[string]interface{}{
			"foo": 123,
			"bar": func() {
				panic("oops")
			},
		})
	}
	m.SetLazyloadModules(starlet.ModuleLoaderMap{"lucy": lazyMap})
	m.SetScript("panic.star", []byte(`
load("lucy", "foo", "bar")
ans = foo * 2
bar()
`), nil)
	out, err = m.Run()
	expectErr(t, err, `starlark: exec: panic in func fn: oops`)
	if out == nil {
		t.Errorf("Unexpected empty result: %v", out)
	} else if n := out["ans"]; n != int64(246) {
		t.Errorf("Unexpected result: %v", out)
	}
	t.Logf("got result after run #3: %v", out)

	// fourth run with panic
	lazyMap = func() (starlark.StringDict, error) {
		return convert.MakeStringDict(map[string]interface{}{
			"foo": 500,
			"panic": starlark.NewBuiltin("panic", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
				var s string
				if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &s); err != nil {
					return starlark.None, err
				}
				panic(s)
			}),
		})
	}
	m.SetLazyloadModules(starlet.ModuleLoaderMap{"lucky": lazyMap})
	m.SetScript("panic.star", []byte(`
load("lucky", "foo", "panic")
ans = foo * 2
panic("ohohoh")
`), nil)
	out, err = m.Run()
	expectErr(t, err, `starlark: exec: panic: ohohoh`)
	t.Logf("got result after run #4: %v", out)

	// fifth run normally
	m.SetScript("panic.star", []byte(`
res = ans % 100
`), nil)
	out, err = m.Run()
	if err != nil {
		t.Errorf("Expected no errors, got error: %v", err)
		return
	}
	if out == nil {
		t.Errorf("Unexpected empty result: %v", out)
	} else if n := out["res"]; n != int64(46) {
		t.Errorf("Unexpected result: %v", out)
	}
	t.Logf("got result after run #5: %v", out)
}

func TestRunScript(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		extras  starlet.StringAnyMap
		wantRes starlet.StringAnyMap
		wantErr bool
	}{
		{
			name:    "no code",
			wantRes: starlet.StringAnyMap{},
		},
		{
			name:    "only extra",
			extras:  starlet.StringAnyMap{"a": 123},
			wantRes: starlet.StringAnyMap{},
		},
		{
			name:    "simple assignment",
			code:    `a = 123`,
			wantRes: starlet.StringAnyMap{"a": int64(123)},
		},
		{
			name:    "simple assignment with extra",
			code:    `a = 123`,
			extras:  starlet.StringAnyMap{"b": 456},
			wantRes: starlet.StringAnyMap{"a": int64(123)},
		},
		{
			name:    "simple assignment overrides extra",
			code:    `a = 123`,
			extras:  starlet.StringAnyMap{"a": 456},
			wantRes: starlet.StringAnyMap{"a": int64(123)},
		},
		{
			name:    "simple assignment and extra",
			code:    `b = a`,
			extras:  starlet.StringAnyMap{"a": 456},
			wantRes: starlet.StringAnyMap{"b": int64(456)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, res, err := starlet.RunScript([]byte(tt.code), tt.extras)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if m == nil {
				t.Errorf("RunScript() got nil machine")
				return
			}
			if !reflect.DeepEqual(res, tt.wantRes) {
				t.Errorf("RunScript() got = %v, want %v", res, tt.wantRes)
				return
			}
			if tt.wantErr && err == nil {
				t.Errorf("RunScript() expected error, got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("RunScript() expected no error, got %v", err)
			}
		})
	}
}

func TestRunFile(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		inputFS   fs.FS
		extras    starlet.StringAnyMap
		wantRes   starlet.StringAnyMap
		wantErr   bool
	}{
		{
			name:      "no file",
			inputName: "no-file.star",
			inputFS:   MemFS{},
			wantErr:   true,
		},
		{
			name:      "no file name",
			inputName: "",
			inputFS:   MemFS{"a.star": `a = 123`},
			wantErr:   true,
		},
		{
			name:      "no file system",
			inputName: "no-fs.star",
			wantErr:   true,
		},
		{
			name:      "incorrect file name",
			inputName: "a.star",
			inputFS:   MemFS{"b.star": `a = 123`},
			wantErr:   true,
		},
		{
			name:      "empty file",
			inputName: "a.star",
			inputFS:   MemFS{"a.star": ``},
			wantRes:   starlet.StringAnyMap{},
		},
		{
			name:      "simple assignment",
			inputName: "a.star",
			inputFS:   MemFS{"a.star": `a = 123`},
			wantRes:   starlet.StringAnyMap{"a": int64(123)},
		},
		{
			name:      "simple assignment with extra",
			inputName: "a.star",
			inputFS:   MemFS{"a.star": `a = b`},
			extras:    starlet.StringAnyMap{"b": 456},
			wantRes:   starlet.StringAnyMap{"a": int64(456)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, res, err := starlet.RunFile(tt.inputName, tt.inputFS, tt.extras)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err != nil {
				return
			}
			if m == nil {
				t.Errorf("RunFile() got nil machine")
				return
			}
			if !reflect.DeepEqual(res, tt.wantRes) {
				t.Errorf("RunFile() got = %v, want %v", res, tt.wantRes)
				return
			}
			if tt.wantErr && err == nil {
				t.Errorf("RunFile() expected error, got nil")
			}
		})
	}
}

func TestRunTrustedScript(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		globals starlet.StringAnyMap
		extras  starlet.StringAnyMap
		wantRes starlet.StringAnyMap
		wantErr bool
	}{
		{
			name:    "no code",
			wantRes: starlet.StringAnyMap{},
		},
		{
			name:    "only extra",
			extras:  starlet.StringAnyMap{"a": 123},
			wantRes: starlet.StringAnyMap{},
		},
		{
			name:    "only globals",
			globals: starlet.StringAnyMap{"a": 123},
			wantRes: starlet.StringAnyMap{},
		},
		{
			name:    "simple assignment",
			code:    `a = 123`,
			wantRes: starlet.StringAnyMap{"a": int64(123)},
		},
		{
			name:    "simple assignment with extra",
			code:    `a = 123`,
			extras:  starlet.StringAnyMap{"b": 456},
			wantRes: starlet.StringAnyMap{"a": int64(123)},
		},
		{
			name:    "simple assignment with globals",
			code:    `a = 123`,
			globals: starlet.StringAnyMap{"b": 456},
			wantRes: starlet.StringAnyMap{"a": int64(123)},
		},
		{
			name:    "assignment with extra and globals",
			code:    `a = x + y`,
			globals: starlet.StringAnyMap{"x": 100},
			extras:  starlet.StringAnyMap{"y": 200},
			wantRes: starlet.StringAnyMap{"a": int64(300)},
		},
		{
			name:    "access to builtins",
			code:    `a = math.sqrt(4)`,
			wantRes: starlet.StringAnyMap{"a": 2.0},
		},
		{
			name:    "access to builtins with globals",
			code:    `a = math.sqrt(x)`,
			globals: starlet.StringAnyMap{"x": 4},
			wantRes: starlet.StringAnyMap{"a": 2.0},
		},
		{
			name:    "access to builtins with extra",
			code:    `a = math.sqrt(x)`,
			extras:  starlet.StringAnyMap{"x": 4},
			wantRes: starlet.StringAnyMap{"a": 2.0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, res, err := starlet.RunTrustedScript([]byte(tt.code), tt.globals, tt.extras)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunTrustedScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if m == nil {
				t.Errorf("RunTrustedScript() got nil machine")
				return
			}
			if !reflect.DeepEqual(res, tt.wantRes) {
				t.Errorf("RunTrustedScript() got = %v, want %v", res, tt.wantRes)
				return
			}
			if tt.wantErr && err == nil {
				t.Errorf("RunTrustedScript() expected error, got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("RunTrustedScript() expected no error, got %v", err)
			}
		})
	}
}

func TestRunTrustedFile(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		inputFS   fs.FS
		globals   starlet.StringAnyMap
		extras    starlet.StringAnyMap
		wantRes   starlet.StringAnyMap
		wantErr   bool
	}{
		{
			name:    "no file",
			inputFS: MemFS{},
			wantRes: starlet.StringAnyMap{},
			wantErr: true,
		},
		{
			name:      "empty file",
			inputName: "a.star",
			inputFS:   MemFS{"a.star": ``},
			wantRes:   starlet.StringAnyMap{},
		},
		{
			name:      "simple assignment",
			inputName: "a.star",
			inputFS:   MemFS{"a.star": `a = 123`},
			wantRes:   starlet.StringAnyMap{"a": int64(123)},
		},
		{
			name:      "simple assignment with extra",
			inputName: "a.star",
			inputFS:   MemFS{"a.star": `a = b`},
			extras:    starlet.StringAnyMap{"b": 456},
			wantRes:   starlet.StringAnyMap{"a": int64(456)},
		},
		{
			name:      "use of builtins",
			inputName: "a.star",
			inputFS:   MemFS{"a.star": `a = math.sqrt(x) + math.sqrt(y)`},
			globals:   starlet.StringAnyMap{"x": 4},
			extras:    starlet.StringAnyMap{"y": 9},
			wantRes:   starlet.StringAnyMap{"a": 5.0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, res, err := starlet.RunTrustedFile(tt.inputName, tt.inputFS, tt.globals, tt.extras)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunTrustedFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err != nil {
				return
			}
			if m == nil {
				t.Errorf("RunTrustedFile() got nil machine")
				return
			}
			if !reflect.DeepEqual(res, tt.wantRes) {
				t.Errorf("RunTrustedFile() got = %v, want %v", res, tt.wantRes)
				return
			}
			if tt.wantErr && err == nil {
				t.Errorf("RunTrustedFile() expected error, got nil")
			}
		})
	}
}

func TestMachine_REPL_OK(t *testing.T) {
	m := starlet.NewDefault()
	m.REPL()
}

func TestMachine_REPL_Error(t *testing.T) {
	m := starlet.NewDefault()
	m.SetGlobals(starlet.StringAnyMap{"x": make(chan int, 1)})
	m.REPL()
}
