package internal

import (
	"testing"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func TestNewAssertLoader(t *testing.T) {
	// Dummy ModuleLoadFunc that returns a fixed StringDict
	moduleLoadFunc := func() (starlark.StringDict, error) {
		return starlark.StringDict{
			"dummy": starlarkstruct.FromStringDict(starlark.String("dummy"), starlark.StringDict{
				"foo": starlark.String("bar"),
			}),
		}, nil
	}

	tests := []struct {
		name          string
		moduleName    string
		loadFunc      ModuleLoadFunc
		expectedError string
	}{
		{
			name:       "Load dummy module",
			moduleName: "dummy",
			loadFunc:   moduleLoadFunc,
		},
		{
			name:       "Load struct.star module",
			moduleName: "struct.star",
			loadFunc:   moduleLoadFunc,
		},
		{
			name:       "Load assert.star module",
			moduleName: "assert.star",
			loadFunc:   moduleLoadFunc,
		},
		{
			name:          "Invalid module",
			moduleName:    "invalid",
			loadFunc:      moduleLoadFunc,
			expectedError: "invalid module",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loader := NewAssertLoader(test.moduleName, test.loadFunc)
			thread := &starlark.Thread{Name: "main"}
			_, err := loader(thread, test.moduleName)

			if err != nil && err.Error() != test.expectedError {
				t.Errorf("expected error %s, got %s", test.expectedError, err)
			}
		})
	}
}
