package internal

import "testing"

func TestGetBytesMD5(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "Test 0",
			data: []byte(""),
			want: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name: "Test 1",
			data: []byte("test"),
			want: "098f6bcd4621d373cade4e832627b4f6",
		},
		{
			name: "Test 2",
			data: []byte("hello world"),
			want: "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
		{
			name: "Test 3",
			data: []byte("The quick brown fox jumps over the lazy dog"),
			want: "9e107d9d372bb6826bd81d3542a419d6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBytesMD5(tt.data); got != tt.want {
				t.Errorf("GetBytesMD5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStringMD5(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "Test 0",
			data: "",
			want: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name: "Test 1",
			data: "test",
			want: "098f6bcd4621d373cade4e832627b4f6",
		},
		{
			name: "Test 2",
			data: "hello world",
			want: "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
		{
			name: "Test 3",
			data: "The quick brown fox jumps over the lazy dog",
			want: "9e107d9d372bb6826bd81d3542a419d6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStringMD5(tt.data); got != tt.want {
				t.Errorf("GetStringMD5() = %v, want %v", got, tt.want)
			}
		})
	}
}
