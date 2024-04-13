package atom

import (
	"math"
	"strings"
	"testing"
)

func Test_hashInt64(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  uint32
	}{
		{
			name:  "zero",
			value: 0,
			want:  2615243109,
		},
		{
			name:  "+1",
			value: 1,
			want:  1048580676,
		},
		{
			name:  "-1",
			value: -1,
			want:  1823345245,
		},
		{
			name:  "max",
			value: math.MaxInt64,
			want:  3970880477,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hashInt64(tt.value); got != tt.want {
				t.Errorf("hashInt64(%d) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func Test_hashFloat64(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  uint32
	}{
		{
			name:  "zero",
			value: 0,
			want:  2615243109,
		},
		{
			name:  "+1",
			value: 1,
			want:  2355796088,
		},
		{
			name:  "-1",
			value: -1,
			want:  208260856,
		},
		{
			name:  "max",
			value: math.MaxFloat64,
			want:  3968320621,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hashFloat64(tt.value); got != tt.want {
				t.Errorf("hashFloat64(%f) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func Test_hashString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  uint32
	}{
		{
			name:  "empty",
			input: "",
			want:  2166136261,
		},
		{
			name:  "single",
			input: "a",
			want:  3826002220,
		},
		{
			name:  "next",
			input: "b",
			want:  3876335077,
		},
		{
			name:  "add",
			input: "ab",
			want:  1294271946,
		},
		{
			name:  "hello",
			input: "hello",
			want:  1335831723,
		},
		{
			name:  "long",
			input: strings.Repeat("this is a long string", 100),
			want:  229378413,
		},
	}
	for _, tt := range tests {
		got := hashString(tt.input)
		if got != tt.want {
			t.Errorf("hashString(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
