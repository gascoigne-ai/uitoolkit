package ebitenrender

import (
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// EbitenImageOf is the ebiten image a canvas draws for img, so that a pixel
// test draws its reference from that very image.
func EbitenImageOf(img render.Image) *ebiten.Image { return pictureOf(img) }
