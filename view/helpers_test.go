package view

import (
	"image"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/gascoigne-ai/uitoolkit/render/text"
	"github.com/gascoigne-ai/uitoolkit/theming"
)

// Helpers shared by the tests of this package: a context to lay out and
// update views with, a view of sample lines with the pointer helpers that
// press, drag and release on it, and the keyboard helpers for a focused
// code view. A helper used by one test file stays in that file.

func testCtx(t *testing.T) *core.Ctx {
	t.Helper()
	c, err := core.NewCtx(theming.Dark(), testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func box(items ...Item) *El { return newEl(plain{}, items) }

func click(ctx *core.Ctx, posX, posY int) { ctx.In = core.Input{X: posX, Y: posY, Click: true} }

type memClip struct{ text string }

func (m *memClip) Read() (string, error) { return m.text, nil }

func (m *memClip) Write(content string) error { m.text = content; return nil }

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

// NewTestDevice is testDevice for the tests of package view_test.
func NewTestDevice() render.Device { return testDevice{} }
