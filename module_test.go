package starlet_test

import (
	"reflect"
	"starlet"
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
