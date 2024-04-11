package atom

import (
	"math"
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
