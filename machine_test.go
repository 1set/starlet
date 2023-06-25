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
	l, err := starlet.MakeBuiltinModuleLoaderMap("time", "math")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	g := starlet.StringAnyMap{"x": 4}
	m := starlet.NewWithNames(g, []string{"json"}, []string{"math", "time"})
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

func TestNewWithNames_PreNotExist(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, got no panic")
		}
	}()
	g := starlet.StringAnyMap{"x": 5}
	_ = starlet.NewWithNames(g, []string{"json", "not-exist"}, []string{"math", "time"})
}

func TestNewWithNames_LazyNotExist(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, got no panic")
		}
	}()
	g := starlet.StringAnyMap{"x": 6}
	_ = starlet.NewWithNames(g, []string{"json"}, []string{"math", "time", "not-exist"})
}

func TestMachine_Field_Globals(t *testing.T) {
	g := starlet.StringAnyMap{"x": 7}
	g2 := starlet.StringAnyMap{"y": 8}
	m := starlet.NewDefault()
	// empty
	if gg := m.GetGlobals(); len(gg) != 0 {
		t.Errorf("expected empty globals, got %v", gg)
	}
	// empty set
	m.SetGlobals(g)
	if gg := m.GetGlobals(); !expectEqualStringAnyMap(t, gg, g) {
		return
	}
	// non-empty set
	m.SetGlobals(nil)
	if gg := m.GetGlobals(); len(gg) != 0 {
		t.Errorf("expected empty globals, got %v", gg)
	}
	// empty add
	m.AddGlobals(g)
	if gg := m.GetGlobals(); !expectEqualStringAnyMap(t, gg, g) {
		return
	}
	// non-empty add
	m.AddGlobals(g2)
	if gg := m.GetGlobals(); !expectEqualStringAnyMap(t, gg, starlet.StringAnyMap{"x": 7, "y": 8}) {
		return
	}
}

func TestMachine_Field_PreloadModules(t *testing.T) {
	p1, err := starlet.MakeBuiltinModuleLoaderList("json")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	p2, err := starlet.MakeBuiltinModuleLoaderList("math")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	p3, err := starlet.MakeBuiltinModuleLoaderList("time")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	m := starlet.NewDefault()
	// empty
	if pp := m.GetPreloadModules(); len(pp) != 0 {
		t.Errorf("expected empty preload modules, got %v", pp)
	}
	// empty add
	m.AddPreloadModules(p3)
	if pp := m.GetPreloadModules(); !expectEqualModuleList(t, pp, p3) {
		return
	}
	// set
	m.SetPreloadModules(p1)
	if pp := m.GetPreloadModules(); !expectEqualModuleList(t, pp, p1) {
		return
	}
	// add
	m.AddPreloadModules(p2)
	if pp := m.GetPreloadModules(); !expectEqualModuleList(t, pp, append(p1, p2...)) {
		return
	}
}

func TestMachine_Field_LazyloadModules(t *testing.T) {
	l1, err := starlet.MakeBuiltinModuleLoaderMap("json")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	l2, err := starlet.MakeBuiltinModuleLoaderMap("math")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	l3, err := starlet.MakeBuiltinModuleLoaderMap("time")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	m := starlet.NewDefault()
	// empty
	if ll := m.GetLazyloadModules(); len(ll) != 0 {
		t.Errorf("expected empty lazyload modules, got %v", ll)
	}
	// empty add
	m.AddLazyloadModules(l3)
	if ll := m.GetLazyloadModules(); !expectEqualModuleMap(t, ll, l3) {
		return
	}
	// set
	m.SetLazyloadModules(l1)
	if ll := m.GetLazyloadModules(); !expectEqualModuleMap(t, ll, l1) {
		return
	}
	// add
	m.AddLazyloadModules(l2)
	l2.Merge(l1) // new l2
	if ll := m.GetLazyloadModules(); !expectEqualModuleMap(t, ll, l2) {
		return
	}
}
