package widget

import (
	"image"
	"image/color"
	"math"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/gascoigne-ai/uitoolkit/theming"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// Every built-in icon parses, draws something inside its own viewBox, and
// draws without a window at the scales the UI uses, hovered or not.
func TestBuiltInIconsParseAndStayInsideTheirViewBox(t *testing.T) {
	names := IconNames()
	if len(names) < 11 {
		t.Fatalf("only %d icons: %v", len(names), names)
	}
	for _, name := range names {
		ic, _ := lookupIcon(name)
		if ic.w <= 0 || ic.h <= 0 || len(ic.shapes) == 0 {
			t.Errorf("%s: empty icon", name)
		}
		for _, sh := range ic.shapes {
			for _, cmd := range sh.cmds {
				n := map[byte]int{'M': 1, 'L': 1, 'C': 3, 'Z': 0}[cmd.op]
				for i := 0; i < n; i++ {
					x, y := cmd.p[2*i], cmd.p[2*i+1]
					if x < -0.01 || y < -0.01 || x > ic.w+0.01 || y > ic.h+0.01 {
						t.Errorf("%s: point %.2f,%.2f outside the %gx%g viewBox", name, x, y, ic.w, ic.h)
					}
				}
			}
		}
		for _, scale := range []float64{1, 1.5, 2} {
			c := testCtx(t)
			if err := c.SetScale(scale); err != nil {
				t.Fatal(err)
			}
			c.In = core.Input{X: 5, Y: 5} // hovered
			drawn(t, c, Row(W(30), H(30), Icon(name)), 30, 30)
			c.In = core.Input{X: 500, Y: 500}
			drawn(t, c, Row(W(30), H(30), Icon(name, W(20), H(20))), 30, 30)
		}
	}
}

func TestIconsHaveTheirNaturalSizeAndScaleToFit(t *testing.T) {
	c := testCtx(t)
	nat := Icon("trash")
	Layout(c, Row(AlignItems(AlignCenter), nat), 100, 100)
	if nat.Rect().W != 14 || nat.Rect().H != 16 {
		t.Fatalf("natural size %dx%d", nat.Rect().W, nat.Rect().H)
	}
	if !hasIcon(TriangleIcon(true), "triangle-down") || !hasIcon(TriangleIcon(false), "triangle-right") {
		t.Fatal("the disclosure triangle picks its icon by state")
	}
}

func TestUnknownIconPanicsAndListsTheNames(t *testing.T) {
	defer func() {
		msg, _ := recover().(string)
		if !strings.Contains(msg, `no icon "nope"`) || !strings.Contains(msg, "trash") {
			t.Fatalf("panic %q", msg)
		}
	}()
	Icon("nope")
}

func TestRegisterIconsAddsAndReplacesAndReportsBrokenFiles(t *testing.T) {
	err := RegisterIcons(fstest.MapFS{
		"star.svg":   {Data: []byte(`<svg viewBox="0 0 10 10"><polygon points="5,0 6,4 10,4 7,6 8,10 5,7.5 2,10 3,6 0,4 4,4"/></svg>`)},
		"broken.svg": {Data: []byte("<svg viewBox=\"0 0 4 4\">\n<blink/></svg>")},
		"notes.txt":  {Data: []byte("ignored")},
	})
	if err == nil || !strings.Contains(err.Error(), "broken.svg:2") {
		t.Fatalf("err = %v", err)
	}
	if !HasIcon("star") || HasIcon("broken") {
		t.Fatal("registration")
	}
	t.Cleanup(func() { icons.Lock(); delete(icons.byName, "star"); icons.Unlock() })
}

func TestSVGSubsetErrorsNameFileAndLine(t *testing.T) {
	for name, tt := range map[string]struct{ src, want string }{
		"no root":       {`<path d="M0 0"/>`, "starts with <svg>"},
		"no viewBox":    {`<svg><path d="M0 0L1 1"/></svg>`, "viewBox"},
		"bad paint":     {`<svg viewBox="0 0 4 4"><path d="M0 0L1 1" stroke="red"/></svg>`, "none, currentColor or token"},
		"unknown tag":   {"<svg viewBox=\"0 0 4 4\">\n\n<text/></svg>", "svg:3: unknown element <text>"},
		"bad path":      {`<svg viewBox="0 0 4 4"><path d="L1 1"/></svg>`, "must start with M"},
		"short path":    {`<svg viewBox="0 0 4 4"><path d="M1"/></svg>`, "number is missing"},
		"draws nothing": {`<svg viewBox="0 0 4 4"></svg>`, "draws nothing"},
		"bad width":     {`<svg viewBox="0 0 4 4" stroke-width="x"><path d="M0 0L1 1"/></svg>`, "stroke-width"},
	} {
		_, err := parseSVG("t.svg", []byte(tt.src))
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: err = %v, want %q", name, err, tt.want)
		}
	}
}

func TestPathDataCommands(t *testing.T) {
	tests := []struct {
		name, d string
		last    [2]float64
		kinds   string
	}{
		{"absolute lines", "M1 2L3 4L5 6", [2]float64{5, 6}, "MLL"},
		{"relative", "m1 2l2 2h3v-1", [2]float64{6, 3}, "MLLL"},
		{"implicit lineto after M", "M0 0 4 4 8 0", [2]float64{8, 0}, "MLL"},
		{"packed numbers", "M1-2.5.5.5", [2]float64{0.5, 0.5}, "ML"},
		{"cubic", "M0 0C1 1 2 1 3 0", [2]float64{3, 0}, "MC"},
		{"smooth cubic", "M0 0C1 1 2 1 3 0S5 -1 6 0", [2]float64{6, 0}, "MCC"},
		{"quadratic", "M0 0Q1 2 2 0", [2]float64{2, 0}, "MC"},
		{"close returns to start", "M1 1L3 1L3 3Z", [2]float64{3, 3}, "MLLZ"},
		{"arc is cubics ending on the point", "M0 0A5 5 0 0 1 10 0", [2]float64{10, 0}, "MCC"},
		{"two subpaths", "M0 0L1 1M5 5L6 6", [2]float64{6, 6}, "MLML"},
	}
	for _, tt := range tests {
		cmds, err := parsePathData(tt.d)
		if err != nil {
			t.Errorf("%s: %v", tt.name, err)
			continue
		}
		var kinds string
		for _, c := range cmds {
			kinds += string(c.op)
		}
		if kinds != tt.kinds {
			t.Errorf("%s: commands %q, want %q", tt.name, kinds, tt.kinds)
		}
		last := cmds[len(cmds)-1]
		if last.op != 'Z' {
			var got [2]float64
			switch last.op {
			case 'C':
				got = [2]float64{last.p[4], last.p[5]}
			default:
				got = [2]float64{last.p[0], last.p[1]}
			}
			if got != tt.last {
				t.Errorf("%s: ends at %v, want %v", tt.name, got, tt.last)
			}
		}
	}
	// a half circle of radius 5 passes through (5, -5) or (5, 5) at its middle
	cmds, _ := parsePathData("M0 0A5 5 0 0 1 10 0")
	mid := cmds[1].p
	if d := mid[4] - 5; d < -0.01 || d > 0.01 {
		t.Errorf("arc middle x = %v", mid[4])
	}
}

// parsedIcon parses an icon of the tests or fails.
func parsedIcon(t *testing.T, src string) *svgIcon {
	t.Helper()
	ic, err := parseSVG("t.svg", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return ic
}

const strokedIcon = `<svg viewBox="0 0 10 10"><path d="M1 1L9 9" stroke="currentColor" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="bevel"/></svg>`

func TestRenderShapesStrokesAndFillsTheShapesAsPaths(t *testing.T) {
	ic := parsedIcon(t, `<svg viewBox="0 0 10 10"><path d="M1 1L9 9" stroke="currentColor" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="bevel"/><path d="M0 0L4 0L4 4Z" fill="currentColor"/></svg>`)
	canvas := rendertest.NewCanvas(100, 100)
	tint := color.RGBA{9, 8, 7, 255}
	renderShapes(canvas, ic, 10, 20, 3, 1.6, func(svgPaint) color.RGBA { return tint })
	clip := image.Rect(0, 0, 100, 100)
	var line render.Path
	line.MoveTo(13, 23)
	line.LineTo(37, 47)
	var triangle render.Path
	triangle.MoveTo(10, 20)
	triangle.LineTo(22, 20)
	triangle.LineTo(22, 32)
	triangle.Close()
	want := []rendertest.Op{
		rendertest.StrokePath{Clip: clip, Segments: line.Segments(), Color: tint,
			Stroke: render.Stroke{Width: float32(2 * 3 * 1.6), Cap: render.LineCapRound, Join: render.LineJoinBevel}},
		rendertest.FillPath{Clip: clip, Segments: triangle.Segments(), Color: tint},
	}
	if got := canvas.Ops(); !reflect.DeepEqual(got, want) {
		t.Errorf("ops = %#v\nwant %#v", got, want)
	}
}

func TestATokenIconIsDrawnAsPathsInTheColorOfEachShape(t *testing.T) {
	ic := parsedIcon(t, `<svg viewBox="0 0 4 4"><path d="M0 0L4 4" stroke="token(error)" fill="none"/><path d="M0 0L2 0L2 2Z" fill="token(accent)"/><path d="M0 4L4 4" stroke="currentColor" fill="none"/></svg>`)
	if ic.cacheable() {
		t.Fatal("an icon with a token paint is drawn directly")
	}
	c := testCtx(t)
	canvas := rendertest.NewCanvas(100, 100)
	paintIcon(c, canvas, ic, 0, 0, 1, core.Muted)
	var colors []color.RGBA
	for _, op := range canvas.Ops() {
		switch op := op.(type) {
		case rendertest.StrokePath:
			colors = append(colors, op.Color)
		case rendertest.FillPath:
			colors = append(colors, op.Color)
		default:
			t.Errorf("unexpected op %#v: a token icon makes no mask", op)
		}
	}
	want := []color.RGBA{c.Color("error"), c.Color("accent"), c.Color(core.Muted)}
	if !reflect.DeepEqual(colors, want) {
		t.Errorf("colors %v, want %v", colors, want)
	}
	if c.MaskCount() != 0 {
		t.Errorf("%d masks for a token icon", c.MaskCount())
	}
}

func TestThemeIconStrokeScalesTheLines(t *testing.T) {
	ic := parsedIcon(t, strokedIcon)
	widthAt := func(weight float64) float32 {
		c := testCtx(t)
		c.Th.IconStroke = weight
		paintIcon(c, rendertest.NewCanvas(100, 100), ic, 0, 0, 1, core.Muted)
		for _, op := range iconMask(c, ic, 1, weight).(*rendertest.Offscreen).Ops() {
			if stroke, ok := op.(rendertest.StrokePath); ok {
				return stroke.Stroke.Width
			}
		}
		t.Fatal("no stroke in the mask")
		return 0
	}
	thin, thick := widthAt(1), widthAt(1.6)
	if thin != 2 || thick != float32(2*1.6) {
		t.Errorf("stroke widths %v and %v, want 2 and %v", thin, thick, float32(2*1.6))
	}
}

func TestACurrentColorIconIsAWhiteMaskTintedWhenDrawn(t *testing.T) {
	ic := parsedIcon(t, strokedIcon)
	for _, scale := range []float64{1, 1.5, 2} {
		c := testCtx(t)
		if err := c.SetScale(scale); err != nil {
			t.Fatal(err)
		}
		canvas := rendertest.NewCanvas(100, 100)
		paintIcon(c, canvas, ic, 1.3, 2.4, 1, core.Info)
		mask := iconMask(c, ic, scale, 1)
		want := []rendertest.Op{rendertest.DrawMask{
			Clip: image.Rect(0, 0, 100, 100), Mask: mask,
			Left: math.Round(1.3*scale) - iconMaskMargin, Top: math.Round(2.4*scale) - iconMaskMargin,
			Color: c.Color(core.Info),
		}}
		if got := canvas.Ops(); !reflect.DeepEqual(got, want) {
			t.Errorf("scale %v: ops = %#v\nwant %#v", scale, got, want)
		}
		white := color.RGBA{255, 255, 255, 255}
		for _, op := range mask.(*rendertest.Offscreen).Ops() {
			if stroke, ok := op.(rendertest.StrokePath); !ok || stroke.Color != white {
				t.Errorf("scale %v: mask op %#v, want a white stroke", scale, op)
			}
		}
		if got := mask.Bounds(); got.Dx() != int(math.Ceil(10*scale))+2*iconMaskMargin {
			t.Errorf("scale %v: mask %v", scale, got)
		}
	}
}

func TestIconMasksAreKeptByTheCtxPerSizeAndWeight(t *testing.T) {
	c := testCtx(t)
	ic, _ := lookupIcon("trash")
	first := iconMask(c, ic, 2, 1)
	if iconMask(c, ic, 2, 1) != first {
		t.Fatal("the same icon, scale and weight must reuse its mask")
	}
	if iconMask(c, ic, 3, 1) == first || iconMask(c, ic, 2, 1.5) == first {
		t.Fatal("another scale or weight is another mask")
	}
	if b := first.Bounds(); b.Dx() != 28+2*iconMaskMargin || b.Dy() != 32+2*iconMaskMargin {
		t.Fatalf("mask %v for a 14x16 icon at 2x", b)
	}
	if !ic.cacheable() {
		t.Fatal("a currentColor icon is cacheable")
	}
	other := testCtx(t)
	if other.MaskCount() != 0 || iconMask(other, ic, 2, 1) == first {
		t.Fatal("masks belong to one Ctx, not to the package")
	}
}

func TestDrawingAnIconAgainReusesItsMaskUntilTheScaleOrThemeChanges(t *testing.T) {
	c := testCtx(t)
	for range 3 { // several frames, hovered and not: one mask per size
		drawn(t, c, Row(W(30), H(30), Icon("trash")), 30, 30)
	}
	if c.MaskCount() != 1 {
		t.Fatalf("%d masks after drawing one icon three times", c.MaskCount())
	}
	if err := c.SetScale(2); err != nil {
		t.Fatal(err)
	}
	if c.MaskCount() != 0 {
		t.Errorf("%d masks after SetScale", c.MaskCount())
	}
	drawn(t, c, Row(W(30), H(30), Icon("trash")), 30, 30)
	if err := c.SetTheme(theming.Light()); err != nil {
		t.Fatal(err)
	}
	if c.MaskCount() != 0 {
		t.Errorf("%d masks after SetTheme", c.MaskCount())
	}
}

// hasIcon reports whether the element is the icon called name, as a Dump
// shows it.
func hasIcon(element *El, name string) bool { return contains(props(element), "icon="+name) }

func TestIconsHaveTheirNaturalWidths(t *testing.T) {
	c := testCtx(t)
	icons := Row(CrossIcon(), TriangleIcon(false), DotsIcon())
	Layout(c, icons, 200, 30)
	want := []int{10, 10, 4}
	for i, k := range icons.Children() {
		if k.Rect().W != want[i] {
			t.Errorf("icon %d width %d, want %d", i, k.Rect().W, want[i])
		}
	}
}

// GearIcon is the built-in cogwheel at its natural 18 px, and it is among
// the registered icons.
func TestGearIconIsTheCogwheelAt18Px(t *testing.T) {
	if !HasIcon("gear") {
		t.Fatalf("no gear among %v", IconNames())
	}
	c := testCtx(t)
	gear := GearIcon()
	Layout(c, Row(AlignItems(AlignCenter), gear), 100, 100)
	if gear.Rect().W != 18 || gear.Rect().H != 18 {
		t.Fatalf("natural size %dx%d", gear.Rect().W, gear.Rect().H)
	}
	if !hasIcon(gear, "gear") {
		t.Fatal("GearIcon does not name the gear icon")
	}
}

func TestReplayIconExists(t *testing.T) {
	if !HasIcon("replay") {
		t.Fatalf("no replay icon; have %v", IconNames())
	}
}

// The checkbox's check and the code view's fold arrows are SVG icons, drawn
// as icon masks like every other icon, not paths built in the draw code.
func TestCheckboxCheckAndFoldArrowsAreIcons(t *testing.T) {
	c := testCtx(t)
	drawn(t, c, Row(W(30), H(30), Checkbox(true)), 30, 30)
	checks := c.MaskCount()
	if !HasIcon("check") || checks != 1 {
		t.Fatalf("check icon %v, %d masks after drawing a checked box, want 1", HasIcon("check"), checks)
	}
	var code CodeView
	code.SetLines([]Line{{Spans: []Span{{Text: "{"}}, Foldable: true}, {Spans: []Span{{Text: "}"}}}})
	drawn(t, c, Row(W(200), H(80), Code(&code)), 200, 80)
	folds := c.MaskCount()
	if folds != 2 {
		t.Fatalf("%d masks after drawing a fold arrow, want the check and the arrow", folds)
	}
}

// A code view span with an Icon is painted as that icon in the span's
// cell; its Text stays what search and copy see.
func TestCodeViewSpanWithAnIconDrawsTheIconNotItsText(t *testing.T) {
	var code CodeView
	code.SetLines([]Line{{Spans: []Span{{Text: "↑", Icon: "arrow-up", Style: core.Info}, {Text: " sent"}}}})
	c := testCtx(t)
	drawn(t, c, Row(W(200), H(40), Code(&code)), 200, 40)
	count := c.MaskCount()
	if count != 1 {
		t.Fatalf("%d icon masks, want the arrow's", count)
	}
	if got := code.Lines[0].Spans[0].Text; got != "↑" {
		t.Fatalf("text %q, want it kept for copy and search", got)
	}
}
