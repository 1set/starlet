package starlet_test

import (
	"fmt"
	"testing"

	"github.com/1set/starlet"
	"go.starlark.net/starlark"
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

func TestMachine_Export_New(t *testing.T) {
	m := starlet.NewDefault()
	ed := m.Export()
	if ed == nil {
		t.Errorf("expected not nil, got nil ExportData")
	}
	if len(ed) != 0 {
		t.Errorf("expected empty, got %v", ed)
	}
}

func TestMachine_Export_Run(t *testing.T) {
	m := starlet.NewDefault()
	g := starlet.StringAnyMap{"x": 9}
	// only set
	m.SetGlobals(g)
	ed := m.Export()
	if ed == nil {
		t.Errorf("expected not nil, got nil ExportData")
		return
	}
	if len(ed) != 0 {
		t.Errorf("expected empty, got %v", ed)
		return
	}
	// run with variables
	rd, err := m.RunScript([]byte(`a = 100`), starlet.StringAnyMap{"y": 10})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
		return
	}
	if !expectEqualStringAnyMap(t, rd, starlet.StringAnyMap{"a": int64(100)}) {
		return
	}
	if len(ed) != 0 {
		t.Errorf("expected unchanged empty, got %v", ed)
		return
	}
	ed = m.Export()
	if !expectEqualStringAnyMap(t, ed, starlet.StringAnyMap{"a": int64(100), "x": int64(9), "y": int64(10)}) {
		return
	}
	// run with load
	ll := starlet.GetBuiltinModuleMap()
	m.SetLazyloadModules(ll)
	rd, err = m.RunScript([]byte(`load("math", "sqrt"); b = sqrt(x)`), nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	ed = m.Export()
	if !expectEqualStringAnyMap(t, ed, starlet.StringAnyMap{"a": int64(100), "b": float64(3), "x": int64(9), "y": int64(10)}) {
		return
	}
	// reset it
	m.Reset()
	ed = m.Export()
	if len(ed) != 0 {
		t.Errorf("expected empty after reset, got %v", ed)
		return
	}
	// add preload modules
	m.AddPreloadModules(starlet.ModuleLoaderList{starlet.GetBuiltinModule("math")})
	rd, err = m.RunScript([]byte(`x = math.sqrt(100)`), nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	ed = m.Export()
	expKeys := []string{"math", "x"}
	if len(ed) != len(expKeys) {
		t.Errorf("expected %d keys, got %d", len(expKeys), len(ed))
		return
	}
	for _, k := range expKeys {
		if _, ok := ed[k]; !ok {
			t.Errorf("expected key %s, got none", k)
			return
		}
	}
}

func TestMachine_DisableInputConversion(t *testing.T) {
	getMac := func(g starlet.StringAnyMap, code string) *starlet.Machine {
		m := starlet.NewWithGlobals(g)
		m.SetInputConversionEnabled(false)
		m.SetScript("test", []byte(code), nil)
		return m
	}

	// nil input
	{
		m := getMac(nil, `a = 1`)
		_, err := m.Run()
		if err != nil {
			t.Errorf("expected no error for nil input, got %v", err)
		}
	}

	// empty input
	{
		m := getMac(starlet.StringAnyMap{}, `a = 1`)
		_, err := m.Run()
		if err != nil {
			t.Errorf("expected no error for empty input, got %v", err)
		}
	}

	// converted
	{
		m := getMac(starlet.StringAnyMap{"a": starlark.MakeInt(100)}, `b = a + 1`)
		res, err := m.Run()
		if err != nil {
			t.Errorf("expected no error for converted input, got %v", err)
		}
		if exp := starlet.StringAnyMap(map[string]interface{}{"b": int64(101)}); !expectEqualStringAnyMap(t, res, exp) {
			t.Errorf("expected result of converted input %v, got %v", exp, res)
			return
		}
	}

	// unconverted -- error
	{
		m := getMac(starlet.StringAnyMap{"a": 100}, `b = a + 1`)
		_, err := m.Run()
		if err == nil {
			t.Errorf("expected error for unconverted input, got none")
		}
	}
}

func TestMachine_DisableOutputConversion(t *testing.T) {
	getMac := func(code string) *starlet.Machine {
		g := starlet.StringAnyMap{"a": 100}
		m := starlet.NewWithGlobals(g)
		m.SetOutputConversionEnabled(false)
		m.SetScript("test", []byte(code), nil)
		return m
	}

	// empty output
	{
		m := getMac(`a + 100`)
		res, err := m.Run()
		if err != nil {
			t.Errorf("expected no error for empty output, got %v", err)
		}
		if len(res) != 0 {
			t.Errorf("expected empty result for empty output, got %v", res)
		}
	}

	// has output
	{
		m := getMac(`b = a << 3`)
		res, err := m.Run()
		if err != nil {
			t.Errorf("expected no error for output, got %v", err)
		}
		if exp := starlet.StringAnyMap(map[string]interface{}{"b": starlark.MakeInt(800)}); !expectEqualStringAnyMap(t, res, exp) {
			t.Errorf("expected result for output %v, got %v", exp, res)
			return
		}
	}

	// for export
	{
		m := getMac(`b = a << 2`)
		res, err := m.Run()
		if err != nil {
			t.Errorf("expected no error for output, got %v", err)
		}
		if exp := starlet.StringAnyMap(map[string]interface{}{"b": starlark.MakeInt(400)}); !expectEqualStringAnyMap(t, res, exp) {
			t.Errorf("expected result for output %v, got %v", exp, res)
			return
		}
		ed := m.Export()
		if exp := starlet.StringAnyMap(map[string]interface{}{"a": starlark.MakeInt(100), "b": starlark.MakeInt(400)}); !expectEqualStringAnyMap(t, ed, exp) {
			t.Errorf("expected export for output %v, got %v", exp, ed)
			return
		}
	}
}

func TestMachine_DisableBothConversion(t *testing.T) {
	getMac := func(g starlet.StringAnyMap, code string) *starlet.Machine {
		m := starlet.NewWithGlobals(g)
		m.SetInputConversionEnabled(false)
		m.SetOutputConversionEnabled(false)
		m.SetScript("test", []byte(code), nil)
		return m
	}

	// nil input
	{
		m := getMac(nil, `a = 1`)
		_, err := m.Run()
		if err != nil {
			t.Errorf("expected no error for nil input, got %v", err)
		}
	}

	// empty input
	{
		m := getMac(starlet.StringAnyMap{}, `a = 1`)
		_, err := m.Run()
		if err != nil {
			t.Errorf("expected no error for empty input, got %v", err)
		}
	}

	// converted
	{
		m := getMac(starlet.StringAnyMap{"a": starlark.MakeInt(100)}, `b = a * 2`)
		res, err := m.Run()
		if err != nil {
			t.Errorf("expected no error for converted input, got %v", err)
		}
		if exp := starlet.StringAnyMap(map[string]interface{}{"b": starlark.MakeInt(200)}); !expectEqualStringAnyMap(t, res, exp) {
			t.Errorf("expected result of converted input %v, got %v", exp, res)
			return
		}
	}

	// unconverted -- error
	{
		m := getMac(starlet.StringAnyMap{"a": 100}, `b = a + 1`)
		_, err := m.Run()
		if err == nil {
			t.Errorf("expected error for unconverted input, got none")
		}
	}
}

func TestMachine_SetScriptCache(t *testing.T) {
	var (
		sname    = "test"
		ckey     = fmt.Sprintf("%d:%s", starlark.CompilerVersion, sname)
		script1  = "a = 10"
		script2  = "a = 20"
		script3  = "b ="
		checkRes = func(c int, err error, res starlet.StringAnyMap, expVal int) {
			if err != nil {
				t.Errorf("[case#%d] expected no error, got %v", c, err)
				return
			}
			if exp := starlet.StringAnyMap(map[string]interface{}{"a": int64(expVal)}); !expectEqualStringAnyMap(t, res, exp) {
				t.Errorf("[case#%d] expected result %v, got %v", c, exp, res)
				return
			}
		}
	)

	// no cache
	{
		m := starlet.NewDefault()
		m.SetScript(sname, []byte(script1), nil)
		res, err := m.Run()
		checkRes(101, err, res, 10)

		m.SetScript(sname, []byte(script2), nil)
		res, err = m.Run()
		checkRes(102, err, res, 20)

		m.Reset()
		m.SetScript(sname, []byte(script2), nil)
		res, err = m.Run()
		checkRes(103, err, res, 20)
	}

	// builtin cache
	{
		m := starlet.NewDefault()
		m.SetScriptCacheEnabled(true)
		m.SetScript(sname, []byte(script1), nil)
		res, err := m.Run()
		checkRes(201, err, res, 10)

		m.SetScript(sname, []byte(script2), nil)
		res, err = m.Run()
		checkRes(202, err, res, 10)

		m.SetScriptCacheEnabled(false)
		m.SetScript(sname, []byte(script2), nil)
		res, err = m.Run()
		checkRes(203, err, res, 20)
	}

	// outside cache
	{
		m := starlet.NewDefault()
		mc := starlet.NewMemoryCache()
		m.SetScriptCache(mc)
		m.SetScript(sname, []byte(script1), nil)
		res, err := m.Run()
		checkRes(301, err, res, 10)

		m.SetScript(sname, []byte(script2), nil)
		res, err = m.Run()
		checkRes(302, err, res, 10)

		m.SetScript(sname+"diff", []byte(script2), nil)
		res, err = m.Run()
		checkRes(303, err, res, 20)
	}

	// broken cache
	{
		m := starlet.NewDefault()
		mc := starlet.NewMemoryCache()
		m.SetScriptCache(mc)
		m.SetScript(sname, []byte(script1), nil)
		res, err := m.Run()
		checkRes(401, err, res, 10)

		m.SetScript(sname, []byte(script2), nil)
		res, err = m.Run()
		checkRes(402, err, res, 10)

		mc.Set(ckey, []byte("broken_data"))
		m.SetScript(sname, []byte(script2), nil)
		res, err = m.Run()
		checkRes(403, err, res, 20)
	}

	// broken code
	{
		m1 := starlet.NewDefault()
		m1.SetScript(sname, []byte(script3), nil)
		_, err1 := m1.Run()
		if err1 == nil {
			t.Errorf("expected error 1, got none")
			return
		}

		m2 := starlet.NewDefault()
		m2.SetScriptCacheEnabled(true)
		m2.SetScript(sname, []byte(script3), nil)
		_, err2 := m2.Run()
		if err2 == nil {
			t.Errorf("expected error 2, got none")
			return
		}

		m2.SetScript(sname, []byte(script3), nil)
		_, err3 := m2.Run()
		if err3 == nil {
			t.Errorf("expected error 3, got none")
			return
		}

		if err1.Error() != err2.Error() || err2.Error() != err3.Error() {
			t.Errorf("expected same error, got different --- err1: %v, err2: %v, err3: %v", err1, err2, err3)
		}
	}

	// ignore cache
	{
		m := starlet.NewDefault()
		m.SetScriptCacheEnabled(true)
		res, err := m.RunScript([]byte(script1), nil)
		checkRes(501, err, res, 10)

		res, err = m.RunScript([]byte(script2), nil)
		checkRes(502, err, res, 20)

		m.SetScriptCacheEnabled(false)
		res, err = m.RunScript([]byte(script2), nil)
		checkRes(503, err, res, 20)
	}
}

func TestMachine_GetStarlarkPredeclared(t *testing.T) {
	m := starlet.NewDefault()
	pd := m.GetStarlarkPredeclared()
	if pd != nil {
		t.Errorf("expected nil, got %v", pd)
	}
	_, err := m.RunScript([]byte(`a = 1 + 2`), nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	pd = m.GetStarlarkPredeclared()
	if pd == nil {
		t.Errorf("expected not nil, got nil")
	} else {
		if _, ok := pd["a"]; !ok {
			t.Errorf("expected 'a' in predeclared, got none")
		}
	}
}

func TestMachine_GetStarlarkThread(t *testing.T) {
	m := starlet.NewDefault()
	th := m.GetStarlarkThread()
	if th != nil {
		t.Errorf("expected nil, got %v", th)
	}
	_, err := m.RunScript([]byte(`a = 1 + 2`), nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	th = m.GetStarlarkThread()
	if th == nil {
		t.Errorf("expected not nil, got nil")
	}
}

func Test_SetPrintFunc(t *testing.T) {
	m := starlet.NewDefault()
	// set print function
	printFunc, cmpFunc := getPrintCompareFunc(t)
	m.SetPrintFunc(printFunc)
	// set code
	code := `print("Aloha, Honua!")`
	m.SetScript("aloha.star", []byte(code), nil)
	// run
	_, err := m.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// compare
	cmpFunc("Aloha, Honua!\n")

	// set print function to nil
	m.SetPrintFunc(nil)
	m.SetScript("aloha.star", []byte(`print("Aloha, Honua! Nil")`), nil)
	if _, err = m.Run(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// set print function to noop
	m.SetPrintFunc(starlet.NoopPrintFunc)
	m.SetScript("aloha.star", []byte(`print("Aloha, Honua! Noop")`), nil)
	if _, err = m.Run(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
