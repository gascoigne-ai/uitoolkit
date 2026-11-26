package sections

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

func drawnText(canvas *rendertest.Canvas) string {
	var out strings.Builder
	for _, op := range canvas.Ops() {
		if drawn, ok := op.(rendertest.DrawText); ok {
			out.WriteString(drawn.Text + "\n")
		}
	}
	return out.String()
}

func TestOverlayShowsLastFrameStats(t *testing.T) {
	showcase, ctx, clock := newTestShowcase(t)
	ctx.Stats = &core.DrawStats{}
	if drawn := drawnText(frame(t, showcase, ctx)); strings.Contains(drawn, "mask fills") {
		t.Fatal("the overlay shows while it is off")
	}
	ctx.In = core.Input{Keys: []render.Key{render.KeyF3}}
	frame(t, showcase, ctx)
	ctx.In = core.Input{}
	if !showcase.Overlay {
		t.Fatal("F3 did not turn the overlay on")
	}
	clock.at = clock.at.Add(16 * time.Millisecond)
	frame(t, showcase, ctx)
	last := *ctx.Stats
	if last.Elements == 0 || last.MaskFills == 0 {
		t.Fatalf("the frame counted nothing: %+v", last)
	}
	clock.at = clock.at.Add(20 * time.Millisecond)
	drawn := drawnText(frame(t, showcase, ctx))
	for _, want := range []string{
		"frame 20.0 ms",
		fmt.Sprintf("elements %d", last.Elements),
		fmt.Sprintf("mask fills %d", last.MaskFills),
		fmt.Sprintf("ring fills %d", last.RingFills),
		fmt.Sprintf("widths %d", showcase.stats.Widths),
		fmt.Sprintf("advances %d", showcase.stats.Advances),
	} {
		if !strings.Contains(drawn, want+"\n") {
			t.Errorf("the overlay lacks %q:\n%s", want, drawn)
		}
	}
	if showcase.stats.Elements != last.Elements {
		t.Errorf("kept stats %+v, want those of the last frame %+v", showcase.stats, last)
	}
	ctx.In = core.Input{Keys: []render.Key{render.KeyF3}}
	frame(t, showcase, ctx)
	if showcase.Overlay {
		t.Error("F3 did not turn the overlay off")
	}
}
