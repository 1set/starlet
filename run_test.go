package starlet_test

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"starlet"
	"testing"

	"go.starlark.net/starlark"
)

func Test_DefaultMachine_Run_NoCode(t *testing.T) {
	m := starlet.NewDefault()
	// run with empty script
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: run: no script to execute`)
}

func Test_DefaultMachine_Run_NoSpecificFile(t *testing.T) {
	m := starlet.NewDefault()
	m.SetScript("", nil, os.DirFS("testdata"))
	// run with no specific file name
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: run: no specific file`)
}

func Test_DefaultMachine_Run_APlusB(t *testing.T) {
	m := starlet.NewDefault()
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

func Test_DefaultMachine_Run_HelloWorld(t *testing.T) {
	m := starlet.NewDefault()
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

func Test_DefaultMachine_Run_LocalFile(t *testing.T) {
	m := starlet.NewDefault()
	// set print function
	printFunc, cmpFunc := getPrintCompareFunc(t)
	m.SetPrintFunc(printFunc)
	// set code
	m.SetScript("aloha.star", nil, os.DirFS("testdata"))
	// run
	_, err := m.Run(context.Background())
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
	_, err := m.Run(context.Background())
	if isOnWindows {
		expectErr(t, err, `starlet: open: open`, `The system cannot find the file specified.`)
	} else {
		expectErr(t, err, `starlet: open: open`, `: no such file or directory`)
	}
}

func Test_DefaultMachine_Run_FSNonExist(t *testing.T) {
	m := starlet.NewDefault()
	// set code
	m.SetScript("aloha.star", nil, os.DirFS("not-found-dir"))
	// run
	_, err := m.Run(context.Background())
	if isOnWindows {
		expectErr(t, err, `starlet: open: open`, `The system cannot find the path specified.`)
	} else {
		expectErr(t, err, `starlet: open: open`, `: no such file or directory`)
	}
}

func Test_DefaultMachine_Run_LoadFunc(t *testing.T) {
	m := starlet.NewDefault()
	// set code
	code := `load("fibonacci.star", "fibonacci"); val = fibonacci(10)[-1]`
	m.SetScript("test.star", []byte(code), os.DirFS("testdata"))
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

func Test_DefaultMachine_Run_LoadNonExist(t *testing.T) {
	m := starlet.NewDefault()
	// set code
	code := `load("nonexist.star", "a")`
	m.SetScript("test.star", []byte(code), os.DirFS("testdata"))
	// run
	_, err := m.Run(context.Background())
	// check result
	if isOnWindows {
		expectErr(t, err, `starlet: exec: cannot load nonexist.star: open`, `The system cannot find the file specified.`)
	} else {
		expectErr(t, err, `starlet: exec: cannot load nonexist.star: open`, `: no such file or directory`)
	}
}

func Test_Machine_Run_Globals(t *testing.T) {
	m := starlet.NewWithNames(map[string]interface{}{
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

func Test_Machine_Run_File_Globals(t *testing.T) {
	m := starlet.NewWithNames(map[string]interface{}{
		"magic_number": 30,
	}, nil, nil)
	// set code
	m.SetScript("magic.star", nil, os.DirFS("testdata"))
	// run
	out, err := m.Run(context.Background())
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
	out, err := m.Run(context.Background())
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
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: exec: magic.star:5:32: undefined: magic_number`)
}

func Test_Machine_Run_PreloadModules(t *testing.T) {
	m := starlet.NewWithNames(nil, []string{"go_idiomatic"}, nil)
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

func Test_Machine_Run_LazyloadModules(t *testing.T) {
	m := starlet.NewWithNames(nil, nil, []string{"go_idiomatic"})
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

func Test_Machine_Run_Load_Shadow_Globals(t *testing.T) {
	// enable global reassign only for this test, if it's not enabled, it will fail for: local variable fibonacci referenced before assignment
	starlet.EnableGlobalReassign()
	defer func() {
		starlet.DisableGlobalReassign()
	}()
	// create machine
	m := starlet.NewWithNames(map[string]interface{}{"fibonacci": 123}, nil, nil)
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
	out, err := m.Run(context.Background())
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

func Test_Machine_Run_Load_With_Globals(t *testing.T) {
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
	out, err := m.Run(context.Background())
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

func Test_Machine_Run_LoadErrors(t *testing.T) {
	testFS := os.DirFS("testdata")
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
			name:        "Unsupport Globals Type",
			globals:     map[string]interface{}{"a": make(chan int)},
			code:        `b = a`,
			expectedErr: `starlet: convert: type chan int is not a supported starlark type`,
		},
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
			preloadMods: []string{},
			code:        `a = nil == None`,
			expectedErr: `starlet: exec: test.star:1:5: undefined: nil`,
		},
		{
			name:          "NonExist Preload Modules",
			preloadMods:   []string{"nonexist"},
			code:          `a = nil == None`,
			expectedErr:   `starlet: module "nonexist": module not found`,
			expectedPanic: true,
		},
		// for lazyload modules
		{
			name:        "Missed load() for LazyLoad Modules",
			lazyMods:    []string{"go_idiomatic"},
			code:        `a = nil == None`,
			expectedErr: `starlet: exec: test.star:1:5: undefined: nil`,
		},
		{
			name:          "NonExist LazyLoad Modules",
			lazyMods:      []string{"nonexist"},
			code:          `load("nonexist", "nil"); a = nil == None`,
			expectedErr:   `starlet: module "nonexist": module not found`,
			expectedPanic: true,
		},
		{
			name:        "NonExist Function in LazyLoad Modules",
			lazyMods:    []string{"go_idiomatic"},
			code:        `load("go_idiomatic", "fake"); a = fake == None`,
			expectedErr: `starlet: exec: load: name fake not found in module go_idiomatic`,
		},
		// for load fs --- user modules
		{
			name:        "No FS for User Modules",
			code:        `load("fibonacci.star", "fibonacci"); val = fibonacci(10)[-1]`,
			expectedErr: `starlet: exec: cannot load fibonacci.star: no file system given`,
		},
		{
			name:        "Missed load() for User Modules",
			code:        `val = fibonacci(10)[-1]`,
			modFS:       testFS,
			expectedErr: `starlet: exec: test.star:1:7: undefined: fibonacci`,
		},
		{
			name:        "Duplicate load() for User Modules",
			code:        `load("fibonacci.star", "fibonacci"); load("fibonacci.star", "fibonacci"); val = fibonacci(10)[-1]`,
			modFS:       testFS,
			expectedErr: `starlet: exec: test.star:1:62: cannot reassign top-level fibonacci`,
		},
		{
			name:        "NonExist User Modules",
			code:        `load("nonexist.star", "fibonacci"); val = fibonacci(10)[-1]`,
			modFS:       testFS,
			expectedErr: `starlet: exec: cannot load nonexist.star: open`,
		},
		{
			name:        "NonExist Function in User Modules",
			code:        `load("fibonacci.star", "fake"); val = fake(10)[-1]`,
			modFS:       testFS,
			expectedErr: `starlet: exec: load: name fake not found in module fibonacci.star`,
		},
		{
			name:        "Existing and NonExist Function in User Modules",
			code:        `load("fibonacci.star", "fibonacci", "fake"); val = fibonacci(10)[-1]`,
			modFS:       testFS,
			expectedErr: `starlet: exec: load: name fake not found in module fibonacci.star`,
		},
		// for globals + user modules
		{
			name:        "User Modules Override Globals",
			globals:     map[string]interface{}{"fibonacci": 2},
			code:        `x = fibonacci * 10; load("fibonacci.star", "fibonacci"); val = fibonacci(10)[-1]; print(x, val)`,
			modFS:       testFS,
			expectedErr: `starlet: exec: local variable fibonacci referenced before assignment`,
			// NOTE: for this behavior: read the comments before `r.useToplevel(use)` in `func (r *resolver) use(id *syntax.Ident)` in file: go.starlark.net@v0.0.0-20230525235612-a134d8f9ddca/resolve/resolve.go
		},
	}

	//starlet.EnableGlobalReassign()
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
			_, err := m.Run(context.Background())
			expectErr(t, err, tc.expectedErr)
		})
	}
}

func Test_Machine_Run_Loaders(t *testing.T) {
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
			expectedErr: `starlet: nil module loader`,
		},
		{
			name:        "Nil Loader Map Element",
			lazyMap:     starlet.ModuleLoaderMap{"nil_loader": nil},
			code:        `load("nil_loader", "num"); val = 5 + 6`,
			modFS:       testFS,
			expectedErr: `starlet: exec: cannot load nil_loader: nil module loader "nil_loader"`,
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
			expectedErr: `starlet: failed to load module: invalid module loader`,
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
			expectedErr: fmt.Sprintf(`starlet: exec: cannot load %s: invalid module loader`, failName),
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
			expectedErr: `starlet: exec: test.star:3:25: cannot reassign top-level number`,
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
			out, err := m.Run(context.Background())

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
