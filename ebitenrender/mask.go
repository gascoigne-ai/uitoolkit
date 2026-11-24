// The corner masks and the shapes drawn from them are a copy of
// core/cornermask.go and of the helpers in core/ctx.go, so that a canvas draws
// the pixels core draws today. Task 003 removes core's copy.

package ebitenrender

import (
	"image/color"
	"math"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// maskLimit bounds the cached corner masks: a UI uses a handful of radii, so
// the limit is reached only by a runaway (a radius animated every frame).
const maskLimit = 64

// maskKind tells a filled disc from a ring.
type maskKind uint8

const (
	maskFill maskKind = iota
	maskRing
)

// cornerKey names one cached mask: a ring is also told by its line width.
type cornerKey struct {
	radius int
	kind   maskKind
	width  int
}

// cornerMasks caches white corner masks by radius, in device pixels. It is
// owned by one Device and used from the draw goroutine only. A mask is made in
// device pixels, so a change of scale or theme needs no reset. The makers are
// fields so a test can count builds without a GPU.
type cornerMasks struct {
	images map[cornerKey]*ebiten.Image
	make   func(radius int) *ebiten.Image
	// makeRing builds a ring mask of the given radius and line width.
	makeRing func(radius, width int) *ebiten.Image
}

// get returns the mask for radius, building it on a miss. When the cache is
// full it drops every mask first: a simple bound beats tracking use.
func (masks *cornerMasks) get(radius int) *ebiten.Image {
	return masks.lookup(cornerKey{radius: radius}, func() *ebiten.Image { return masks.make(radius) })
}

// getRing is get for the ring mask of a line width, in device pixels.
func (masks *cornerMasks) getRing(radius, width int) *ebiten.Image {
	key := cornerKey{radius: radius, kind: maskRing, width: width}
	return masks.lookup(key, func() *ebiten.Image { return masks.makeRing(radius, width) })
}

func (masks *cornerMasks) lookup(key cornerKey, build func() *ebiten.Image) *ebiten.Image {
	if img, ok := masks.images[key]; ok {
		return img
	}
	if masks.images == nil || len(masks.images) >= maskLimit {
		masks.images = make(map[cornerKey]*ebiten.Image)
	}
	img := build()
	masks.images[key] = img
	return img
}

// buildCornerMask draws a white square of side 2*radius with all corners
// rounded by radius, anti-aliased as a fill of that shape is, and one solid
// column to its right (the solid texel of sliceQuads). Quarters of it are the corners of a
// fill of any size.
func buildCornerMask(radius int) *ebiten.Image {
	size := float32(2 * radius)
	img := ebiten.NewImage(2*radius+1, 2*radius)
	verts, idx := cornersPath(render.Box{X1: size, Y1: size}, float32(radius), render.CornersAll).AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range verts {
		verts[i].ColorR, verts[i].ColorG, verts[i].ColorB, verts[i].ColorA = 1, 1, 1, 1
		verts[i].SrcX, verts[i].SrcY = 0.5, 0.5
	}
	img.DrawTriangles(verts, idx, gradientSrc, &ebiten.DrawTrianglesOptions{AntiAlias: true})
	vector.FillRect(img, size, 0, 1, size, color.White, false)
	return img
}

// buildRingMask draws a white ring of line width pixels along the edge of a
// 2*radius square with rounded corners, plus the solid column the straight
// edges sample. The stroke is the one the outline had before it came from a
// mask, so a corner is the same anti-aliased pixels.
func buildRingMask(radius, width int) *ebiten.Image {
	size := float32(2 * radius)
	half := float32(width) / 2
	img := ebiten.NewImage(2*radius+1, 2*radius)
	path := cornersPath(render.Box{X0: half, Y0: half, X1: size - half, Y1: size - half}, float32(radius)-half, render.CornersAll)
	var stroke vector.StrokeOptions
	stroke.Width = float32(width)
	var options vector.DrawPathOptions
	options.AntiAlias = true
	options.ColorScale.ScaleWithColor(color.White)
	vector.StrokePath(img, path, &stroke, &options)
	vector.FillRect(img, size, 0, 1, size, color.White, false)
	return img
}

// quad is one piece of a rounded shape: where it lands and the part of the
// mask it samples.
type quad struct{ dst, src render.Box }

// sliceQuads cuts the rectangle x0,y0 to x1,y1 into the pieces of a rounded
// fill: four squares of side radius at the corners (a quarter of the mask disc
// for a rounded corner, a solid texel for the others), four edge strips and
// the centre. radius is clamped to half the shorter side; empty pieces are
// left out.
func sliceQuads(x0, y0, x1, y1, radius float32, corners render.Corners) []quad {
	radius = min(radius, (x1-x0)/2, (y1-y0)/2)
	// The mask is the disc, 2*radius square, and a solid column right of it.
	solid := render.Box{X0: 2*radius + 0.5, Y0: 0.5, X1: 2*radius + 0.5, Y1: 0.5}
	piece := func(corner render.Corners, disc render.Box) render.Box {
		if corners&corner != 0 {
			return disc
		}
		return solid
	}
	xs := [4]float32{x0, x0 + radius, x1 - radius, x1}
	ys := [4]float32{y0, y0 + radius, y1 - radius, y1}
	cell := func(col, row int) render.Box {
		return render.Box{X0: xs[col], Y0: ys[row], X1: xs[col+1], Y1: ys[row+1]}
	}
	sources := [3][3]render.Box{
		{piece(render.CornerTL, render.Box{X1: radius, Y1: radius}), solid, piece(render.CornerTR, render.Box{X0: radius, X1: 2 * radius, Y1: radius})},
		{solid, solid, solid},
		{piece(render.CornerBL, render.Box{Y0: radius, X1: radius, Y1: 2 * radius}), solid, piece(render.CornerBR, render.Box{X0: radius, Y0: radius, X1: 2 * radius, Y1: 2 * radius})},
	}
	// Corners first (TL, TR, BR, BL), then the edges and the centre.
	order := [9][2]int{{0, 0}, {2, 0}, {2, 2}, {0, 2}, {1, 0}, {1, 2}, {0, 1}, {2, 1}, {1, 1}}
	quads := make([]quad, 0, len(order))
	for _, at := range order {
		dst := cell(at[0], at[1])
		if dst.X1 > dst.X0 && dst.Y1 > dst.Y0 {
			quads = append(quads, quad{dst, sources[at[1]][at[0]]})
		}
	}
	return quads
}

// ringQuads cuts the border of the rectangle x0,y0 to x1,y1, a line of the
// given width just inside it, into pieces: four corner squares sampling a
// quarter of the ring mask (radius is that of the mask, at most half the
// shorter side and at least the width), then the four straight edges, which
// sample the solid texel. What lies inside the line stays uncovered, apart
// from the corner squares, whose mask is clear there.
func ringQuads(x0, y0, x1, y1, radius, width float32) []quad {
	radius = min(radius, (x1-x0)/2, (y1-y0)/2)
	width = min(width, radius)
	solid := render.Box{X0: 2*radius + 0.5, Y0: 0.5, X1: 2*radius + 0.5, Y1: 0.5}
	quads := []quad{
		{render.Box{X0: x0, Y0: y0, X1: x0 + radius, Y1: y0 + radius}, render.Box{X1: radius, Y1: radius}},
		{render.Box{X0: x1 - radius, Y0: y0, X1: x1, Y1: y0 + radius}, render.Box{X0: radius, X1: 2 * radius, Y1: radius}},
		{render.Box{X0: x1 - radius, Y0: y1 - radius, X1: x1, Y1: y1}, render.Box{X0: radius, Y0: radius, X1: 2 * radius, Y1: 2 * radius}},
		{render.Box{X0: x0, Y0: y1 - radius, X1: x0 + radius, Y1: y1}, render.Box{Y0: radius, X1: radius, Y1: 2 * radius}},
	}
	edges := [4]render.Box{
		{X0: x0 + radius, Y0: y0, X1: x1 - radius, Y1: y0 + width},
		{X0: x0 + radius, Y0: y1 - width, X1: x1 - radius, Y1: y1},
		{X0: x0, Y0: y0 + radius, X1: x0 + width, Y1: y1 - radius},
		{X0: x1 - width, Y0: y0 + radius, X1: x1, Y1: y1 - radius},
	}
	for _, edge := range edges {
		if edge.X1 > edge.X0 && edge.Y1 > edge.Y0 {
			quads = append(quads, quad{edge, solid})
		}
	}
	return quads
}

// gradientAt is the color of a vertical gradient from top (at y0) to bottom
// (at y1) at height, as the color scales a vertex carries.
func gradientAt(top, bottom color.RGBA, y0, y1, height float32) (red, green, blue, alpha float32) {
	t := min(max((height-y0)/max(y1-y0, 1), 0), 1)
	mix := func(value, target uint8) float32 {
		return (float32(value) + (float32(target)-float32(value))*t) / 255
	}
	return mix(top.R, bottom.R), mix(top.G, bottom.G), mix(top.B, bottom.B), mix(top.A, bottom.A)
}

// gradientSrc is the white texel that triangles colored by their vertices
// sample. It is written once, here.
var gradientSrc = func() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.White)
	return img
}()

// rectPath is the square-cornered rectangle as a path.
func rectPath(box render.Box) *vector.Path {
	var path vector.Path
	path.MoveTo(box.X0, box.Y0)
	path.LineTo(box.X1, box.Y0)
	path.LineTo(box.X1, box.Y1)
	path.LineTo(box.X0, box.Y1)
	path.Close()
	return &path
}

// cornersPath is box with the corners in the set rounded by radius.
func cornersPath(box render.Box, radius float32, corners render.Corners) *vector.Path {
	radius = min(radius, (box.X1-box.X0)/2, (box.Y1-box.Y0)/2)
	of := func(corner render.Corners) float32 {
		if corners&corner != 0 {
			return radius
		}
		return 0
	}
	tl, tr, br, bl := of(render.CornerTL), of(render.CornerTR), of(render.CornerBR), of(render.CornerBL)
	var path vector.Path
	path.MoveTo(box.X0+tl, box.Y0)
	path.LineTo(box.X1-tr, box.Y0)
	if tr > 0 {
		path.Arc(box.X1-tr, box.Y0+tr, tr, -math.Pi/2, 0, vector.Clockwise)
	}
	path.LineTo(box.X1, box.Y1-br)
	if br > 0 {
		path.Arc(box.X1-br, box.Y1-br, br, 0, math.Pi/2, vector.Clockwise)
	}
	path.LineTo(box.X0+bl, box.Y1)
	if bl > 0 {
		path.Arc(box.X0+bl, box.Y1-bl, bl, math.Pi/2, math.Pi, vector.Clockwise)
	}
	path.LineTo(box.X0, box.Y0+tl)
	if tl > 0 {
		path.Arc(box.X0+tl, box.Y0+tl, tl, math.Pi, 3*math.Pi/2, vector.Clockwise)
	}
	path.Close()
	return &path
}
