package starlet_test

import (
	"github.com/1set/starlet"
	"reflect"
	"testing"
)

func TestNewDefault(t *testing.T) {
	t.Skip()
	m := starlet.NewDefault()
	if m == nil {
		t.Errorf("expected not nil, got nil machine")
	}
	// check the rest of the fields
	if g := m.GetGlobals(); g != nil {
		t.Errorf("expected nil globals, got %v", g)
	}
	if p := m.GetPreloadModules(); len(p) != 0 {
		t.Errorf("expected empty preload modules, got %v", p)
	}
	if l := m.GetLazyloadModules(); len(l) != 0 {
		t.Errorf("expected empty lazyload modules, got %v", l)
	}
}

func TestNewWithGlobals(t *testing.T) {
	t.Skip()
	g := starlet.StringAnyMap{"x": 1}
	m := starlet.NewWithGlobals(g)
	if m == nil {
		t.Errorf("expected not nil, got nil machine")
	}
	if gg := m.GetGlobals(); !reflect.DeepEqual(g, gg) {
		t.Errorf("expected %v, got %v", g, gg)
	}
	// check the rest of the fields
	if p := m.GetPreloadModules(); len(p) != 0 {
		t.Errorf("expected empty preload modules, got %v", p)
	}
	if l := m.GetLazyloadModules(); len(l) != 0 {
		t.Errorf("expected empty lazyload modules, got %v", l)
	}
}

func TestNewWithLoaders(t *testing.T) {
	t.Skip()
	p, err := starlet.MakeBuiltinModuleLoaderList("json")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	l, err := starlet.MakeBuiltinModuleLoaderMap("time")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	g := starlet.StringAnyMap{"x": 1}
	m := starlet.NewWithLoaders(g, p, l)
	if m == nil {
		t.Errorf("expected not nil, got nil machine")
	}
	if gg := m.GetGlobals(); !reflect.DeepEqual(g, gg) {
		t.Errorf("expected %v, got %v", g, gg)
	}
	if pp := m.GetPreloadModules(); !reflect.DeepEqual(p, pp) {
		t.Errorf("expected %v, got %v", p, pp)
	}
	if ll := m.GetLazyloadModules(); !reflect.DeepEqual(l, ll) {
		t.Errorf("expected %v, got %v", l, ll)
	}
}
