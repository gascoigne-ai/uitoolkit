package view

import (
	"go/parser"
	"go/token"
	"image"
	"maps"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

// Only the files that declare the extension point (Behavior, OverDrawer) and
// the text behavior may name the GPU library, so that the runtime draws on a
// canvas of any renderer.
func TestOnlyCustomAndLeafImportEbiten(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") || name == "custom.go" || name == "leaf.go" {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(path, "hajimehoshi") || strings.HasSuffix(path, "/ebitenrender") {
				t.Errorf("%s imports %s; only custom.go and leaf.go may", name, path)
			}
		}
	}
}

// canvasCtx is a context whose drawing is easy to predict: flat fills, theme
// radii as written, a copy of the control table to write into, and a pointer
// that hovers nothing.
func canvasCtx(t *testing.T) *core.Ctx {
	t.Helper()
	ctx := testCtx(t)
	ctx.Th.NoGradients = true
	ctx.Th.RadiusScale = 1
	controls := maps.Clone(ctx.Th.Controls)
	if controls == nil {
		controls = map[string]core.ControlStyle{}
	}
	ctx.Th.Controls = controls
	ctx.In.X, ctx.In.Y = -1, -1
	return ctx
}

func drawOnCanvas(ctx *core.Ctx, root *El, width, height int, canvas *rendertest.Canvas) []rendertest.Op {
	Layout(ctx, root, width, height)
	Draw(ctx, root, canvas)
	return canvas.Ops()
}

func TestDrawOnACanvasRecordsTheElementsLook(t *testing.T) {
	full := image.Rect(0, 0, 400, 400)
	shadow := float32(core.ShadowOffset)
	tests := []struct {
		name    string
		scale   float64
		options []Option
		want    func(ctx *core.Ctx) []rendertest.Op
	}{
		{"Bg", 1, []Option{Bg(core.Panel)}, func(ctx *core.Ctx) []rendertest.Op {
			return []rendertest.Op{rendertest.FillRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 100, Y1: 40}, Color: ctx.Color(core.Panel)}}
		}},
		{"BorderT", 1, []Option{BorderT(core.Rule)}, func(ctx *core.Ctx) []rendertest.Op {
			return []rendertest.Op{rendertest.FillRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 100, Y1: 1}, Color: ctx.Color(core.Rule)}}
		}},
		{"BorderB", 1, []Option{BorderB(core.Rule)}, func(ctx *core.Ctx) []rendertest.Op {
			return []rendertest.Op{rendertest.FillRect{Clip: full, Box: render.Box{X0: 0, Y0: 39, X1: 100, Y1: 40}, Color: ctx.Color(core.Rule)}}
		}},
		{"BorderL and BorderR inset", 1, []Option{BorderL(core.Rule), BorderR(core.Muted), BorderInset(3)}, func(ctx *core.Ctx) []rendertest.Op {
			return []rendertest.Op{
				rendertest.FillRect{Clip: full, Box: render.Box{X0: 0, Y0: 3, X1: 1, Y1: 37}, Color: ctx.Color(core.Rule)},
				rendertest.FillRect{Clip: full, Box: render.Box{X0: 99, Y0: 3, X1: 100, Y1: 37}, Color: ctx.Color(core.Muted)},
			}
		}},
		{"Border outlines after the fill", 1, []Option{Bg(core.Panel), Border(core.BorderStrong)}, func(ctx *core.Ctx) []rendertest.Op {
			return []rendertest.Op{
				rendertest.FillRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 100, Y1: 40}, Color: ctx.Color(core.Panel)},
				rendertest.StrokeRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 100, Y1: 40}, Width: 1, Color: ctx.Color(core.BorderStrong)},
			}
		}},
		{"square control outline", 1, []Option{Control("square")}, func(ctx *core.Ctx) []rendertest.Op {
			return []rendertest.Op{rendertest.StrokeRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 100, Y1: 40}, Width: 1, Color: ctx.Color(core.BorderStrong)}}
		}},
		{"rounded control outline", 1, []Option{Control("round")}, func(ctx *core.Ctx) []rendertest.Op {
			return []rendertest.Op{rendertest.StrokeRoundRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 100, Y1: 40}, Radius: 6, Width: 1, Color: ctx.Color(core.BorderStrong)}}
		}},
		{"rounded control outline at scale 2", 2, []Option{Control("round")}, func(ctx *core.Ctx) []rendertest.Op {
			return []rendertest.Op{rendertest.StrokeRoundRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 200, Y1: 80}, Radius: 12, Width: 2, Color: ctx.Color(core.BorderStrong)}}
		}},
		{"Shadow", 1, []Option{Shadow()}, func(ctx *core.Ctx) []rendertest.Op {
			return []rendertest.Op{rendertest.FillRect{Clip: full, Box: render.Box{X0: shadow, Y0: shadow, X1: 100 + shadow, Y1: 40 + shadow}, Color: ctx.Color(core.Shadow)}}
		}},
		{"Bg and BorderT at scale 2", 2, []Option{Bg(core.Panel), BorderT(core.Rule)}, func(ctx *core.Ctx) []rendertest.Op {
			return []rendertest.Op{
				rendertest.FillRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 200, Y1: 80}, Color: ctx.Color(core.Panel)},
				rendertest.FillRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 200, Y1: 2}, Color: ctx.Color(core.Rule)},
			}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := canvasCtx(t)
			ctx.Th.Controls["square"] = core.ControlStyle{Border: core.BorderStrong}
			ctx.Th.Controls["round"] = core.ControlStyle{Radius: 6, Border: core.BorderStrong}
			if err := ctx.SetScale(tc.scale); err != nil {
				t.Fatal(err)
			}
			root := Col(append([]Item{W(100), H(40)}, optionItems(tc.options)...)...)
			got := drawOnCanvas(ctx, root, 100, 40, rendertest.NewCanvas(400, 400))
			if want := tc.want(ctx); !reflect.DeepEqual(got, want) {
				t.Errorf("ops = %#v\nwant %#v", got, want)
			}
		})
	}
}

func optionItems(opts []Option) []Item {
	items := make([]Item, len(opts))
	for i, opt := range opts {
		items[i] = opt
	}
	return items
}

func TestDrawPaintsTheBackgroundThenChildrenThenTheOutline(t *testing.T) {
	ctx := canvasCtx(t)
	full := image.Rect(0, 0, 200, 200)
	kid := Col(W(50), H(10), Bg(core.Alt))
	root := Col(W(100), H(40), Bg(core.Panel), Border(core.BorderStrong), kid)
	got := drawOnCanvas(ctx, root, 100, 40, rendertest.NewCanvas(200, 200))
	want := []rendertest.Op{
		rendertest.FillRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 100, Y1: 40}, Color: ctx.Color(core.Panel)},
		rendertest.FillRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 50, Y1: 10}, Color: ctx.Color(core.Alt)},
		rendertest.StrokeRect{Clip: full, Box: render.Box{X0: 0, Y0: 0, X1: 100, Y1: 40}, Width: 1, Color: ctx.Color(core.BorderStrong)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ops = %#v\nwant %#v", got, want)
	}
}

func TestClippingParentCullsAndClipsChildren(t *testing.T) {
	canvasRect := image.Rect(0, 0, 200, 200)
	parentRect := image.Rect(0, 0, 100, 100)
	tests := []struct {
		name   string
		clip   bool
		parent core.Rect
		kid    core.Rect
		want   []rendertest.Op
	}{
		{"child beyond the draw reach", true, core.Rect{W: 100, H: 100}, core.Rect{Y: 150, W: 10, H: 10}, nil},
		{"child straddling the edge", true, core.Rect{W: 100, H: 100}, core.Rect{X: 95, Y: 50, W: 20, H: 10},
			[]rendertest.Op{rendertest.FillRect{Clip: parentRect, Box: render.Box{X0: 95, Y0: 50, X1: 115, Y1: 60}}}},
		{"parent that does not clip", false, core.Rect{W: 100, H: 100}, core.Rect{X: 95, Y: 50, W: 20, H: 10},
			[]rendertest.Op{rendertest.FillRect{Clip: canvasRect, Box: render.Box{X0: 95, Y0: 50, X1: 115, Y1: 60}}}},
		{"parent fully off the canvas", true, core.Rect{X: 300, Y: 300, W: 100, H: 100}, core.Rect{W: 10, H: 10}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := canvasCtx(t)
			kid := Col(W(10), H(10), Bg(core.Alt))
			root := Col(W(100), H(100), kid)
			root.clip = tc.clip
			Layout(ctx, root, 100, 100)
			root.rect, kid.rect = tc.parent, tc.kid
			canvas := rendertest.NewCanvas(200, 200)
			Draw(ctx, root, canvas)
			want := tc.want
			for i, op := range want {
				fill := op.(rendertest.FillRect)
				fill.Color = ctx.Color(core.Alt)
				want[i] = fill
			}
			if got := canvas.Ops(); !reflect.DeepEqual(got, want) {
				t.Errorf("ops = %#v\nwant %#v", got, want)
			}
		})
	}
}

// spy counts the calls of the two ebiten-typed extension points and the bounds
// of the images they got.
type spy struct {
	Base
	draws, overs []image.Rectangle
}

func (s *spy) Draw(_ *El, _ *core.Ctx, dst *ebiten.Image) { s.draws = append(s.draws, dst.Bounds()) }
func (s *spy) DrawOver(_ *El, _ *core.Ctx, dst *ebiten.Image) {
	s.overs = append(s.overs, dst.Bounds())
}

func TestBehaviorsAreSkippedOnANonEbitenCanvas(t *testing.T) {
	ctx := canvasCtx(t)
	probe := &spy{}
	custom := Custom(probe, W(100), H(20), Bg(core.Panel))
	root := Col(W(100), H(60), custom, Text("hi"))
	canvas := rendertest.NewCanvas(200, 200)
	got := drawOnCanvas(ctx, root, 100, 60, canvas)
	if len(probe.draws)+len(probe.overs) != 0 {
		t.Errorf("behavior called on a canvas that is not ebiten's: Draw %d, DrawOver %d", len(probe.draws), len(probe.overs))
	}
	want := []rendertest.Op{rendertest.FillRect{Clip: image.Rect(0, 0, 200, 200), Box: render.Box{X0: 0, Y0: 0, X1: 100, Y1: 20}, Color: ctx.Color(core.Panel)}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ops = %#v\nwant %#v", got, want)
	}
}

func TestBehaviorsGetTheClippedEbitenImage(t *testing.T) {
	ctx := canvasCtx(t)
	probe := &spy{}
	custom := Custom(probe, W(20), H(20))
	root := Col(W(100), H(100), Clip(), custom)
	Layout(ctx, root, 100, 100)
	Draw(ctx, root, ebiten.NewImage(200, 200))
	want := []image.Rectangle{image.Rect(0, 0, 100, 100)}
	if !reflect.DeepEqual(probe.draws, want) || !reflect.DeepEqual(probe.overs, want) {
		t.Errorf("Draw got %v, DrawOver got %v, want one image of %v each", probe.draws, probe.overs, want)
	}
}

func TestDrawBelowThenDrawFromRecordTheOpsOfDraw(t *testing.T) {
	ctx := canvasCtx(t)
	root := Col(W(100), H(100), Bg(core.Panel), Clip(),
		Col(W(50), H(10), Bg(core.Alt)),
		Col(W(50), H(10), Bg(core.Muted)),
	)
	Layout(ctx, root, 100, 100)
	whole := rendertest.NewCanvas(200, 200)
	Draw(ctx, root, whole)
	split := rendertest.NewCanvas(200, 200)
	DrawBelow(ctx, root, 1, split)
	DrawFrom(ctx, root, 1, split)
	if len(whole.Ops()) != 3 {
		t.Fatalf("Draw recorded %d ops, want 3", len(whole.Ops()))
	}
	if !reflect.DeepEqual(split.Ops(), whole.Ops()) {
		t.Errorf("DrawBelow and DrawFrom recorded %#v\nwant %#v", split.Ops(), whole.Ops())
	}
}
