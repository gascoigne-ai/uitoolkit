package main

import (
	"fmt"
	"math"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/examples/showcase/sections"
	"github.com/gascoigne-ai/uitoolkit/render"
)

// host is the showcase as a renderer runs it.
type host struct {
	showcase  *sections.Showcase
	ctx       *core.Ctx
	readInput func() render.InputState
	sampler   core.Sampler
	// width and height are the area of the view in logical pixels, as the
	// last Layout found it.
	width, height int
	// scaleErr is why the last Layout could not size the faces; Update
	// returns it, which ends the program.
	scaleErr error
}

var _ render.Host = (*host)(nil)

func newHost(showcase *sections.Showcase, ctx *core.Ctx, readInput func() render.InputState) *host {
	return &host{showcase: showcase, ctx: ctx, readInput: readInput}
}

// Layout implements render.Host. The canvas is the window in device pixels.
// The view draws on it at the display's factor times the chosen scale step,
// so a larger step leaves it fewer logical pixels.
func (h *host) Layout(width, height int, device float64) (int, int) {
	canvasW, canvasH := int(math.Ceil(float64(width)*device)), int(math.Ceil(float64(height)*device))
	scale := device * h.showcase.ScaleFactor()
	if err := h.ctx.SetScale(scale); err != nil {
		h.scaleErr = fmt.Errorf("set scale %v: %w", scale, err)
	}
	h.width, h.height = int(math.Ceil(float64(canvasW)/scale)), int(math.Ceil(float64(canvasH)/scale))
	return canvasW, canvasH
}

// Update implements render.Host.
func (h *host) Update() error {
	if h.scaleErr != nil {
		return h.scaleErr
	}
	h.ctx.Tick++
	h.ctx.In = h.sampler.Sample(h.readInput()).Scaled(h.ctx.Scale)
	return h.showcase.Update(h.ctx, h.width, h.height)
}

// Draw implements render.Host. Every frame is drawn: the overlay shows the
// time between frames.
func (h *host) Draw(screen render.Canvas) bool {
	h.showcase.Draw(h.ctx, h.width, h.height, screen)
	return true
}

// Cursor implements render.Host.
func (h *host) Cursor() render.Cursor { return h.ctx.Cursor }
