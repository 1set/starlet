package starlet_test

import (
	"io"
	"reflect"
	"starlet"
	"strings"
	"testing"

	"go.starlark.net/starlark"
)

var (
	builtinModules = []string{"go_idiomatic", "json", "math", "struct", "time"}
)

func TestListBuiltinModules(t *testing.T) {
	modules := starlet.ListBuiltinModules()

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
			wantErr:       "starlet: cannot load modules into nil dict",
		},
		{
			name:          "nil module loader",
			moduleLoaders: starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic"), nil},
			dict:          make(starlark.StringDict),
			wantErr:       "starlet: nil module loader",
		},
		{
			name:          "invalid module",
			moduleLoaders: starlet.ModuleLoaderList{starlet.GetBuiltinModule("go_idiomatic"), failLoader},
			dict:          make(starlark.StringDict),
			wantErr:       "starlet: failed to load module: invalid module loader",
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
			wantErr:       `nil module loader "unknown"`,
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
			wantErr:      "starlet: module \"non_existent\": module not found",
			expectedSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loaderList, err := starlet.MakeBuiltinModuleLoaderList(tt.modules)
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
			wantErr:      "starlet: module \"non_existent\": module not found",
			expectedSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loaderMap, err := starlet.MakeBuiltinModuleLoaderMap(tt.modules)
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
			loader := starlet.ModuleLoaderFromString(tt.fileName, tt.source, tt.predeclared)
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
			loader := starlet.ModuleLoaderFromReader(tt.fileName, tt.source, tt.predeclared)
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
