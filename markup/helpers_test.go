package markup

import (
	"image"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/gascoigne-ai/uitoolkit/render/text"
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
