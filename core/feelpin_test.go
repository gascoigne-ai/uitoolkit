package core

import (
	"testing"
	"time"
)

func TestBuiltInMultiClickWindowIs400msAnd4px(t *testing.T) {
	tests := []struct {
		name  string
		after time.Duration
		x, y  int
		want  int
	}{
		{"at the edge", 400 * time.Millisecond, 14, 6, 2},
		{"late", 401 * time.Millisecond, 10, 10, 1},
		{"far in x", 100 * time.Millisecond, 15, 10, 1},
		{"far in y", 100 * time.Millisecond, 10, 5, 1},
	}
	for _, test := range tests {
		var sampler Sampler
		start := time.Unix(100, 0)
		sampler.Click(start, 10, 10, Feel{})
		if got := sampler.Click(start.Add(test.after), test.x, test.y, Feel{}); got != test.want {
			t.Errorf("%s: count = %d, want %d", test.name, got, test.want)
		}
	}
}
