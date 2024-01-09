package internal

import (
	"fmt"
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
	errLoadFunc := func() (starlark.StringDict, error) {
		return nil, fmt.Errorf("invalid loader")
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
		{
			name:          "Nil module",
			moduleName:    "nil",
			loadFunc:      nil,
			expectedError: "nil module",
		},
		{
			name:          "Error module",
			moduleName:    "error",
			loadFunc:      errLoadFunc,
			expectedError: "invalid loader",
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
