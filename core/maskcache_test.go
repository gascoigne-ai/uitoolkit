package core

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

func maskCtx(t *testing.T) *Ctx {
	t.Helper()
	return newTestCtx(t, &otherDevice{font: rendertest.FixedFont{AdvanceEm: 0.5, AscentEm: 0.8, DescentEm: 0.2}})
}

func TestMaskIsPaintedOnceAndThenReturned(t *testing.T) {
	ctx := maskCtx(t)
	paints := 0
	paint := func(render.Canvas) { paints++ }
	key := MaskKey{Source: new(int), Scale: 1, Weight: 1}
	first := ctx.Mask(key, 10, 12, paint)
	again := ctx.Mask(key, 10, 12, paint)
	if first != again || paints != 1 {
		t.Fatalf("same key: same mask %v, paints %d, want true and 1", first == again, paints)
	}
	if got := first.Bounds(); got.Dx() != 10 || got.Dy() != 12 {
		t.Errorf("mask size %v, want 10x12", got)
	}
	other := ctx.Mask(MaskKey{Source: key.Source, Scale: 2, Weight: 1}, 20, 24, paint)
	if other == first || paints != 2 {
		t.Errorf("a new scale is a new mask: paints %d", paints)
	}
}

func TestMaskCacheIsBoundedAndDropsWholeReleasingWhatItDrops(t *testing.T) {
	ctx := maskCtx(t)
	paint := func(render.Canvas) {}
	first := ctx.Mask(MaskKey{Source: 0}, 1, 1, paint).(*rendertest.Offscreen)
	for i := 1; i < maskLimit; i++ {
		ctx.Mask(MaskKey{Source: i}, 1, 1, paint)
	}
	if ctx.MaskCount() != maskLimit {
		t.Fatalf("held %d masks, want %d", ctx.MaskCount(), maskLimit)
	}
	ctx.Mask(MaskKey{Source: maskLimit}, 1, 1, paint)
	if ctx.MaskCount() != 1 {
		t.Errorf("held %d masks after the limit, want 1: the cache is dropped whole", ctx.MaskCount())
	}
	assertReleased(t, first)
}

func TestSetScaleAndSetThemeDropTheMasks(t *testing.T) {
	paint := func(render.Canvas) {}
	tests := []struct {
		name  string
		apply func(ctx *Ctx) error
	}{
		{"SetScale", func(ctx *Ctx) error { return ctx.SetScale(2) }},
		{"SetTheme", func(ctx *Ctx) error { return ctx.SetTheme(testTheme()) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := maskCtx(t)
			mask := ctx.Mask(MaskKey{Source: 1}, 1, 1, paint).(*rendertest.Offscreen)
			if err := tc.apply(ctx); err != nil {
				t.Fatal(err)
			}
			if ctx.MaskCount() != 0 {
				t.Errorf("%d masks left", ctx.MaskCount())
			}
			assertReleased(t, mask)
		})
	}
}

func assertReleased(t *testing.T, mask *rendertest.Offscreen) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("the dropped mask was not released")
		}
	}()
	mask.Clear()
}
