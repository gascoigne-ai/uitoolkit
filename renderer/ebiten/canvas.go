package ebitenrender

import (
	"image"
	"image/color"
	"math"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Canvas is a render.Canvas that draws on an ebiten image. Device.NewCanvas
// and Sub make one; the zero value cannot draw.
type Canvas struct {
	device *Device
	dst    *ebiten.Image
}

var _ render.Canvas = (*Canvas)(nil)

// The render values as ebiten's, indexed by the render value.
var (
	lineCaps = [...]vector.LineCap{
		render.LineCapButt:   vector.LineCapButt,
		render.LineCapRound:  vector.LineCapRound,
		render.LineCapSquare: vector.LineCapSquare,
	}
	lineJoins = [...]vector.LineJoin{
		render.LineJoinMiter: vector.LineJoinMiter,
		render.LineJoinBevel: vector.LineJoinBevel,
		render.LineJoinRound: vector.LineJoinRound,
	}
	filters = [...]ebiten.Filter{
		render.FilterNearest: ebiten.FilterNearest,
		render.FilterLinear:  ebiten.FilterLinear,
	}
)

// Bounds implements render.Canvas.
func (c *Canvas) Bounds() image.Rectangle { return c.dst.Bounds() }

// Sub implements render.Canvas.
func (c *Canvas) Sub(clip image.Rectangle) render.Canvas {
	part := c.dst.SubImage(c.Bounds().Intersect(clip)).(*ebiten.Image)
	return &Canvas{device: c.device, dst: part}
}

// whole reports whether all four edges of box are on whole pixels, so that
// drawing it needs no anti-aliasing.
func whole(box render.Box) bool {
	for _, edge := range [4]float32{box.X0, box.Y0, box.X1, box.Y1} {
		if edge != float32(math.Trunc(float64(edge))) {
			return false
		}
	}
	return true
}

// FillRect implements render.Canvas.
func (c *Canvas) FillRect(box render.Box, clr color.RGBA) {
	vector.FillRect(c.dst, box.X0, box.Y0, box.X1-box.X0, box.Y1-box.Y0, clr, !whole(box))
}

// FillRectGradient implements render.Canvas. The gradient rides on the vertex
// colors of one DrawTriangles.
func (c *Canvas) FillRectGradient(box render.Box, top, bottom color.RGBA) {
	verts, idx := rectPath(box).AppendVerticesAndIndicesForFilling(nil, nil)
	if len(verts) == 0 {
		return
	}
	for i := range verts {
		verts[i].ColorR, verts[i].ColorG, verts[i].ColorB, verts[i].ColorA = gradientAt(top, bottom, box.Y0, box.Y1, verts[i].DstY)
		verts[i].SrcX, verts[i].SrcY = 0.5, 0.5
	}
	c.dst.DrawTriangles(verts, idx, gradientSrc, &ebiten.DrawTrianglesOptions{AntiAlias: !whole(box)})
}

// maskRadius is the radius in whole pixels of the corner mask a rounded shape
// in box uses: radius rounded, at least one, at most half the shorter side.
// Below one the box is too small to round.
func maskRadius(box render.Box, radius float32) float32 {
	rounded := float32(max(1, math.Round(float64(radius))))
	return min(rounded, float32(math.Floor(float64(min(box.X1-box.X0, box.Y1-box.Y0))/2)))
}

// FillRoundRect implements render.Canvas. The corners are quarters of a cached
// anti-aliased disc, so the whole fill is one DrawTriangles. As in core, a
// radius below one pixel or an empty set of corners is the square fill, and a
// box too small for its radius is the square gradient.
func (c *Canvas) FillRoundRect(box render.Box, radius float32, corners render.Corners, top, bottom color.RGBA) {
	if radius < 1 || corners == 0 {
		c.fillSquare(box, top, bottom)
		return
	}
	r := maskRadius(box, radius)
	if r < 1 {
		c.FillRectGradient(box, top, bottom)
		return
	}
	quads := sliceQuads(box.X0, box.Y0, box.X1, box.Y1, r, corners)
	c.drawQuads(quads, c.device.masks.get(int(r)), func(height float32) (red, green, blue, alpha float32) {
		return gradientAt(top, bottom, box.Y0, box.Y1, height)
	})
}

// fillSquare fills box without rounding: one color on the rectangle path, two
// as a gradient.
func (c *Canvas) fillSquare(box render.Box, top, bottom color.RGBA) {
	if top == bottom {
		c.FillRect(box, top)
		return
	}
	c.FillRectGradient(box, top, bottom)
}

// drawQuads draws the quads from mask in one DrawTriangles, each vertex in the
// color tint gives for its height.
func (c *Canvas) drawQuads(quads []quad, mask *ebiten.Image, tint func(height float32) (red, green, blue, alpha float32)) {
	verts := make([]ebiten.Vertex, 0, 4*len(quads))
	idx := make([]uint16, 0, 6*len(quads))
	for _, piece := range quads {
		base := uint16(len(verts))
		for _, corner := range [4][4]float32{
			{piece.dst.X0, piece.dst.Y0, piece.src.X0, piece.src.Y0}, {piece.dst.X1, piece.dst.Y0, piece.src.X1, piece.src.Y0},
			{piece.dst.X1, piece.dst.Y1, piece.src.X1, piece.src.Y1}, {piece.dst.X0, piece.dst.Y1, piece.src.X0, piece.src.Y1},
		} {
			red, green, blue, alpha := tint(corner[1])
			verts = append(verts, ebiten.Vertex{DstX: corner[0], DstY: corner[1], SrcX: corner[2], SrcY: corner[3], ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha})
		}
		idx = append(idx, base, base+1, base+2, base, base+2, base+3)
	}
	c.dst.DrawTriangles(verts, idx, mask, &ebiten.DrawTrianglesOptions{})
}

// StrokeRect implements render.Canvas.
func (c *Canvas) StrokeRect(box render.Box, width float32, clr color.RGBA) {
	antialias := !whole(box) || width != float32(math.Trunc(float64(width)))
	vector.StrokeRect(c.dst, box.X0+width/2, box.Y0+width/2, box.X1-box.X0-width, box.Y1-box.Y0-width, width, clr, antialias)
}

// StrokeRoundRect implements render.Canvas. The outline is one DrawTriangles
// from a cached ring mask, whose line is a whole number of pixels: a fraction
// of width is cut off in the corners.
func (c *Canvas) StrokeRoundRect(box render.Box, radius, width float32, clr color.RGBA) {
	if radius < 1 {
		c.StrokeRect(box, width, clr)
		return
	}
	r := maskRadius(box, radius)
	if r < 1 {
		c.StrokeRect(box, width, clr)
		return
	}
	width = min(width, r)
	quads := ringQuads(box.X0, box.Y0, box.X1, box.Y1, r, width)
	red, green, blue, alpha := float32(clr.R)/255, float32(clr.G)/255, float32(clr.B)/255, float32(clr.A)/255
	c.drawQuads(quads, c.device.masks.getRing(int(r), int(width)), func(float32) (float32, float32, float32, float32) {
		return red, green, blue, alpha
	})
}

// StrokeLine implements render.Canvas.
func (c *Canvas) StrokeLine(x0, y0, x1, y1, width float32, clr color.RGBA) {
	vector.StrokeLine(c.dst, x0, y0, x1, y1, width, clr, true)
}

// vectorPath is path as ebiten's.
func vectorPath(path *render.Path) *vector.Path {
	var converted vector.Path
	for _, segment := range path.Segments() {
		at := segment.Coords
		switch segment.Verb {
		case render.VerbMove:
			converted.MoveTo(at[0], at[1])
		case render.VerbLine:
			converted.LineTo(at[0], at[1])
		case render.VerbCubic:
			converted.CubicTo(at[0], at[1], at[2], at[3], at[4], at[5])
		case render.VerbClose:
			converted.Close()
		}
	}
	return &converted
}

// pathOptions draws a path anti-aliased in clr.
func pathOptions(clr color.RGBA) *vector.DrawPathOptions {
	options := &vector.DrawPathOptions{AntiAlias: true}
	options.ColorScale.ScaleWithColor(clr)
	return options
}

// FillPath implements render.Canvas.
func (c *Canvas) FillPath(path *render.Path, clr color.RGBA) {
	vector.FillPath(c.dst, vectorPath(path), &vector.FillOptions{FillRule: vector.FillRuleNonZero}, pathOptions(clr))
}

// StrokePath implements render.Canvas.
func (c *Canvas) StrokePath(path *render.Path, stroke render.Stroke, clr color.RGBA) {
	options := &vector.StrokeOptions{
		Width:      stroke.Width,
		LineCap:    lineCaps[stroke.Cap],
		LineJoin:   lineJoins[stroke.Join],
		MiterLimit: stroke.MiterLimit,
	}
	vector.StrokePath(c.dst, vectorPath(path), options, pathOptions(clr))
}

// DrawImage implements render.Canvas.
func (c *Canvas) DrawImage(img render.Image, left, top, scale float64, filter render.Filter) {
	options := &ebiten.DrawImageOptions{Filter: filters[filter]}
	options.GeoM.Scale(scale, scale)
	options.GeoM.Translate(left, top)
	c.dst.DrawImage(pictureOf(img), options)
}

// DrawMask implements render.Canvas.
func (c *Canvas) DrawMask(mask render.Image, left, top float64, clr color.RGBA) {
	options := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	options.GeoM.Translate(left, top)
	options.ColorScale.ScaleWithColor(clr)
	c.dst.DrawImage(pictureOf(mask), options)
}

// DrawText implements render.Canvas. It draws content as it is: tabs are not
// expanded and no second pass adds weight.
func (c *Canvas) DrawText(face render.Face, content string, left, top float64, clr color.RGBA) {
	source, ok := face.Source.(*FontSource)
	if !ok {
		panic("ebitenrender: font source from another renderer")
	}
	options := &text.DrawOptions{}
	options.GeoM.Translate(left, top)
	options.ColorScale.ScaleWithColor(clr)
	text.Draw(c.dst, content, source.face(face.Size), options)
}
