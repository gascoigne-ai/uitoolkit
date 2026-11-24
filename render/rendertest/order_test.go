package rendertest_test

import (
	"image"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

func TestOpsKeepPaintOrderAcrossSubs(t *testing.T) {
	whole := image.Rect(0, 0, 100, 80)
	left, right := image.Rect(0, 0, 50, 80), image.Rect(50, 0, 100, 80)
	canvas := rendertest.NewCanvas(100, 80)
	leftSub, rightSub := canvas.Sub(left), canvas.Sub(right)
	box := render.Box{X0: 1, Y0: 2, X1: 3, Y1: 4}
	var path render.Path
	stroke := render.Stroke{Width: 2, Cap: render.LineCapRound}
	picture := &rendertest.Image{Rect: image.Rect(0, 0, 8, 8)}
	face := render.Face{Source: &rendertest.FixedFont{}, Size: 12}

	canvas.FillRect(box, red)
	leftSub.FillRectGradient(box, red, blue)
	rightSub.FillRoundRect(box, 5, render.CornersTop, red, blue)
	canvas.StrokeRect(box, 1, red)
	leftSub.StrokeRoundRect(box, 5, 1, blue)
	rightSub.StrokeLine(1, 2, 3, 4, 5, red)
	canvas.FillPath(&path, blue)
	leftSub.StrokePath(&path, stroke, red)
	rightSub.DrawImage(picture, 6, 7, 2, render.FilterLinear)
	canvas.DrawMask(picture, 8, 9, blue)
	leftSub.DrawText(face, "text", 10, 11, red)

	// A path without steps records as a nil slice of segments.
	assertOps(t, canvas.Ops(), []rendertest.Op{
		rendertest.FillRect{Clip: whole, Box: box, Color: red},
		rendertest.FillRectGradient{Clip: left, Box: box, Top: red, Bottom: blue},
		rendertest.FillRoundRect{Clip: right, Box: box, Radius: 5, Corners: render.CornersTop, Top: red, Bottom: blue},
		rendertest.StrokeRect{Clip: whole, Box: box, Width: 1, Color: red},
		rendertest.StrokeRoundRect{Clip: left, Box: box, Radius: 5, Width: 1, Color: blue},
		rendertest.StrokeLine{Clip: right, X0: 1, Y0: 2, X1: 3, Y1: 4, Width: 5, Color: red},
		rendertest.FillPath{Clip: whole, Color: blue},
		rendertest.StrokePath{Clip: left, Stroke: stroke, Color: red},
		rendertest.DrawImage{Clip: right, Image: picture, Left: 6, Top: 7, Scale: 2, Filter: render.FilterLinear},
		rendertest.DrawMask{Clip: whole, Mask: picture, Left: 8, Top: 9, Color: blue},
		rendertest.DrawText{Clip: left, Face: face, Text: "text", Left: 10, Top: 11, Color: red},
	})
}
