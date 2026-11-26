package sections

import (
	"image"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/gascoigne-ai/uitoolkit/render/text"
	"github.com/gascoigne-ai/uitoolkit/theming"
)

// testDevice is a device with no GPU: faces measured by render/text, and
// offscreens and images that only record.
type testDevice struct{}

func (testDevice) NewFontSource(data []byte) (render.FontSource, error) {
	return text.NewFontSource(data)
}

func (testDevice) NewOffscreen(width, height int) render.Offscreen {
	return rendertest.NewOffscreen(width, height)
}

func (testDevice) NewImage(source image.Image) render.Image {
	return &rendertest.Image{Rect: source.Bounds()}
}

// fakeClock is a clock the test moves by hand.
type fakeClock struct{ at time.Time }

func (c *fakeClock) now() time.Time { return c.at }

func newClock() *fakeClock {
	return &fakeClock{at: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
}

const testW, testH = 1024, 720

func newTestShowcase(t *testing.T) (*Showcase, *core.Ctx, *fakeClock) {
	t.Helper()
	clock := newClock()
	ctx, err := core.NewCtx(theming.Dark(), testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	ctx.Now, ctx.HostOS = clock.now, "linux"
	return New(clock.now), ctx, clock
}

// open selects the section called id.
func open(t *testing.T, showcase *Showcase, ctx *core.Ctx, id string) {
	t.Helper()
	for i, section := range Sections {
		if section.ID == id {
			showcase.Select(ctx, i)
			return
		}
	}
	t.Fatalf("no section %q", id)
}

// frame runs one tick and draws it.
func frame(t *testing.T, showcase *Showcase, ctx *core.Ctx) *rendertest.Canvas {
	t.Helper()
	ctx.Tick++
	if err := showcase.Update(ctx, testW, testH); err != nil {
		t.Fatal(err)
	}
	canvas := rendertest.NewCanvas(testW, testH)
	showcase.Draw(ctx, testW, testH, canvas)
	return canvas
}
