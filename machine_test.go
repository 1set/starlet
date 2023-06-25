package starlet_test

import (
	"testing"

	"github.com/1set/starlet"
)

func TestNewDefault(t *testing.T) {
	m := starlet.NewDefault()
	if m == nil {
		t.Errorf("expected not nil, got nil machine")
	}
	// check the rest of the fields
	if gg := m.GetGlobals(); len(gg) != 0 {
		t.Errorf("expected empty globals, got %v", gg)
	}
	if pp := m.GetPreloadModules(); len(pp) != 0 {
		t.Errorf("expected empty preload modules, got %v", pp)
	}
	if ll := m.GetLazyloadModules(); len(ll) != 0 {
		t.Errorf("expected empty lazyload modules, got %v", ll)
	}
}

func TestNewWithGlobals(t *testing.T) {
	g := starlet.StringAnyMap{"x": 1}
	m := starlet.NewWithGlobals(g)
	if m == nil {
		t.Errorf("expected not nil, got nil machine")
	}
	if gg := m.GetGlobals(); !expectEqualStringAnyMap(t, gg, g) {
		return
	}
	// check the rest of the fields
	if pp := m.GetPreloadModules(); len(pp) != 0 {
		t.Errorf("expected empty preload modules, got %v", pp)
	}
	if ll := m.GetLazyloadModules(); len(ll) != 0 {
		t.Errorf("expected empty lazyload modules, got %v", ll)
	}
}

func TestNewWithLoaders(t *testing.T) {
	p, err := starlet.MakeBuiltinModuleLoaderList("json")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	l, err := starlet.MakeBuiltinModuleLoaderMap("time")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	g := starlet.StringAnyMap{"x": 2}
	m := starlet.NewWithLoaders(g, p, l)
	if m == nil {
		t.Errorf("expected not nil, got nil machine")
	}
	if gg := m.GetGlobals(); !expectEqualStringAnyMap(t, gg, g) {
		return
	}
	if pp := m.GetPreloadModules(); !expectEqualModuleList(t, pp, p) {
		return
	}
	if ll := m.GetLazyloadModules(); !expectEqualModuleMap(t, ll, l) {
		return
	}
}

func TestNewWithBuiltins(t *testing.T) {
	bp := starlet.GetAllBuiltinModules()
	bl := starlet.GetBuiltinModuleMap()
	g := starlet.StringAnyMap{"x": 3}
	m := starlet.NewWithBuiltins(g, nil, nil)
	if m == nil {
		t.Errorf("expected not nil, got nil machine")
	}
	if gg := m.GetGlobals(); !expectEqualStringAnyMap(t, gg, g) {
		return
	}
	if pp := m.GetPreloadModules(); !expectEqualModuleList(t, pp, bp) {
		return
	}
	if ll := m.GetLazyloadModules(); !expectEqualModuleMap(t, ll, bl) {
		return
	}
}

func TestNewWithNames(t *testing.T) {
	p, err := starlet.MakeBuiltinModuleLoaderList("json")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	l, err := starlet.MakeBuiltinModuleLoaderMap("time")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	g := starlet.StringAnyMap{"x": 4}
	m := starlet.NewWithNames(g, []string{"json"}, []string{"time"})
	if m == nil {
		t.Errorf("expected not nil, got nil machine")
	}
	if gg := m.GetGlobals(); !expectEqualStringAnyMap(t, gg, g) {
		return
	}
	if pp := m.GetPreloadModules(); !expectEqualModuleList(t, pp, p) {
		return
	}
	if ll := m.GetLazyloadModules(); !expectEqualModuleMap(t, ll, l) {
		return
	}
}
