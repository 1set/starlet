package starlet_test

import (
	"starlet"
	"testing"
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
