package ebitenrender

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// Users store shortcuts by ebiten's key names, so every one of them must
// resolve to the render key of the same name.
func TestKeyNamesRoundTripEveryEbitenKey(t *testing.T) {
	hit := map[render.Key]bool{}
	for key := ebiten.Key(0); key <= ebiten.KeyMax; key++ {
		got := KeyFromEbiten(key)
		if got == render.KeyUnknown {
			t.Errorf("KeyFromEbiten(%q) = KeyUnknown", key.String())
			continue
		}
		if got.String() != key.String() {
			t.Errorf("KeyFromEbiten(%q) is named %q", key.String(), got.String())
		}
		back, ok := KeyToEbiten(got)
		if !ok || back != key {
			t.Errorf("KeyToEbiten(%q) = %v, %v, want %v, true", got.String(), back, ok, key)
		}
		hit[got] = true
	}
	for key := render.KeyUnknown + 1; key <= render.KeyMax; key++ {
		if !hit[key] {
			t.Errorf("no ebiten key maps to %q", key.String())
		}
	}
	for _, outside := range []ebiten.Key{-1, ebiten.KeyMax + 1} {
		if got := KeyFromEbiten(outside); got != render.KeyUnknown {
			t.Errorf("KeyFromEbiten(%d) = %q, want KeyUnknown", outside, got.String())
		}
	}
	for _, outside := range []render.Key{render.KeyUnknown, render.KeyMax + 1} {
		if _, ok := KeyToEbiten(outside); ok {
			t.Errorf("KeyToEbiten(%d) reports a key", outside)
		}
	}
}

func TestCursorShapesCoverEveryCursor(t *testing.T) {
	want := map[render.Cursor]ebiten.CursorShapeType{
		render.CursorDefault:    ebiten.CursorShapeDefault,
		render.CursorResizeH:    ebiten.CursorShapeEWResize,
		render.CursorResizeV:    ebiten.CursorShapeNSResize,
		render.CursorResizeNWSE: ebiten.CursorShapeNWSEResize,
		render.CursorResizeNESW: ebiten.CursorShapeNESWResize,
		render.CursorMove:       ebiten.CursorShapeMove,
	}
	if len(cursorShapes) != len(want) {
		t.Errorf("%d rows, want %d", len(cursorShapes), len(want))
	}
	seen := map[ebiten.CursorShapeType]bool{}
	for cursor, shape := range want {
		got, ok := cursorShapes[cursor]
		if !ok || got != shape {
			t.Errorf("cursorShapes[%d] = %v, %v, want %v", cursor, got, ok, shape)
		}
		seen[got] = true
	}
	if len(seen) != len(want) {
		t.Errorf("%d distinct shapes, want %d", len(seen), len(want))
	}
}
