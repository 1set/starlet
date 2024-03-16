package starlet_test

import (
	"io"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/1set/starlet"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	builtinModules = []string{"base64", "file", "go_idiomatic", "hashlib", "http", "json", "log", "math", "random", "re", "runtime", "string", "struct", "time"}
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
	sdLoad := func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"a": starlark.MakeInt(1),
			"b": starlark.MakeInt(2),
		}, nil
	}
	smLoad := func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"foo": &starlarkstruct.Module{
				Name: "foo",
				Members: starlark.StringDict{
					"a": starlark.MakeInt(1),
					"b": starlark.MakeInt(2),
				},
			},
		}, nil
	}
	ssLoad := func() (starlark.StringDict, error) {
		sd := starlark.StringDict{
			"a": starlark.MakeInt(1),
			"b": starlark.MakeInt(2),
		}
		ss := starlarkstruct.FromStringDict(starlark.String("bar"), sd)
		return starlark.StringDict{
			"bar": ss,
		}, nil
	}

	tests := []struct {
		name     string
		preload  starlet.ModuleLoaderList
		lazyload starlet.ModuleLoaderMap
		script   string
		wantErr  bool
		checkRes func(anyMap starlet.StringAnyMap) bool
	}{
		{
			name:    "Preload StringDict",
			preload: starlet.ModuleLoaderList{sdLoad},
			script: `
print(a, b)
`,
			wantErr: false,
			checkRes: func(anyMap starlet.StringAnyMap) bool {
				return true
			},
		},
		{
			name:    "Preload Module",
			preload: starlet.ModuleLoaderList{smLoad},
			script: `
print(foo.a, foo.b)
`,
			wantErr: false,
			checkRes: func(anyMap starlet.StringAnyMap) bool {
				return true
			},
		},
		{
			name:    "Preload Struct",
			preload: starlet.ModuleLoaderList{ssLoad},
			script: `
print(bar.a, bar.b)
`,
			wantErr: false,
			checkRes: func(anyMap starlet.StringAnyMap) bool {
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set code
			m := starlet.NewWithLoaders(nil, tt.preload, tt.lazyload)
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
			if !tt.checkRes(out) {
				t.Errorf("unexpected output: %v", out)
			}
		})
	}
}
