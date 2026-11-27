package view

import (
	"image"
	"reflect"
	"slices"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

// themedSizes are the default sizes with the named ones pinned.
func themedSizes(t *testing.T, pins map[string]int) core.Sizes {
	t.Helper()
	var metrics core.Metrics
	for name, pixels := range pins {
		pinnedMetrics, err := metrics.Pin(name, pixels)
		if err != nil {
			t.Fatal(err)
		}
		metrics = pinnedMetrics
	}
	base := core.DefaultSizes()
	return core.SizesFor(base.Line, base.SmallLine, metrics)
}

func TestSizedOptionsFollowCtxSizes(t *testing.T) {
	pins := map[string]int{"gutter": 20, "nearGap": 3, "rowH": 50, "iconBtnW": 70, "labelPad": 7}
	gutter, near := core.GutterOf, core.NearGapOf
	margin := func(element *El) []int { return []int{element.mL, element.mT, element.mR, element.mB} }
	padding := func(element *El) []int { return []int{element.padL, element.padT, element.padR, element.padB} }
	tests := []struct {
		name            string
		option          Option
		read            func(element *El) []int
		unthemed, after []int
	}{
		{"PadOf", PadOf(gutter), padding, []int{12, 12, 12, 12}, []int{20, 20, 20, 20}},
		{"PadXOf", PadXOf(gutter), padding, []int{12, 0, 12, 0}, []int{20, 0, 20, 0}},
		{"PadYOf", PadYOf(gutter), padding, []int{0, 12, 0, 12}, []int{0, 20, 0, 20}},
		{"PadLOf", PadLOf(gutter), padding, []int{12, 0, 0, 0}, []int{20, 0, 0, 0}},
		{"PadTOf", PadTOf(gutter), padding, []int{0, 12, 0, 0}, []int{0, 20, 0, 0}},
		{"PadROf", PadROf(gutter), padding, []int{0, 0, 12, 0}, []int{0, 0, 20, 0}},
		{"PadBOf", PadBOf(gutter), padding, []int{0, 0, 0, 12}, []int{0, 0, 0, 20}},
		{"GapOf", GapOf(near), func(element *El) []int { return []int{element.gap} }, []int{4}, []int{3}},
		{"MinWOf", MinWOf(core.IconBtnOf), func(element *El) []int { return []int{element.minW} }, []int{33}, []int{70}},
		{"FloatAtOf", FloatAtOf(gutter, near), func(element *El) []int { return []int{element.fl.x, element.fl.y} }, []int{12, 4}, []int{20, 3}},
		{"HOf", HOf(core.RowOf), func(element *El) []int { return []int{element.h} }, []int{29}, []int{50}},
		{"WOf", WOf(core.IconBtnOf), func(element *El) []int { return []int{element.w} }, []int{33}, []int{70}},
		{"SeparatorOf", SeparatorOf(core.LabelPadOf), func(element *El) []int { return []int{element.borderInset} }, []int{10}, []int{7}},
		{"MarginOf", MarginOf(gutter), margin, []int{12, 12, 12, 12}, []int{20, 20, 20, 20}},
		{"MarginXOf", MarginXOf(gutter), margin, []int{12, 0, 12, 0}, []int{20, 0, 20, 0}},
		{"MarginYOf", MarginYOf(gutter), margin, []int{0, 12, 0, 12}, []int{0, 20, 0, 20}},
		{"MarginLOf", MarginLOf(gutter), margin, []int{12, 0, 0, 0}, []int{20, 0, 0, 0}},
		{"MarginTOf", MarginTOf(gutter), margin, []int{0, 12, 0, 0}, []int{0, 20, 0, 0}},
		{"MarginROf", MarginROf(gutter), margin, []int{0, 0, 12, 0}, []int{0, 0, 20, 0}},
		{"MarginBOf", MarginBOf(gutter), margin, []int{0, 0, 0, 12}, []int{0, 0, 0, 20}},
		{"MaxHOf", MaxHOf(core.IconBtnOf), func(element *El) []int { return []int{element.MaxHeight()} }, []int{33}, []int{70}},
		{"DismissMarginOf", DismissMarginOf(core.NearGapOf), func(element *El) []int { return []int{element.dismissMargin} }, []int{4}, []int{3}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			element := box(test.option)
			if got := test.read(element); !slices.Equal(got, test.unthemed) {
				t.Errorf("as built = %v, want %v", got, test.unthemed)
			}
			ctx := testCtx(t)
			for _, step := range []struct {
				sizes core.Sizes
				want  []int
			}{{themedSizes(t, pins), test.after}, {core.DefaultSizes(), test.unthemed}} {
				ctx.Sizes = step.sizes
				Layout(ctx, Col(element), 300, 300)
				if got := test.read(element); !slices.Equal(got, step.want) {
					t.Errorf("laid out = %v, want %v", got, step.want)
				}
			}
			if test.name == "FloatAtOf" && (!element.fl.on || element.fl.mode != floatPoint) {
				t.Errorf("float = %+v, want a floating point", element.fl)
			}
		})
	}
}

// A literal given after a role replaces it, side by side: the sides it does
// not set keep following the theme.
func TestLaterLiteralReplacesRole(t *testing.T) {
	themed := themedSizes(t, map[string]int{"gutter": 20, "iconBtnW": 70})
	margin := func(element *El) []int { return []int{element.mL, element.mT, element.mR, element.mB} }
	maxHeight := func(element *El) []int { return []int{element.maxH} }
	tests := []struct {
		name    string
		options []Option
		read    func(element *El) []int
		want    []int
	}{
		{"Pad", []Option{PadOf(core.GutterOf), Pad(3)}, nil, []int{3, 3, 3, 3}},
		{"PadX", []Option{PadOf(core.GutterOf), PadX(3)}, nil, []int{3, 20, 3, 20}},
		{"PadY", []Option{PadOf(core.GutterOf), PadY(3)}, nil, []int{20, 3, 20, 3}},
		{"PadL", []Option{PadOf(core.GutterOf), PadL(3)}, nil, []int{3, 20, 20, 20}},
		{"PadT", []Option{PadOf(core.GutterOf), PadT(3)}, nil, []int{20, 3, 20, 20}},
		{"PadR", []Option{PadOf(core.GutterOf), PadR(3)}, nil, []int{20, 20, 3, 20}},
		{"PadB", []Option{PadOf(core.GutterOf), PadB(3)}, nil, []int{20, 20, 20, 3}},
		{"Gap", []Option{GapOf(core.GutterOf), Gap(3)}, func(element *El) []int { return []int{element.gap} }, []int{3}},
		{"MinW", []Option{MinWOf(core.IconBtnOf), MinW(3)}, func(element *El) []int { return []int{element.minW} }, []int{3}},
		{"BorderInset", []Option{SeparatorOf(core.GutterOf), BorderInset(3)}, func(element *El) []int { return []int{element.borderInset} }, []int{3}},
		{"Separator", []Option{SeparatorOf(core.GutterOf), Separator(3)}, func(element *El) []int { return []int{element.borderInset} }, []int{3}},
		{"FloatAt", []Option{FloatAtOf(core.GutterOf, core.GutterOf), FloatAt(3, 4)}, func(element *El) []int { return []int{element.fl.x, element.fl.y} }, []int{3, 4}},
		{"DismissMargin", []Option{DismissMarginOf(core.GutterOf), DismissMargin(1)}, func(element *El) []int { return []int{element.dismissMargin} }, []int{1}},
		{"Margin", []Option{MarginOf(core.GutterOf), Margin(3)}, margin, []int{3, 3, 3, 3}},
		{"MarginX", []Option{MarginOf(core.GutterOf), MarginX(3)}, margin, []int{3, 20, 3, 20}},
		{"MarginY", []Option{MarginOf(core.GutterOf), MarginY(3)}, margin, []int{20, 3, 20, 3}},
		{"MarginL", []Option{MarginOf(core.GutterOf), MarginL(3)}, margin, []int{3, 20, 20, 20}},
		{"MarginT", []Option{MarginOf(core.GutterOf), MarginT(3)}, margin, []int{20, 3, 20, 20}},
		{"MarginR", []Option{MarginOf(core.GutterOf), MarginR(3)}, margin, []int{20, 20, 3, 20}},
		{"MarginB", []Option{MarginOf(core.GutterOf), MarginB(3)}, margin, []int{20, 20, 20, 3}},
		{"MaxH", []Option{MaxHOf(core.GutterOf), MaxH(3)}, maxHeight, []int{3}},
		{"MaxH 0 lifts a role", []Option{MaxHOf(core.GutterOf), MaxH(0)}, maxHeight, []int{0}},
		{"Margin 0 lifts a role", []Option{MarginOf(core.GutterOf), Margin(0)}, margin, []int{0, 0, 0, 0}},
		{"role after a literal", []Option{Pad(3), PadXOf(core.GutterOf)}, nil, []int{20, 3, 20, 3}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items := make([]Item, len(test.options))
			for i, option := range test.options {
				items[i] = option
			}
			element := box(items...)
			ctx := testCtx(t)
			ctx.Sizes = themed
			Layout(ctx, Col(element), 300, 300)
			read := test.read
			if read == nil {
				read = func(element *El) []int { return []int{element.padL, element.padT, element.padR, element.padB} }
			}
			if got := read(element); !slices.Equal(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestShadowAndEdgeFollowCtxSizes(t *testing.T) {
	ctx := testCtx(t)
	ctx.Sizes = themedSizes(t, map[string]int{"shadowOffset": 9, "edgeW": 5})
	full := image.Rect(0, 0, 400, 400)
	canvas := rendertest.NewCanvas(400, 400)
	ops := drawOnCanvas(ctx, box(Shadow(), EdgeT(core.Rule), EdgeL(core.Rule), EdgeR(core.Rule), EdgeB(core.Rule)), 100, 40, canvas)
	fill := func(x0, y0, x1, y1 float32, token core.Token) rendertest.Op {
		return rendertest.FillRect{Clip: full, Box: render.Box{X0: x0, Y0: y0, X1: x1, Y1: y1}, Color: ctx.Color(token)}
	}
	want := []rendertest.Op{
		fill(9, 9, 109, 49, core.Shadow),
		fill(0, 0, 100, 5, core.Rule),
		fill(0, 0, 5, 40, core.Rule),
		fill(95, 0, 100, 40, core.Rule),
		fill(0, 35, 100, 40, core.Rule),
	}
	if !reflect.DeepEqual(ops, want) {
		t.Errorf("ops =\n%v, want\n%v", ops, want)
	}
}

func TestDragEdgeFollowsEdgeGrab(t *testing.T) {
	tests := []struct {
		name  string
		grab  int
		drags bool
	}{
		{"default edge, press 10 inside", 6, true},
		{"wide edge, the same press", 16, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := testCtx(t)
			ctx.Sizes = themedSizes(t, map[string]int{"edgeGrab": test.grab})
			state := &MoveState{}
			root := Stack(box(W(100), H(60), FloatCenter(), Moved(state), DragHandle()))
			Layout(ctx, root, 300, 300)
			rect := root.kids[0].rect
			ctx.In = core.Input{X: rect.X + 10, Y: rect.Y + 30, Click: true, Held: true, Clicks: 1}
			Update(ctx, root)
			if state.active != test.drags {
				t.Errorf("drag started = %v, want %v", state.active, test.drags)
			}
		})
	}
}

// A child just outside the clip still draws when a themed shadow reaches back
// into it.
func TestDrawReachFollowsTheShadow(t *testing.T) {
	if got, want := drawReach(1, 16)-drawReach(1, 4), 12; got != want {
		t.Errorf("reach grew by %d, want %d", got, want)
	}
	if got := drawReach(1.5, 4); got != 13 {
		t.Errorf("reach at the default shadow = %d, want 13", got)
	}
}
