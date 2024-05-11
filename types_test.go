package starlet

import (
	"reflect"
	"testing"

	"go.starlark.net/starlark"
)

func TestStringAnyMap_Clone(t *testing.T) {
	tests := []struct {
		name string
		d    StringAnyMap
		want StringAnyMap
	}{
		{
			name: "test_nil_map_clone",
			d:    nil,
			want: make(StringAnyMap),
		},
		{
			name: "test_empty_map_clone",
			d:    make(StringAnyMap),
			want: make(StringAnyMap),
		},
		{
			name: "test_non_empty_map_clone",
			d:    StringAnyMap{"key1": "val1", "key2": 2, "key3": true},
			want: StringAnyMap{"key1": "val1", "key2": 2, "key3": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Clone(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Clone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringAnyMap_Merge(t *testing.T) {
	tests := []struct {
		name       string
		d          StringAnyMap
		other      StringAnyMap
		want       StringAnyMap
		wantEffect bool
	}{
		{
			name:       "test_nil_merge",
			d:          nil,
			other:      StringAnyMap{"key1": "val1"},
			wantEffect: false,
		},
		{
			name:       "test_merge_map",
			d:          StringAnyMap{"key1": "val1", "key2": 2},
			other:      StringAnyMap{"key2": "val2", "key3": true},
			want:       StringAnyMap{"key1": "val1", "key2": "val2", "key3": true},
			wantEffect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := tt.d.Clone()
			if tt.d == nil {
				orig = nil
			}
			// do the merge and check the result
			tt.d.Merge(tt.other)
			if tt.wantEffect && !reflect.DeepEqual(tt.d, tt.want) {
				t.Errorf("Merge() got = %v, want %v", tt.d, tt.want)
			} else if !tt.wantEffect && !reflect.DeepEqual(tt.d, orig) {
				t.Errorf("Merge() got = %v, want original map: %v", tt.d, orig)
			}
		})
	}
}

// Assuming starlark.StringDict is equivalent to map[string]interface{} for this example.
func TestStringAnyMap_MergeDict(t *testing.T) {
	tests := []struct {
		name       string
		d          StringAnyMap
		other      starlark.StringDict
		want       StringAnyMap
		wantEffect bool
	}{
		{
			name:       "test_nil_merge_dict",
			d:          nil,
			other:      starlark.StringDict{"key1": starlark.String("val1")},
			wantEffect: false,
		},
		{
			name:       "test_merge_dict",
			d:          StringAnyMap{"key1": "val1", "key2": 2},
			other:      starlark.StringDict{"key2": starlark.String("val2"), "key3": starlark.Bool(true)},
			want:       StringAnyMap{"key1": "val1", "key2": starlark.String("val2"), "key3": starlark.Bool(true)},
			wantEffect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := tt.d.Clone()
			if tt.d == nil {
				orig = nil
			}
			// do the merge and check the result
			tt.d.MergeDict(tt.other)
			if tt.wantEffect && !reflect.DeepEqual(tt.d, tt.want) {
				t.Errorf("MergeDict() got = %v, want %v", tt.d, tt.want)
			} else if !tt.wantEffect && !reflect.DeepEqual(tt.d, orig) {
				t.Errorf("MergeDict() got = %v, want original map: %v", tt.d, orig)
			}
		})
	}
}
