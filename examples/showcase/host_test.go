package main

import (
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/examples/showcase/sections"
	"github.com/gascoigne-ai/uitoolkit/render"
)

func newTestHost(t *testing.T, readInput func() render.InputState) *host {
	t.Helper()
	opened, err := openRenderer("ebiten")
	if err != nil {
		t.Fatal(err)
	}
	showcase := sections.New(time.Now)
	ctx, err := core.NewCtx(showcase.Theme(), opened.Device)
	if err != nil {
		t.Fatal(err)
	}
	return newHost(showcase, ctx, readInput)
}

// The canvas stays the window in device pixels whatever the scale step, so
// nothing is resampled; the step shrinks the logical area the view has.
func TestLayoutScalesByDeviceFactorAndStep(t *testing.T) {
	tests := []struct {
		name               string
		device             float64
		step               int
		canvasW, canvasH   int
		scale              float64
		logicalW, logicalH int
	}{
		{"plain", 1, 0, 1024, 720, 1, 1024, 720},
		{"high-density display", 2, 0, 2048, 1440, 2, 1024, 720},
		{"step on a fractional display", 1.5, 3, 1536, 1080, 3, 512, 360},
		{"the last logical pixel may be cut", 2, 1, 2048, 1440, 2.5, 820, 576},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newTestHost(t, func() render.InputState { return render.InputState{X: test.canvasW / 2, Y: test.canvasH / 2} })
			h.showcase.SetScale(test.step)
			canvasW, canvasH := h.Layout(1024, 720, test.device)
			if canvasW != test.canvasW || canvasH != test.canvasH {
				t.Errorf("canvas = %dx%d, want %dx%d", canvasW, canvasH, test.canvasW, test.canvasH)
			}
			if h.ctx.Scale != test.scale {
				t.Errorf("ctx.Scale = %v, want %v", h.ctx.Scale, test.scale)
			}
			if h.width != test.logicalW || h.height != test.logicalH {
				t.Errorf("logical area = %dx%d, want %dx%d", h.width, h.height, test.logicalW, test.logicalH)
			}
			if err := h.Update(); err != nil {
				t.Fatal(err)
			}
			if want := int(float64(test.canvasW/2) / test.scale); h.ctx.In.X != want || h.ctx.Tick != 1 {
				t.Errorf("pointer x = %d, tick %d, want %d and 1", h.ctx.In.X, h.ctx.Tick, want)
			}
		})
	}
}
