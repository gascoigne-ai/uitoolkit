package rendertest_test

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

var (
	red  = color.RGBA{R: 255, A: 255}
	blue = color.RGBA{B: 255, A: 255}
)

func assertOps(t *testing.T, got, want []rendertest.Op) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d ops, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("op %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}
