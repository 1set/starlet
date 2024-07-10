package starlet_test

import (
	"io"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/1set/starlet"
	"github.com/1set/starlet/dataconv"
	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	builtinModules = []string{"atom", "base64", "csv", "file", "go_idiomatic", "hashlib", "http", "json", "log", "math", "path", "random", "re", "runtime", "stats", "string", "struct", "time"}
)

func TestListBuiltinModules(t *testing.T) {
	modules := starlet.GetAllBuiltinModuleNames()

	expectedModules := builtinModules
	if len(modules) != len(expectedModules) {
		t.Errorf("Expected %d modules, got %d", len(expectedModules), len(modules))
	}

	for i, module := range modules {
		if module != expectedModules[i] {
			t.Errorf("Expected module %s, got %s", expectedModules[i], module)
		}
	}
}

func TestGetBuiltinModule(t *testing.T) {
	tests := []struct {
		name  string
		found bool
	}{
		{"go_idiomatic", true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := starlet.GetBuiltinModule(tt.name)
			if tt.found && got == nil {
				t.Errorf("Expected module %q, got nil", tt.name)
			} else if !tt.found && got != nil {
				t.Errorf("Expected nil, got module %q", tt.name)
			}
		})
	}
}

func Test_ModuleLoaderList_Clone(t *testing.T) {
	moduleLoaderList := starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic"), starlet.GetBuiltinModule("struct")}

	clone := moduleLoaderList.Clone()
	if len(clone) != len(moduleLoaderList) {
		t.Errorf("Expected clone length %d, got %d", len(moduleLoaderList), len(clone))
	}
	for i := range clone {
		if reflect.DeepEqual(clone[i], moduleLoaderList[i]) {
			t.Errorf("Expected different clone at index %d to be %T, got %T", i, moduleLoaderList[i], clone[i])
		}
	}
}

func Test_ModuleLoaderList_LoadAll(t *testing.T) {
	_, failLoader := getErrorModuleLoader()
	tests := []struct {
		name          string
		moduleLoaders starlet.ModuleLoaderList
		dict          starlark.StringDict
		wantErr       string
	}{
		{
			name:          "valid modules",
			moduleLoaders: starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic"), starlet.GetBuiltinModule("struct")},
			dict:          make(starlark.StringDict),
		},
		{
			name:          "nil dict",
			moduleLoaders: starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic"), starlet.GetBuiltinModule("struct")},
			dict:          nil,
			wantErr:       "starlet: load: cannot load modules into nil dict",
		},
		{
			name:          "nil module loader",
			moduleLoaders: starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic"), nil},
			dict:          make(starlark.StringDict),
			wantErr:       "starlet: load: nil module loader",
		},
		{
			name:          "invalid module",
			moduleLoaders: starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic"), failLoader},
			dict:          make(starlark.StringDict),
			wantErr:       "starlet: load: invalid module loader",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.moduleLoaders.LoadAll(tt.dict)
			if tt.wantErr != "" {
				expectErr(t, err, tt.wantErr)
			} else if err != nil {
				t.Errorf("Expected no error, got '%v'", err)
			}
		})
	}
}

func Test_ModuleLoaderMap_Clone(t *testing.T) {
	moduleLoaderMap := starlet.ModuleLoaderMap{
		"go_idiomatic": starlet.GetBuiltinModule("go_idiomatic"),
		"struct":       starlet.GetBuiltinModule("struct"),
	}

	clone := moduleLoaderMap.Clone()
	if len(clone) != len(moduleLoaderMap) {
		t.Errorf("Expected clone length %d, got %d", len(moduleLoaderMap), len(clone))
	}
	for k := range clone {
		if reflect.DeepEqual(clone[k], moduleLoaderMap[k]) {
			t.Errorf("Expected different clone at key %q to be %T, got %T", k, moduleLoaderMap[k], clone[k])
		}
	}
}

func Test_ModuleLoaderMap_Merge(t *testing.T) {
	original := starlet.ModuleLoaderMap{
		"go_idiomatic": starlet.GetBuiltinModule("go_idiomatic"),
		"struct":       starlet.GetBuiltinModule("struct"),
	}
	other := starlet.ModuleLoaderMap{
		"go_idiomatic": starlet.GetBuiltinModule("go_idiomatic"),
		"time":         starlet.GetBuiltinModule("time"),
	}
	expected := starlet.ModuleLoaderMap{
		"go_idiomatic": starlet.GetBuiltinModule("go_idiomatic"),
		"struct":       starlet.GetBuiltinModule("struct"),
		"time":         starlet.GetBuiltinModule("time"),
	}
	var nilMap starlet.ModuleLoaderMap

	original.Merge(other)
	if len(original) != 3 {
		t.Errorf("Expected merged map length %d, got %d", 3, len(original))
	}
	for k := range original {
		if _, ok := expected[k]; !ok {
			t.Errorf("Unexpected key %q in merged map", k)
		}
	}

	nilMap.Merge(other)
	if len(nilMap) != 0 {
		t.Errorf("Expected merged nil map length %d, got %d", 0, len(nilMap))
	}

	other.Merge(nilMap)
	if len(other) != 2 {
		t.Errorf("Expected merged other map length %d, got %d", 2, len(other))
	}
}

func Test_ModuleLoaderMap_GetLazyLoader(t *testing.T) {
	failName, failLoader := getErrorModuleLoader()
	tests := []struct {
		name          string
		moduleLoaders starlet.ModuleLoaderMap
		moduleName    string
		wantErr       string
		wantMod       bool
	}{
		{
			name:          "nil map",
			moduleLoaders: nil,
			moduleName:    "unknown",
		},
		{
			name:          "nil module",
			moduleLoaders: starlet.ModuleLoaderMap{"unknown": nil},
			moduleName:    "unknown",
			wantErr:       `nil module loader`,
		},
		{
			name:          "valid module",
			moduleLoaders: starlet.ModuleLoaderMap{"go_idiomatic": starlet.GetBuiltinModule("go_idiomatic")},
			moduleName:    "go_idiomatic",
			wantMod:       true,
		},
		{
			name:          "invalid module",
			moduleLoaders: starlet.ModuleLoaderMap{failName: failLoader},
			moduleName:    failName,
			wantErr:       "invalid module loader",
		},
		{
			name:          "embedded module",
			moduleLoaders: starlet.ModuleLoaderMap{"hashlib": starlet.GetBuiltinModule("hashlib")},
			moduleName:    "hashlib",
			wantMod:       true,
		},
		{
			name:          "embedded struct",
			moduleLoaders: starlet.ModuleLoaderMap{"http": starlet.GetBuiltinModule("http")},
			moduleName:    "http",
			wantMod:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := tt.moduleLoaders.GetLazyLoader()
			mod, err := loader(tt.moduleName)
			if tt.wantErr != "" {
				expectErr(t, err, tt.wantErr)
			} else if err != nil {
				t.Errorf("Expected no error, got '%v'", err)
			}
			if tt.wantMod && mod == nil {
				t.Errorf("Expected loader, got nil")
			} else if !tt.wantMod && mod != nil {
				t.Errorf("Expected nil, got loader: %v", mod)
			}
		})
	}
}

func Test_MakeBuiltinModuleLoaderList(t *testing.T) {
	tests := []struct {
		name         string
		modules      []string
		wantErr      string
		expectedSize int
	}{
		{
			name:         "valid modules",
			modules:      []string{"json", "math", "time", "struct", "go_idiomatic"},
			expectedSize: 5,
		},
		{
			name:         "non-existent module",
			modules:      []string{"non_existent"},
			wantErr:      "starlet: make: module not found: non_existent",
			expectedSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loaderList, err := starlet.MakeBuiltinModuleLoaderList(tt.modules...)
			if tt.wantErr != "" {
				expectErr(t, err, tt.wantErr)
			} else if err != nil {
				t.Errorf("Expected no error, got '%v'", err)
			}
			if len(loaderList) != tt.expectedSize {
				t.Fatalf("Expected %d loaders, got %d", tt.expectedSize, len(loaderList))
			}
		})
	}
}

func Test_MakeBuiltinModuleLoaderMap(t *testing.T) {
	tests := []struct {
		name         string
		modules      []string
		wantErr      string
		expectedSize int
	}{
		{
			name:         "valid modules",
			modules:      []string{"json", "math", "time", "struct", "go_idiomatic"},
			expectedSize: 5,
		},
		{
			name:         "non-existent module",
			modules:      []string{"non_existent"},
			wantErr:      "starlet: make: module not found: non_existent",
			expectedSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loaderMap, err := starlet.MakeBuiltinModuleLoaderMap(tt.modules...)
			if tt.wantErr != "" {
				expectErr(t, err, tt.wantErr)
			} else if err != nil {
				t.Errorf("Expected no error, got '%v'", err)
			}
			if len(loaderMap) != tt.expectedSize {
				t.Fatalf("Expected %d loaders, got %d", tt.expectedSize, len(loaderMap))
			}
		})
	}
}

func Test_ModuleLoaderFromString(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		source      string
		predeclared starlark.StringDict
		wantKeys    []string
		wantErr     string
	}{
		{
			name:        "empty filename",
			fileName:    "",
			source:      "a = 1",
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantKeys:    []string{"a"},
		},
		{
			name:        "empty source",
			fileName:    "test.star",
			source:      "",
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantKeys:    []string{},
		},
		{
			name:        "empty predeclared",
			fileName:    "test.star",
			source:      "a = 1",
			predeclared: nil,
			wantKeys:    []string{"a"},
		},
		{
			name:        "override predeclared",
			fileName:    "test.star",
			source:      "a = 10",
			predeclared: map[string]starlark.Value{"a": starlark.MakeInt(2)},
			wantKeys:    []string{"a"}, // source overrides predeclared
		},
		{
			name:     "functions",
			fileName: "test.star",
			source: `
def foo():
  return 1

def bar():
  return 2

val = 3
`,
			wantKeys: []string{"foo", "bar", "val"},
		},
		{
			name:     "multiple keys",
			fileName: "test.star",
			source:   "a = 1\nb = 2",
			wantKeys: []string{"a", "b"},
		},
		{
			name:     "invalid source code",
			fileName: "wrong.star",
			source:   "asfdasf",
			wantErr:  "wrong.star:1:1: undefined: asfdasf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// make and run the module
			loader := starlet.MakeModuleLoaderFromString(tt.fileName, tt.source, tt.predeclared)
			mod, err := loader()
			if tt.wantErr != "" {
				expectErr(t, err, tt.wantErr)
			}
			// check the module result
			la := len(mod)
			le := len(tt.wantKeys)
			if la != le {
				t.Errorf("Expected module has %d keys, got %d", le, la)
				return
			}
			if len(tt.wantKeys) != 0 {
				for _, key := range tt.wantKeys {
					if _, ok := mod[key]; !ok {
						t.Errorf("Expected module to contain key %q, got: %v", key, mod)
						return
					}
				}
			}
		})
	}
}

func Test_ModuleLoaderFromReader(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		source      io.Reader
		predeclared starlark.StringDict
		wantKeys    []string
		wantErr     string
	}{
		{
			name:        "empty filename",
			fileName:    "",
			source:      strings.NewReader("a = 1"),
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantKeys:    []string{"a"},
		},
		{
			name:        "empty source",
			fileName:    "test.star",
			source:      strings.NewReader(""),
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantKeys:    []string{},
		},
		{
			name:        "nil source",
			fileName:    "none.star",
			source:      nil,
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantErr:     "open none.star:",
		},
		{
			name:     "first error reader",
			fileName: "wrong.star",
			source:   newErrorReader("a = 1\nb = 2\nc = 3", 1),
			wantErr:  `read wrong.star: desired error at 1`,
		},
		{
			name:     "second error reader",
			fileName: "wrong.star",
			source:   newErrorReader("a = 1\nb = 2\nc = 3\n", 2),
			wantErr:  `read wrong.star: desired error at 2`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// make and run the module
			loader := starlet.MakeModuleLoaderFromReader(tt.fileName, tt.source, tt.predeclared)
			mod, err := loader()
			if tt.wantErr != "" {
				expectErr(t, err, tt.wantErr)
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			// check the module result
			la := len(mod)
			le := len(tt.wantKeys)
			if la != le {
				t.Errorf("Expected module has %d keys, got %d", le, la)
				return
			}
			if len(tt.wantKeys) != 0 {
				for _, key := range tt.wantKeys {
					if _, ok := mod[key]; !ok {
						t.Errorf("Expected module to contain key %q, got: %v", key, mod)
						return
					}
				}
			}
		})
	}
}

func Test_ModuleLoaderFromFile(t *testing.T) {
	testFS := os.DirFS("testdata")
	nonExistFS := os.DirFS("nonexistent")
	tests := []struct {
		name        string
		fileName    string
		fileSys     fs.FS
		predeclared starlark.StringDict
		wantKeys    []string
		wantErr     string
	}{
		{
			name:        "empty filename",
			fileName:    "",
			fileSys:     testFS,
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantErr:     "no file name given",
		},
		{
			name:        "empty file system",
			fileName:    "test.star",
			fileSys:     nil,
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantErr:     "no file system given",
		},
		{
			name:        "nonexistent file",
			fileName:    "nonexistent.star",
			fileSys:     testFS,
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantKeys:    []string{},
			wantErr:     "open ",
		},
		{
			name:        "nonexistent file system",
			fileName:    "test.star",
			fileSys:     nonExistFS,
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantKeys:    []string{},
			wantErr:     "open ",
		},
		{
			name:        "empty file",
			fileName:    "empty.star",
			fileSys:     testFS,
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantKeys:    []string{},
		},
		{
			name:        "function file",
			fileName:    "fibonacci.star",
			fileSys:     testFS,
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantKeys:    []string{"fibonacci", "fib_last"},
		},
		{
			name:        "omit file extension",
			fileName:    "fibonacci",
			fileSys:     testFS,
			predeclared: map[string]starlark.Value{"b": starlark.MakeInt(2)},
			wantKeys:    []string{"fibonacci", "fib_last"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// make and run the module
			loader := starlet.MakeModuleLoaderFromFile(tt.fileName, tt.fileSys, tt.predeclared)
			mod, err := loader()
			if tt.wantErr != "" {
				expectErr(t, err, tt.wantErr)
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			// check the module result
			la := len(mod)
			le := len(tt.wantKeys)
			if la != le {
				t.Errorf("Expected module has %d keys, got %d", le, la)
				return
			}
			if len(tt.wantKeys) != 0 {
				for _, key := range tt.wantKeys {
					if _, ok := mod[key]; !ok {
						t.Errorf("Expected module to contain key %q, got: %v", key, mod)
						return
					}
				}
			}
		})
	}
}

func TestMakeModuleLoaderFromStringDict(t *testing.T) {
	tests := []struct {
		name     string
		dict     starlark.StringDict
		wantKeys []string
	}{
		{
			name:     "nil dict",
			dict:     nil,
			wantKeys: []string{},
		},
		{
			name:     "empty dict",
			dict:     map[string]starlark.Value{},
			wantKeys: []string{},
		},
		{
			name:     "non-empty dict",
			dict:     map[string]starlark.Value{"a": starlark.MakeInt(1), "b": starlark.MakeInt(2)},
			wantKeys: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := starlet.MakeModuleLoaderFromStringDict(tt.dict)
			mod, err := loader()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			// check the module result
			la := len(mod)
			le := len(tt.wantKeys)
			if la != le {
				t.Errorf("Expected module has %d keys, got %d", le, la)
				return
			}
			if len(tt.wantKeys) != 0 {
				for _, key := range tt.wantKeys {
					if _, ok := mod[key]; !ok {
						t.Errorf("Expected module to contain key %q, got: %v", key, mod)
						return
					}
				}
			}
		})
	}
}

func TestMakeModuleLoaderFromMap(t *testing.T) {
	tests := []struct {
		name     string
		dict     map[string]interface{}
		wantKeys []string
		wantErr  bool
	}{
		{
			name:     "nil dict",
			dict:     nil,
			wantKeys: []string{},
		},
		{
			name:     "empty dict",
			dict:     map[string]interface{}{},
			wantKeys: []string{},
		},
		{
			name:     "non-empty dict",
			dict:     map[string]interface{}{"a": 1, "b": 2},
			wantKeys: []string{"a", "b"},
		},
		{
			name:    "invalid dict",
			dict:    map[string]interface{}{"a": 1, "b": 2, "c": make(chan int)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := starlet.MakeModuleLoaderFromMap(tt.dict)
			mod, err := loader()
			// check errors
			if err != nil && !tt.wantErr {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if err == nil && tt.wantErr {
				t.Errorf("Expected error, got nil")
				return
			}
			if tt.wantErr {
				return
			}
			// check the module result
			la := len(mod)
			le := len(tt.wantKeys)
			if la != le {
				t.Errorf("Expected module has %d keys, got %d", le, la)
				return
			}
			if len(tt.wantKeys) != 0 {
				for _, key := range tt.wantKeys {
					if _, ok := mod[key]; !ok {
						t.Errorf("Expected module to contain key %q, got: %v", key, mod)
						return
					}
				}
			}
		})
	}
}

func TestAllBuiltinModules(t *testing.T) {
	ml := []string{"go_idiomatic", "json", "math", "struct", "time"}
	m := starlet.NewWithNames(map[string]interface{}{}, ml, nil)
	m.SetPrintFunc(getLogPrintFunc(t))
	// set code
	code := `
def assert(cond, msg=None):
	if not cond:
		fail(msg)

assert(true != false, "true is not false")
assert(nil == None, "nil is None")

s = struct(name="test", age=10, tags=["a", "b", "c"])
print(s, type(s))

sj = json.encode(s)
print(sj, type(sj))

sd = json.decode(sj)
print(sd, type(sd))

f = math.sqrt(2)
print(f, type(f))

t1 = time.now()
print("now", t1, type(t1))
t2 = time.time(year=2023, month=5, day=20)
print("birth", t2, type(t2))
d = t1 - t2
print("dh", d.hours, type(d))
`
	// run
	out, err := m.RunScript([]byte(code), nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out == nil {
		t.Errorf("unexpected nil output")
	} else {
		t.Logf("output: %v", out)
		t.Logf("machine: %v", m)
	}
}

// It tests the modifications on module loaders implemented via StringDict, starstruct.Module, and starstruct.Struct and loads via direct load and lazy load.
func Test_ModuleLoader_Modify(t *testing.T) {
	// load 1
	load1 := func() (starlark.StringDict, starlet.ModuleLoader) {
		sd1 := starlark.StringDict{
			"a": starlark.MakeInt(1),
			"b": starlark.MakeInt(2),
			"l": starlark.NewList([]starlark.Value{starlark.MakeInt(3)}),
			"s": starlark.String("test"),
		}
		sdLoad := func() (starlark.StringDict, error) {
			return sd1, nil
		}
		return sd1, sdLoad
	}
	// load 2
	load2 := func() (starlark.StringDict, starlet.ModuleLoader) {
		sd2 := starlark.StringDict{
			"a": starlark.MakeInt(10),
			"b": starlark.MakeInt(20),
			"l": starlark.NewList([]starlark.Value{starlark.MakeInt(30)}),
			"s": starlark.String("test"),
		}
		smLoad := func() (starlark.StringDict, error) {
			return starlark.StringDict{
				"foo": &starlarkstruct.Module{
					Name:    "foo",
					Members: sd2,
				},
			}, nil
		}
		return sd2, smLoad
	}
	// load 3
	load3 := func() (starlark.StringDict, starlet.ModuleLoader) {
		sd3 := starlark.StringDict{
			"a": starlark.MakeInt(100),
			"b": starlark.MakeInt(200),
			"l": starlark.NewList([]starlark.Value{starlark.MakeInt(300)}),
			"s": starlark.String("test"),
		}
		ssLoad := func() (starlark.StringDict, error) {
			ss := starlarkstruct.FromStringDict(starlark.String("bar"), sd3)
			return starlark.StringDict{
				"bar": ss,
			}, nil
		}
		return sd3, ssLoad
	}

	// helpers
	getEqualFunc := func(name string, val interface{}) func(anyMap starlet.StringAnyMap) bool {
		return func(anyMap starlet.StringAnyMap) bool {
			v, found := anyMap[name]
			if !found {
				return false
			}
			return v == val
		}
	}

	// test cases
	tests := []struct {
		name       string
		preload    func() (starlark.StringDict, starlet.ModuleLoader)
		lazyload   func() (starlark.StringDict, starlet.ModuleLoader)
		moduleName string
		script     string
		wantErr    bool
		checkRes   func(anyMap starlet.StringAnyMap) bool
		checkData  func(sd starlark.StringDict) bool
	}{
		// Check Preload Read
		{
			name:    "Preload StringDict",
			preload: load1,
			script: `
c = a + b + l[0]
`,
			wantErr:  false,
			checkRes: getEqualFunc("c", int64(6)),
		},
		{
			name:    "Preload Module",
			preload: load2,
			script: `
c = foo.a + foo.b + foo.l[0]
`,
			wantErr:  false,
			checkRes: getEqualFunc("c", int64(60)),
		},
		{
			name:    "Preload Struct",
			preload: load3,
			script: `
c = bar.a + bar.b + bar.l[0]
`,
			wantErr:  false,
			checkRes: getEqualFunc("c", int64(600)),
		},

		// Check LazyLoad Read
		{
			name:       "LazyLoad StringDict",
			lazyload:   load1,
			moduleName: "play",
			script: `
load("play", "a", "b", "l")
c = a + b + l[0]
`,
			wantErr:  false,
			checkRes: getEqualFunc("c", int64(6)),
		},
		{
			name:       "LazyLoad Module",
			lazyload:   load2,
			moduleName: "play",
			script: `
load("play", "foo")
c = foo.a + foo.b + foo.l[0]
`,
			wantErr:  false,
			checkRes: getEqualFunc("c", int64(60)),
		},
		{
			name:       "LazyLoad Named Module",
			lazyload:   load2,
			moduleName: "foo",
			script: `
load("foo", "a", "b", "l")
c = a + b + l[0]
`,
			wantErr:  false,
			checkRes: getEqualFunc("c", int64(60)),
		},
		{
			name:       "LazyLoad Struct",
			lazyload:   load3,
			moduleName: "play",
			script: `
load("play", "bar")
c = bar.a + bar.b + bar.l[0]
`,
			wantErr:  false,
			checkRes: getEqualFunc("c", int64(600)),
		},
		{
			name:       "LazyLoad Named Struct",
			lazyload:   load3,
			moduleName: "bar",
			script: `
load("bar", "a", "b", "l")
c = a + b + l[0]
`,
			wantErr:  false,
			checkRes: getEqualFunc("c", int64(600)),
		},

		// Edit Preload
		{
			name:    "Edit Preload StringDict",
			preload: load1,
			script: `
a = 100
b = 200
l.append(300)
s = "new"
`,
			wantErr: false,
			checkData: func(sd starlark.StringDict) bool {
				return sd["a"] == starlark.MakeInt(1) && sd["b"] == starlark.MakeInt(2) && sd["l"].(*starlark.List).Len() == 2 && sd["s"] == starlark.String("test")
			},
		},
		{
			name:    "Assign Preload Module",
			preload: load2,
			script: `
print(type(foo), dir(foo), foo)
foo.a = 400
foo.b = 500
foo.s = "new"
`,
			wantErr: true,
		},
		{
			name:    "Modify Preload Module",
			preload: load2,
			script: `
foo.l.append(600)
print(type(foo), dir(foo), foo, foo.l)
`,
			wantErr: false,
			checkData: func(sd starlark.StringDict) bool {
				return sd["l"].(*starlark.List).Len() == 2
			},
		},
		{
			name:    "Assign Preload Struct",
			preload: load2,
			script: `
print(type(bar), dir(bar), bar)
bar.a = 700
bar.b = 800
bar.s = "new"
`,
			wantErr: true,
		},
		{
			name:    "Modify Preload Struct",
			preload: load3,
			script: `
bar.l.append(900)
print(type(bar), dir(bar), bar)
`,
			wantErr: false,
			checkData: func(sd starlark.StringDict) bool {
				return sd["l"].(*starlark.List).Len() == 2
			},
		},

		// Edit Lazyload
		{
			name:       "Edit Lazyload StringDict",
			lazyload:   load1,
			moduleName: "play",
			script: `
load("play", "a", "b", "l", "s")
a = 100
b = 200
s = "new"
`,
			wantErr: true,
		},
		{
			name:       "Modify Lazyload StringDict",
			lazyload:   load1,
			moduleName: "play",
			script: `
load("play", "a", "b", "l", "s")
l.append(300)
`,
			wantErr: false,
			checkData: func(sd starlark.StringDict) bool {
				return sd["a"] == starlark.MakeInt(1) && sd["b"] == starlark.MakeInt(2) && sd["l"].(*starlark.List).Len() == 2 && sd["s"] == starlark.String("test")
			},
		},
		{
			name:       "Assign Lazyload Module",
			lazyload:   load2,
			moduleName: "play",
			script: `
load("play", "foo")
print(type(foo), dir(foo), foo)
foo.a = 400
foo.b = 500
foo.s = "new"
`,
			wantErr: true,
		},
		{
			name:       "Assign Named Lazyload Module",
			lazyload:   load2,
			moduleName: "foo",
			script: `
load("foo", "a", "b", "s")
a = 400
b = 500
s = "new"
`,
			wantErr: true,
		},
		{
			name:       "Modify Lazyload Module",
			lazyload:   load2,
			moduleName: "play",
			script: `
load("play", "foo")
foo.l.append(600)
print(type(foo), dir(foo), foo, foo.l)
`,
			wantErr: false,
			checkData: func(sd starlark.StringDict) bool {
				return sd["l"].(*starlark.List).Len() == 2
			},
		},
		{
			name:       "Modify Named Lazyload Module",
			lazyload:   load2,
			moduleName: "foo",
			script: `
load("foo", "l")
l.append(600)
`,
			wantErr: false,
			checkData: func(sd starlark.StringDict) bool {
				return sd["l"].(*starlark.List).Len() == 2
			},
		},
		{
			name:       "Assign Lazyload Struct",
			lazyload:   load3,
			moduleName: "play",
			script: `
load("play", "bar")
print(type(bar), dir(bar), bar)
bar.a = 700
bar.b = 800
bar.s = "new"
`,
			wantErr: true,
		},
		{
			name:       "Assign Named Lazyload Struct",
			lazyload:   load3,
			moduleName: "bar",
			script: `
load("bar", "a", "b", "s")
a = 700
b = 800
s = "new"
`,
			wantErr: true,
		},
		{
			name:       "Modify Lazyload Struct",
			lazyload:   load3,
			moduleName: "play",
			script: `
load("play", "bar")
bar.l.append(900)
print(type(bar), dir(bar), bar)
`,
			wantErr: false,
			checkData: func(sd starlark.StringDict) bool {
				return sd["l"].(*starlark.List).Len() == 2
			},
		},
		{
			name:       "Modify Named Lazyload Struct",
			lazyload:   load3,
			moduleName: "bar",
			script: `
load("bar", "l")
l.append(900)
`,
			wantErr: false,
			checkData: func(sd starlark.StringDict) bool {
				return sd["l"].(*starlark.List).Len() == 2
			},
		},

		// Add Preload
		{
			name:    "Add Preload Module",
			preload: load2,
			script: `
foo["c"] = 30
`,
			wantErr: true,
		},
		{
			name:    "Add Preload Struct",
			preload: load3,
			script: `
bar["c"] = 30
`,
			wantErr: true,
		},
		{
			name:       "Add Lazyload Module",
			lazyload:   load2,
			moduleName: "play",
			script: `
load("play", "foo")
foo["c"] = 30
`,
			wantErr: true,
		},
		{
			name:       "Add Lazyload Struct",
			lazyload:   load3,
			moduleName: "play",
			script: `
load("play", "bar")
bar["c"] = 30
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// get loaders
			var (
				coreSD   starlark.StringDict
				preload  starlet.ModuleLoaderList
				lazyload starlet.ModuleLoaderMap
			)
			if tt.preload != nil {
				sd, load := tt.preload()
				coreSD = sd
				preload = append(preload, load)
			} else if tt.lazyload != nil {
				sd, load := tt.lazyload()
				coreSD = sd
				lazyload = starlet.ModuleLoaderMap{
					tt.moduleName: load,
				}
			} else {
				t.Errorf("no preload or lazyload")
				return
			}

			// set code
			m := starlet.NewWithLoaders(nil, preload, lazyload)
			m.SetPrintFunc(getLogPrintFunc(t))
			code := tt.script

			// run
			out, err := m.RunScript([]byte(code), nil)

			// check error
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.wantErr {
				return
			}

			// check result
			if out == nil {
				t.Errorf("unexpected nil output")
				return
			}
			if tt.checkRes != nil && !tt.checkRes(out) {
				t.Errorf("unexpected output: %v", out)
			}

			// check original data
			if tt.checkData != nil && !tt.checkData(coreSD) {
				t.Errorf("unexpected original data: %v", coreSD)
			}
		})
	}

	/*
		Things Learned:

		For simple direct StringDict loader:
			When it loads as preload, the original key-value(s) is shadow-copied and the copy is used.
				1. assign doesn't affect the original data (like int or string), but modify like append to list does.
			When it loads as lazyload, the original key-value(s) is shadow-copied and the copy is used.
				1. assign fails for re-assignment, but modify like append to list works.

		For loader with Module:
			When it loads as preload, the starlarkstruct.Module is used directly.
				1. assign to the module's member fails, but modify like append to list works.
				2. add new member fails.
			When it loads as lazyload with different name, the starlarkstruct.Module is used directly.
				1. assign to the module's member fails, but modify like append to list works.
				2. add new member fails.
			When it loads as lazyload with same name, the member of starlarkstruct.Module is used.
				1. assign fails for re-assignment, but modify like append to list works.

		For loader with Struct:
			When it loads as preload, the starlarkstruct.Struct is used directly.
				1. assign to the struct's member fails, but modify like append to list works.
				2. add new member fails.
			When it loads as lazyload with different name, the starlarkstruct.Struct is used directly.
				1. assign to the struct's member fails, but modify like append to list works.
				2. add new member fails.
			When it loads as lazyload with same name, the member of starlarkstruct.Struct is shadow-copied and the copy is used.
				1. assign fails for re-assignment, but modify like append to list works.

		Summary:

		| Loader Type             | Loading Method            | Loading Behavior       | Assign                     | Modify                      | Insert   |
		|:-----------------------:|:--------------------------|:-----------------------|:---------------------------|:----------------------------|:---------|
		| Plain StringDict Loader | Preload                   | Shadow Copy            | ✅ Doesn't affect original | ⚠️ Works (e.g. list append) | N/A      |
		| Plain StringDict Loader | Lazyload                  | Shadow Copy            | ❌ Fails for re-assignment | ⚠️ Works (e.g. list append) | N/A      |
		|      Module Loader      | Preload                   | Direct Usage           | ❌ Fails                   | ⚠️ Works (e.g. list append) | ❌ Fails |
		|      Module Loader      | Lazyload (different name) | Direct Usage           | ❌ Fails                   | ⚠️ Works (e.g. list append) | ❌ Fails |
		|      Module Loader      | Lazyload (same name)      | Shadow Copy of Members | ❌ Fails for re-assignment | ⚠️ Works (e.g. list append) | N/A      |
		|      Struct Loader      | Preload                   | Direct Usage           | ❌ Fails                   | ⚠️ Works (e.g. list append) | ❌ Fails |
		|      Struct Loader      | Lazyload (different name) | Direct Usage           | ❌ Fails                   | ⚠️ Works (e.g. list append) | ❌ Fails |
		|      Struct Loader      | Lazyload (same name)      | Shadow Copy of Members | ❌ Fails for re-assignment | ⚠️ Works (e.g. list append) | N/A      |

		Key:
		- ✅: Successful action
		- ❌: Failed action
		- ⚠️: Action has side effects
		- N/A: Not applicable
	*/
}

func TestWrapModuleData_Edit(t *testing.T) {
	name := "test_module"
	data := starlark.StringDict{
		"foo": starlark.String("bar"),
		"baz": starlark.MakeInt(42),
	}

	wrapFunc := dataconv.WrapModuleData(name, data)
	if _, err := wrapFunc(); err != nil {
		t.Errorf("WrapModuleData() returned an error: %v", err)
	}

	scripts := []string{
		itn.HereDoc(`
			test_module.foo = "bar bar"
		`),
		itn.HereDoc(`
			test_module.baz = 84
		`),
		itn.HereDoc(`
			test_module.qux = "quux"
		`),
		itn.HereDoc(`
			test_module["see"] = "saw"
		`),
	}
	for i, s := range scripts {
		sl := starlet.NewDefault()
		sl.AddPreloadModules(starlet.ModuleLoaderList{wrapFunc})
		if _, err := sl.RunScript([]byte(s), nil); err == nil {
			t.Errorf("Expected error, got nil: %d", i)
		}
	}
}

func TestWrapStructData_Edit(t *testing.T) {
	name := "test_struct"
	data := starlark.StringDict{
		"foo": starlark.String("bar"),
		"baz": starlark.MakeInt(42),
	}

	wrapFunc := dataconv.WrapStructData(name, data)
	if _, err := wrapFunc(); err != nil {
		t.Errorf("WrapStructData() returned an error: %v", err)
	}

	scripts := []string{
		itn.HereDoc(`
			test_struct.foo = "bar bar"
		`),
		itn.HereDoc(`
			test_struct.baz = 84
		`),
		itn.HereDoc(`
			test_struct.qux = "quux"
		`),
		itn.HereDoc(`
			test_struct["see"] = "saw"
		`),
	}
	for i, s := range scripts {
		sl := starlet.NewDefault()
		sl.AddPreloadModules(starlet.ModuleLoaderList{wrapFunc})
		if _, err := sl.RunScript([]byte(s), nil); err == nil {
			t.Errorf("Expected error, got nil: %d", i)
		}
	}
}
