package starlet_test

import (
	"reflect"
	"testing"

	"github.com/1set/starlet"
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
