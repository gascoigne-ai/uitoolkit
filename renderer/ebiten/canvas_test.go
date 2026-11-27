package ebitenrender_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/gascoigne-ai/uitoolkit/renderer/ebiten"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// The side of the surface the pixel tests draw on.
const surface = 64

// Colors are alpha-premultiplied; the translucent ones show a conversion that
// an opaque color would hide.
var (
	red       = color.RGBA{255, 0, 0, 255}
	glass     = color.RGBA{120, 40, 10, 160}
	deepGlass = color.RGBA{10, 60, 90, 100}
)

func readImage(img *ebiten.Image) *image.RGBA {
	bounds := img.Bounds()
	pixels := make([]byte, 4*bounds.Dx()*bounds.Dy())
	img.ReadPixels(pixels)
	return &image.RGBA{Pix: pixels, Stride: 4 * bounds.Dx(), Rect: bounds}
}

// assertPainted fails when every pixel is transparent, so two empty surfaces
// never pass for equal drawings.
func assertPainted(t *testing.T, img *image.RGBA) {
	t.Helper()
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] != 0 {
			return
		}
	}
	t.Fatal("every pixel is transparent")
}

// assertSamePixels requires byte equality and reports how far apart two
// drawings are.
func assertSamePixels(t *testing.T, got, want *image.RGBA) {
	t.Helper()
	assertPainted(t, want)
	assertPainted(t, got)
	if got.Rect != want.Rect {
		t.Fatalf("bounds = %v, want %v", got.Rect, want.Rect)
	}
	differing, worst := 0, 0
	var first image.Point
	for i := 0; i < len(want.Pix); i += 4 {
		delta := 0
		for j := i; j < i+4; j++ {
			delta = max(delta, int(got.Pix[j])-int(want.Pix[j]), int(want.Pix[j])-int(got.Pix[j]))
		}
		if delta == 0 {
			continue
		}
		if differing == 0 {
			first = image.Pt(i/4%want.Rect.Dx(), i/4/want.Rect.Dx())
		}
		differing++
		worst = max(worst, delta)
	}
	if differing > 0 {
		t.Fatalf("%d pixels differ, by at most %d; the first at %v: got %v, want %v",
			differing, worst, first, got.RGBAAt(first.X, first.Y), want.RGBAAt(first.X, first.Y))
	}
}

// assertMatchesEbiten draws once with ebiten itself and once through a canvas
// and requires the same pixels. Both draw on one image, cleared in between:
// Ebitengine puts images on shared atlases, and how a draw rounds depends on
// where on them the image drawn on sits, so two images may get two results.
func assertMatchesEbiten(t *testing.T, direct func(dst *ebiten.Image), through func(canvas render.Canvas)) {
	t.Helper()
	dst := ebiten.NewImage(surface, surface)
	defer dst.Deallocate()
	direct(dst)
	want := readImage(dst)
	dst.Clear()
	through(ebitenrender.NewDevice().NewCanvas(dst))
	assertSamePixels(t, readImage(dst), want)
}

func vectorPath() *vector.Path {
	var path vector.Path
	path.MoveTo(4.5, 4)
	path.LineTo(50, 8.25)
	path.CubicTo(60, 30, 30, 62, 8, 48)
	path.Close()
	path.MoveTo(16, 16)
	path.LineTo(34, 18)
	path.LineTo(20, 36)
	return &path
}

func renderPath() *render.Path {
	var path render.Path
	path.MoveTo(4.5, 4)
	path.LineTo(50, 8.25)
	path.CubicTo(60, 30, 30, 62, 8, 48)
	path.Close()
	path.MoveTo(16, 16)
	path.LineTo(34, 18)
	path.LineTo(20, 36)
	return &path
}

func TestFillPathAndStrokePathMatchVector(t *testing.T) {
	t.Run("fill", func(t *testing.T) {
		assertMatchesEbiten(t, func(dst *ebiten.Image) {
			options := &vector.DrawPathOptions{AntiAlias: true}
			options.ColorScale.ScaleWithColor(glass)
			vector.FillPath(dst, vectorPath(), &vector.FillOptions{FillRule: vector.FillRuleNonZero}, options)
		}, func(canvas render.Canvas) { canvas.FillPath(renderPath(), glass) })
	})
	strokes := []struct {
		name    string
		stroke  render.Stroke
		options vector.StrokeOptions
	}{
		{"zero value bevels", render.Stroke{Width: 5}, vector.StrokeOptions{Width: 5, LineCap: vector.LineCapButt, LineJoin: vector.LineJoinMiter}},
		{"miter", render.Stroke{Width: 5, MiterLimit: 10}, vector.StrokeOptions{Width: 5, LineJoin: vector.LineJoinMiter, MiterLimit: 10}},
		{"round", render.Stroke{Width: 5, Cap: render.LineCapRound, Join: render.LineJoinRound}, vector.StrokeOptions{Width: 5, LineCap: vector.LineCapRound, LineJoin: vector.LineJoinRound}},
		{"square and bevel", render.Stroke{Width: 5, Cap: render.LineCapSquare, Join: render.LineJoinBevel}, vector.StrokeOptions{Width: 5, LineCap: vector.LineCapSquare, LineJoin: vector.LineJoinBevel}},
	}
	for _, testCase := range strokes {
		t.Run(testCase.name, func(t *testing.T) {
			assertMatchesEbiten(t, func(dst *ebiten.Image) {
				options := &vector.DrawPathOptions{AntiAlias: true}
				options.ColorScale.ScaleWithColor(glass)
				vector.StrokePath(dst, vectorPath(), &testCase.options, options)
			}, func(canvas render.Canvas) { canvas.StrokePath(renderPath(), testCase.stroke, glass) })
		})
	}
}

// checker is a small picture whose scaled pixels differ between the filters.
func checker() *image.RGBA {
	picture := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			clr := color.RGBA{200, 30, 0, 255}
			if (x+y)%2 == 1 {
				clr = color.RGBA{0, 40, 100, 128}
			}
			picture.SetRGBA(x, y, clr)
		}
	}
	return picture
}

// imageFilters pairs each render filter with ebiten's.
var imageFilters = []struct {
	name   string
	filter render.Filter
	direct ebiten.Filter
}{
	{"nearest", render.FilterNearest, ebiten.FilterNearest},
	{"linear", render.FilterLinear, ebiten.FilterLinear},
}

// assertDrawImageMatchesEbiten requires Canvas.DrawImage to draw picture as
// ebiten does. Both draws read the ebiten image behind picture: a scaled draw
// also rounds by where on an atlas its source sits, which on a texel edge picks
// the other texel, so a second image with the same pixels is not enough.
func assertDrawImageMatchesEbiten(t *testing.T, picture render.Image, filter render.Filter, direct ebiten.Filter) {
	t.Helper()
	assertMatchesEbiten(t, func(dst *ebiten.Image) {
		options := &ebiten.DrawImageOptions{Filter: direct}
		options.GeoM.Scale(5.5, 5.5)
		options.GeoM.Translate(3.25, 7)
		dst.DrawImage(ebitenrender.EbitenImageOf(picture), options)
	}, func(canvas render.Canvas) { canvas.DrawImage(picture, 3.25, 7, 5.5, filter) })
}

// The fillers take up atlas space, which moves the images made after them, both
// those drawn from and those drawn on. A comparison that depends on either place
// then fails on every run, not only when earlier tests and the garbage collector
// happen to leave such a place. The sides repeat, and the large ones come first,
// so that both filters meet places where they round differently.
func TestDrawImageComparisonIgnoresAtlasPlacement(t *testing.T) {
	device := ebitenrender.NewDevice()
	var fillers []*ebiten.Image
	defer func() {
		for _, filler := range fillers {
			filler.Deallocate()
		}
	}()
	for _, side := range []int{13, 31, 63, 0, 1, 2, 3, 5, 7, 13, 31, 63, 0, 1, 2, 3, 5, 7} {
		if side > 0 {
			source := ebiten.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, side, side)))
			// Filling an image makes it one that is drawn on.
			target := ebiten.NewImage(surface, surface)
			target.Fill(red)
			fillers = append(fillers, source, target)
		}
		picture := device.NewImage(checker())
		for _, testCase := range imageFilters {
			t.Run(fmt.Sprintf("%s after fillers of side %d", testCase.name, side), func(t *testing.T) {
				assertDrawImageMatchesEbiten(t, picture, testCase.filter, testCase.direct)
			})
		}
	}
}

func TestDrawImageScalesWithTheFilter(t *testing.T) {
	device := ebitenrender.NewDevice()
	drawn := map[string][]byte{}
	for _, testCase := range imageFilters {
		t.Run(testCase.name, func(t *testing.T) {
			picture := device.NewImage(checker())
			if got, want := picture.Bounds(), image.Rect(0, 0, 4, 4); got != want {
				t.Fatalf("Bounds = %v, want %v", got, want)
			}
			assertDrawImageMatchesEbiten(t, picture, testCase.filter, testCase.direct)
			off := device.NewOffscreen(surface, surface)
			defer off.Release()
			off.DrawImage(picture, 3.25, 7, 5.5, testCase.filter)
			drawn[testCase.name] = off.ReadPixels().Pix
		})
	}
	if bytes.Equal(drawn["nearest"], drawn["linear"]) {
		t.Error("both filters drew the same pixels")
	}
	t.Run("an offscreen is an image", func(t *testing.T) {
		source := device.NewOffscreen(4, 4)
		defer source.Release()
		source.FillRect(render.Box{X1: 4, Y1: 4}, red)
		target := device.NewOffscreen(8, 8)
		defer target.Release()
		target.DrawImage(source, 2, 2, 1, render.FilterNearest)
		if got := target.ReadPixels().RGBAAt(3, 3); got != red {
			t.Errorf("pixel = %v, want %v", got, red)
		}
	})
}

func TestDrawMaskTintsTheAlpha(t *testing.T) {
	alphas := image.NewRGBA(image.Rect(0, 0, 4, 1))
	for x, alpha := range []uint8{0, 64, 128, 255} {
		alphas.SetRGBA(x, 0, color.RGBA{alpha, alpha, alpha, alpha})
	}
	device := ebitenrender.NewDevice()
	mask := device.NewImage(alphas)
	assertMatchesEbiten(t, func(dst *ebiten.Image) {
		options := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
		options.GeoM.Translate(2, 3)
		options.ColorScale.ScaleWithColor(glass)
		dst.DrawImage(ebitenrender.EbitenImageOf(mask), options)
	}, func(canvas render.Canvas) { canvas.DrawMask(mask, 2, 3, glass) })
	off := device.NewOffscreen(8, 8)
	defer off.Release()
	off.DrawMask(mask, 2, 3, glass)
	pixels := off.ReadPixels()
	if got := pixels.RGBAAt(5, 3); got != glass {
		t.Errorf("under a full mask pixel = %v, want %v", got, glass)
	}
	if got := pixels.RGBAAt(2, 3); got != (color.RGBA{}) {
		t.Errorf("under an empty mask pixel = %v, want transparent", got)
	}
}

func TestDrawTextIsOnePassOfTextDraw(t *testing.T) {
	source := newFont(t)
	direct := textV2Face(t, 17)
	pass := func(dst *ebiten.Image) {
		options := &text.DrawOptions{}
		options.GeoM.Translate(3.5, 7)
		options.ColorScale.ScaleWithColor(glass)
		text.Draw(dst, "Hg wy.\tz", direct, options)
	}
	assertMatchesEbiten(t, pass, func(canvas render.Canvas) {
		canvas.DrawText(render.Face{Source: source, Size: 17}, "Hg wy.\tz", 3.5, 7, glass)
	})
}

func TestSubClipsWithoutTranslating(t *testing.T) {
	off := ebitenrender.NewDevice().NewOffscreen(20, 20)
	defer off.Release()
	clip := image.Rect(5, 6, 12, 14)
	sub := off.Sub(clip)
	if sub.Bounds() != clip {
		t.Fatalf("Sub bounds = %v, want %v", sub.Bounds(), clip)
	}
	nested := sub.Sub(image.Rect(8, 0, 30, 10))
	if got, want := nested.Bounds(), image.Rect(8, 6, 12, 10); got != want {
		t.Errorf("nested bounds = %v, want %v", got, want)
	}
	everything := render.Box{X1: 20, Y1: 20}
	for _, outside := range []image.Rectangle{image.Rect(30, 30, 40, 40), {}} {
		empty := off.Sub(outside)
		if !empty.Bounds().Empty() {
			t.Errorf("Sub(%v) bounds = %v, want empty", outside, empty.Bounds())
		}
		empty.FillRect(everything, red)
		empty.FillRoundRect(everything, 4, render.CornersAll, red, red)
		empty.FillPath(renderPath(), red)
		if !empty.Sub(clip).Bounds().Empty() {
			t.Errorf("a Sub of an empty canvas has bounds %v", empty.Sub(clip).Bounds())
		}
	}
	sub.FillRect(everything, red)
	pixels := off.ReadPixels()
	for y := range 20 {
		for x := range 20 {
			want := color.RGBA{}
			if image.Pt(x, y).In(clip) {
				want = red
			}
			if got := pixels.RGBAAt(x, y); got != want {
				t.Fatalf("pixel %d,%d = %v, want %v", x, y, got, want)
			}
		}
	}
}

func TestReadPixelsIsPremultipliedAndMatchesBounds(t *testing.T) {
	off := ebitenrender.NewDevice().NewOffscreen(7, 5)
	defer off.Release()
	if got, want := off.Bounds(), image.Rect(0, 0, 7, 5); got != want {
		t.Fatalf("Bounds = %v, want %v", got, want)
	}
	off.FillRect(render.Box{X1: 7, Y1: 5}, glass)
	pixels := off.ReadPixels()
	if pixels.Rect != off.Bounds() || pixels.Stride != 28 || len(pixels.Pix) != 140 {
		t.Fatalf("ReadPixels: rect %v, stride %d, %d bytes", pixels.Rect, pixels.Stride, len(pixels.Pix))
	}
	if got := pixels.RGBAAt(3, 2); got != glass {
		t.Errorf("pixel = %v, want %v", got, glass)
	}
}

func TestClearMakesEveryPixelTransparent(t *testing.T) {
	off := ebitenrender.NewDevice().NewOffscreen(6, 6)
	defer off.Release()
	off.FillRect(render.Box{X1: 6, Y1: 6}, red)
	assertPainted(t, off.ReadPixels())
	off.Clear()
	if got := off.ReadPixels().Pix; !bytes.Equal(got, make([]byte, 4*6*6)) {
		t.Errorf("pixels after Clear = %v", got)
	}
}

func assertPanics(t *testing.T, want string, run func()) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Errorf("panic = %v, want %q", got, want)
		}
	}()
	run()
}

func TestForeignImagePanics(t *testing.T) {
	off := ebitenrender.NewDevice().NewOffscreen(8, 8)
	defer off.Release()
	// A Sub canvas has Bounds, so it is a render.Image by shape only.
	part, ok := off.Sub(image.Rect(0, 0, 4, 4)).(render.Image)
	if !ok {
		t.Fatal("a canvas is expected to have the shape of an image")
	}
	for name, foreign := range map[string]render.Image{
		"another renderer": &rendertest.Image{Rect: image.Rect(0, 0, 2, 2)},
		"a sub canvas":     part,
	} {
		t.Run(name, func(t *testing.T) {
			const message = "render: image from another renderer"
			assertPanics(t, message, func() { off.DrawImage(foreign, 0, 0, 1, render.FilterNearest) })
			assertPanics(t, message, func() { off.DrawMask(foreign, 0, 0, red) })
		})
	}
}

func TestForeignFontPanics(t *testing.T) {
	off := ebitenrender.NewDevice().NewOffscreen(8, 8)
	defer off.Release()
	face := render.Face{Source: &rendertest.FixedFont{AdvanceEm: 1, AscentEm: 1}, Size: 8}
	assertPanics(t, "ebitenrender: font source from another renderer", func() {
		off.DrawText(face, "a", 0, 0, red)
	})
}
