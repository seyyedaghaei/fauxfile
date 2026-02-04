package parse

import (
	"testing"
)

func TestBytes(t *testing.T) {
	tests := []struct {
		in   string
		want int64
	}{
		{"1024", 1024},
		{"0", 0},
		{"1k", 1024},
		{"1K", 1024},
		{"1kb", 1024},
		{"1m", 1024 * 1024},
		{"1mb", 1024 * 1024},
		{"1g", 1024 * 1024 * 1024},
		{"10mb", 10 * 1024 * 1024},
		{"1.5mb", int64(1.5 * 1024 * 1024)},
		{"  100  ", 100},
		{"100b", 100},
	}
	for _, tt := range tests {
		got, err := Bytes(tt.in)
		if err != nil {
			t.Errorf("Bytes(%q): %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Bytes(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestBytesErrors(t *testing.T) {
	bad := []string{"", "x", "1x", "1.5.5", "-1", "1tb"}
	for _, s := range bad {
		_, err := Bytes(s)
		if err == nil {
			t.Errorf("Bytes(%q) expected error", s)
		}
	}
}
