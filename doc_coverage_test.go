package starlet_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/1set/starlet"
	"go.starlark.net/starlarkstruct"
)

// libReadmeDir maps a builtin module name to its lib/<dir> documentation
// directory. The only non-identity case is go_idiomatic -> goidiomatic;
// modules backed by go.starlark.net (math, struct, time) have no lib README
// and are skipped below.
func libReadmeDir(module string) string {
	return strings.ReplaceAll(module, "_", "")
}

// moduleSurface enumerates the script-visible names a module exports, across
// the three registration shapes: a starlarkstruct.Module (its Members), a
// starlarkstruct.Struct (its AttrNames), or a flat StringDict (its keys).
func moduleSurface(t *testing.T, name string) []string {
	loader := starlet.GetBuiltinModule(name)
	if loader == nil {
		return nil
	}
	sd, err := loader()
	if err != nil {
		t.Fatalf("load module %q: %v", name, err)
	}
	var out []string
	for k, v := range sd {
		switch m := v.(type) {
		case *starlarkstruct.Module:
			for mk := range m.Members {
				out = append(out, mk)
			}
		case *starlarkstruct.Struct:
			out = append(out, m.AttrNames()...)
		default:
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// TestDocCoverage asserts that every script-visible member of every lib/*
// module is documented in that module's README. The matching logic lives in
// tools/doccov/coverage.star and runs through a starlet Machine, so the check
// dogfoods the regex module.
func TestDocCoverage(t *testing.T) {
	script, err := os.ReadFile(filepath.Join("tools", "doccov", "coverage.star"))
	if err != nil {
		t.Fatalf("read coverage script: %v", err)
	}

	surface := map[string]interface{}{}
	docs := map[string]interface{}{}
	var skipped []string
	for _, name := range starlet.GetAllBuiltinModuleNames() {
		readme, err := os.ReadFile(filepath.Join("lib", libReadmeDir(name), "README.md"))
		if err != nil {
			skipped = append(skipped, name) // external module without a lib README
			continue
		}
		docs[name] = string(readme)
		names := moduleSurface(t, name)
		members := make([]interface{}, len(names))
		for i, n := range names {
			members[i] = n
		}
		surface[name] = members
	}

	m := starlet.NewWithNames(starlet.StringAnyMap{"surface": surface, "docs": docs}, nil, []string{"regex"})
	m.SetScriptContent(script)
	out, err := m.Run()
	if err != nil {
		t.Fatalf("doc coverage script failed: %v", err)
	}
	if report, ok := out["report"].(string); ok {
		t.Log("\n" + report)
	}
	sort.Strings(skipped)
	t.Logf("skipped (go.starlark.net modules, no lib README): %v", skipped)

	if missing, ok := out["missing"].([]interface{}); ok && len(missing) > 0 {
		t.Errorf("%d module member(s) are not documented in their README — see the report above", len(missing))
	}
}
