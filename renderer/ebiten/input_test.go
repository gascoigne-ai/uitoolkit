package ebitenrender_test

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/renderer/ebiten"
	"github.com/hajimehoshi/ebiten/v2"
)

// A smoke run only: ReadInput is thin, and no test can press a key. What a
// report means is tested in core.
func TestReadInputReportsTheTickRate(t *testing.T) {
	if got, want := ebitenrender.ReadInput().TPS, ebiten.TPS(); got != want {
		t.Errorf("TPS = %d, want %d", got, want)
	}
}
