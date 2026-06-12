package starlet_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/1set/starlet"
	itn "github.com/1set/starlet/internal"
)

func TestNewMemoryCache(t *testing.T) {
	mc := starlet.NewMemoryCache()
	if mc == nil {
		t.Errorf("NewMemoryCache() = nil; want not nil")
	}
}

func TestMemoryCache_Get(t *testing.T) {
	mc := starlet.NewMemoryCache()
	mc.Set("test", []byte("value"))

	tests := []struct {
		name     string
		mc       *starlet.MemoryCache
		key      string
		wantData []byte
		wantHit  bool
	}{
		{
			name:     "Key exists",
			mc:       mc,
			key:      "test",
			wantData: []byte("value"),
			wantHit:  true,
		},
		{
			name:     "Key does not exist",
			mc:       mc,
			key:      "nonsense",
			wantData: nil,
			wantHit:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, ok := tt.mc.Get(tt.key)
			if !reflect.DeepEqual(data, tt.wantData) {
				t.Errorf("MemoryCache.Get() got = %v, want data %v", data, tt.wantData)
			}
			if ok != tt.wantHit {
				t.Errorf("MemoryCache.Get() got = %v, want hit %v", ok, tt.wantHit)
			}
		})
	}
}

func TestMemoryCache_Set(t *testing.T) {
	mc := starlet.NewMemoryCache()

	tests := []struct {
		name    string
		mc      *starlet.MemoryCache
		key     string
		value   []byte
		wantErr bool
	}{
		{
			name:    "Valid Key-Value",
			mc:      mc,
			key:     "test",
			value:   []byte("value"),
			wantErr: false,
		},
		{
			name:    "Invalid MemoryCache",
			mc:      &starlet.MemoryCache{},
			key:     "test",
			value:   []byte("value"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mc.Set(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryCache.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// If no error, assert that the value was set correctly
			if !tt.wantErr {
				got, _ := tt.mc.Get(tt.key)
				if !reflect.DeepEqual(got, tt.value) {
					t.Errorf("MemoryCache.Set() = %v, want %v", got, tt.value)
				}
			}
		})
	}
}

var (
	testScriptName  = "test"
	testScriptBytes = []byte(itn.HereDoc(`
		# This is a test script
		a = 10
		b = 20
		def add(x, y):
			return x+y
		c = add(a,b)
	`))
)

func BenchmarkRunNoCache(b *testing.B) {
	m := starlet.NewDefault()
	m.SetScript(testScriptName, testScriptBytes, nil)
	if _, err := m.Run(); err != nil {
		b.Errorf("Run() error = %v", err)
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Run()
	}
}

func BenchmarkRunMemoryCache(b *testing.B) {
	m := starlet.NewDefault()
	m.SetScriptCacheEnabled(true)
	m.SetScript(testScriptName, testScriptBytes, nil)
	if _, err := m.Run(); err != nil {
		b.Errorf("Run() error = %v", err)
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Run()
	}
}

func Test_ScriptCache_PredeclaredMismatch(t *testing.T) {
	src := []byte(`y = x`)
	cache := starlet.NewMemoryCache()

	// machine A knows the predeclared name x and caches the compiled program
	ma := starlet.NewWithNames(map[string]interface{}{"x": 1}, nil, nil)
	ma.SetScriptCache(cache)
	ma.SetScript("shared.star", src, nil)
	if _, err := ma.Run(); err != nil {
		t.Fatalf("machine A expects no error, got: %v", err)
	}

	// machine B does not know x: it must not reuse A's compiled program —
	// the stale hit used to surface as "internal error: predeclared
	// variable x is uninitialized" instead of a clean resolve error
	mb := starlet.NewDefault()
	mb.SetScriptCache(cache)
	mb.SetScript("shared.star", src, nil)
	if _, err := mb.Run(); err == nil || !strings.Contains(err.Error(), "undefined: x") {
		t.Errorf("machine B expects 'undefined: x', got: %v", err)
	}
}

func Test_ScriptCache_PredeclaredMismatchSilent(t *testing.T) {
	// the nastier direction: with a stale cache hit this program ran
	// silently to completion even though it must not compile without x
	src := []byte("def f():\n    return x\ny = 1\n")
	cache := starlet.NewMemoryCache()

	ma := starlet.NewWithNames(map[string]interface{}{"x": 1}, nil, nil)
	ma.SetScriptCache(cache)
	ma.SetScript("shared.star", src, nil)
	if _, err := ma.Run(); err != nil {
		t.Fatalf("machine A expects no error, got: %v", err)
	}

	mb := starlet.NewDefault()
	mb.SetScriptCache(cache)
	mb.SetScript("shared.star", src, nil)
	if _, err := mb.Run(); err == nil || !strings.Contains(err.Error(), "undefined: x") {
		t.Errorf("machine B expects 'undefined: x', got: %v", err)
	}
}

func Test_ScriptCache_DialectMismatch(t *testing.T) {
	src := []byte("i = 0\nwhile i < 3:\n    i += 1\n")
	cache := starlet.NewMemoryCache()

	// machine A compiles the while-dialect program and caches it
	ma := starlet.NewDefault()
	ma.EnableGlobalReassign()
	ma.SetScriptCache(cache)
	ma.SetScript("shared.star", src, nil)
	if _, err := ma.Run(); err != nil {
		t.Fatalf("machine A expects no error, got: %v", err)
	}

	// machine B has the default dialect: a stale hit used to let the
	// while-program run, silently bypassing B's own dialect settings
	mb := starlet.NewDefault()
	mb.SetScriptCache(cache)
	mb.SetScript("shared.star", src, nil)
	if _, err := mb.Run(); err == nil || !strings.Contains(err.Error(), "while") {
		t.Errorf("machine B expects a while-dialect error, got: %v", err)
	}
}
