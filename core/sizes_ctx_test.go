package core

import (
	"reflect"
	"testing"

	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
)

// A 24 px font has a line height near 27: the largest text size must not be
// cut down to a smaller line than its own face.
func TestSizesForTable(t *testing.T) {
	tests := []struct {
		name        string
		line, small int
		want        Sizes
	}{
		{"default", 17, 15, Sizes{Line: 17, SmallLine: 15, Row: 29, Compact: 25, Slim: 21, Label: 23, Note: 19, Toolbar: 37, TabBar: 41, DialogTitle: 45, DialogToolbar: 41, IconBtnW: 33, WordBtnW: 45, Checkbox: 15}},
		{"line 16 (the old default)", 16, 14, Sizes{Line: 16, SmallLine: 14, Row: 28, Compact: 24, Slim: 20, Label: 22, Note: 18, Toolbar: 36, TabBar: 40, DialogTitle: 44, DialogToolbar: 40, IconBtnW: 32, WordBtnW: 44, Checkbox: 14}},
		{"small text", 13, 11, Sizes{Line: 13, SmallLine: 11, Row: 25, Compact: 21, Slim: 17, Label: 19, Note: 15, Toolbar: 33, TabBar: 37, DialogTitle: 41, DialogToolbar: 37, IconBtnW: 29, WordBtnW: 41, Checkbox: 11}},
		{"line 15", 15, 13, Sizes{Line: 15, SmallLine: 13, Row: 27, Compact: 23, Slim: 19, Label: 21, Note: 17, Toolbar: 35, TabBar: 39, DialogTitle: 43, DialogToolbar: 39, IconBtnW: 31, WordBtnW: 43, Checkbox: 13}},
		{"line 19", 19, 15, Sizes{Line: 19, SmallLine: 15, Row: 31, Compact: 27, Slim: 23, Label: 25, Note: 19, Toolbar: 39, TabBar: 43, DialogTitle: 47, DialogToolbar: 43, IconBtnW: 35, WordBtnW: 47, Checkbox: 15}},
		{"a large line is not clamped", 27, 25, Sizes{Line: 27, SmallLine: 25, Row: 39, Compact: 35, Slim: 31, Label: 33, Note: 29, Toolbar: 47, TabBar: 51, DialogTitle: 55, DialogToolbar: 51, IconBtnW: 43, WordBtnW: 55, Checkbox: 25}},
		{"line clamped high", 40, 20, Sizes{Line: 28, SmallLine: 20, Row: 40, Compact: 36, Slim: 32, Label: 34, Note: 24, Toolbar: 48, TabBar: 52, DialogTitle: 56, DialogToolbar: 52, IconBtnW: 44, WordBtnW: 56, Checkbox: 20}},
		{"line clamped low, small clamped to line", 5, 4, Sizes{Line: 13, SmallLine: 4, Row: 25, Compact: 21, Slim: 17, Label: 19, Note: 8, Toolbar: 33, TabBar: 37, DialogTitle: 41, DialogToolbar: 37, IconBtnW: 29, WordBtnW: 41, Checkbox: 4}},
		{"small clamped to line", 14, 30, Sizes{Line: 14, SmallLine: 14, Row: 26, Compact: 22, Slim: 18, Label: 20, Note: 18, Toolbar: 34, TabBar: 38, DialogTitle: 42, DialogToolbar: 38, IconBtnW: 30, WordBtnW: 42, Checkbox: 14}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := SizesFor(test.line, test.small, Metrics{})
			if heights := fontSizes(got); heights != test.want {
				t.Fatalf("SizesFor(%d, %d) = %+v, want %+v", test.line, test.small, got, test.want)
			}
			checkSizeInvariants(t, got)
		})
	}
}

// fontSizes keeps the fields that follow the line heights; the rest are
// pinned by TestDefaultSizesKeepTodaysValues.
func fontSizes(sizes Sizes) Sizes {
	return Sizes{
		Line: sizes.Line, SmallLine: sizes.SmallLine,
		Row: sizes.Row, Compact: sizes.Compact, Slim: sizes.Slim, Label: sizes.Label, Note: sizes.Note,
		Toolbar: sizes.Toolbar, TabBar: sizes.TabBar, DialogTitle: sizes.DialogTitle, DialogToolbar: sizes.DialogToolbar,
		IconBtnW: sizes.IconBtnW, WordBtnW: sizes.WordBtnW, Checkbox: sizes.Checkbox,
	}
}

func checkSizeInvariants(t *testing.T, sizes Sizes) {
	t.Helper()
	if sizes.Line < MinLine || sizes.Line > MaxLine {
		t.Errorf("Line %d outside %d..%d", sizes.Line, MinLine, MaxLine)
	}
	for name, height := range map[string]int{
		"Row": sizes.Row, "Compact": sizes.Compact, "Slim": sizes.Slim, "Label": sizes.Label,
		"Toolbar": sizes.Toolbar, "TabBar": sizes.TabBar, "DialogTitle": sizes.DialogTitle, "DialogToolbar": sizes.DialogToolbar,
	} {
		if height < sizes.Line || (height-sizes.Line)%2 != 0 {
			t.Errorf("%s = %d for line %d: want >= line and an even padding", name, height, sizes.Line)
		}
	}
	if sizes.Note < sizes.SmallLine || (sizes.Note-sizes.SmallLine)%2 != 0 {
		t.Errorf("Note = %d for small line %d", sizes.Note, sizes.SmallLine)
	}
	if !(sizes.Row > sizes.Compact && sizes.Compact > sizes.Slim) {
		t.Errorf("want Row > Compact > Slim, got %d %d %d", sizes.Row, sizes.Compact, sizes.Slim)
	}
}

func TestSizesScopeNames(t *testing.T) {
	want := map[string]any{
		"rowH": 29, "compactH": 25, "slimH": 21, "labelH": 23, "noteH": 19, "toolbarH": 37, "tabBarH": 41,
		"dialogToolbarH": 41, "dialogTitleH": 45, "iconBtnW": 33, "wordBtnW": 45, "checkboxSize": 15,
		"tabLead": 2, "edgeW": 2, "nearGap": 4, "shadowOffset": 4, "tightGap": 6, "tabGap": 6,
		"ruleInset": 6, "edgeGrab": 6, "looseGap": 8, "cellPad": 8, "labelPad": 10, "tabSepInset": 10,
		"gutter": 12, "dialogPad": 16,
		"caretW": 2, "dialogW": 380, "dialogButtonMinW": 80, "dialogMargin": 40, "dialogMaxH": 560,
	}
	if got := DefaultSizes().Scope(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Scope() = %v, want %v", got, want)
	}

	big := SizesFor(20, 16, Metrics{}).Scope()
	if big["rowH"] != 32 || big["gutter"] != 12 || big["tightGap"] != 6 {
		t.Fatalf("a larger line moves heights only: rowH %v gutter %v tightGap %v", big["rowH"], big["gutter"], big["tightGap"])
	}
}

func TestCtxStartsWithSizesOfItsFaces(t *testing.T) {
	ctx, err := NewCtx(testTheme(), testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Sizes != DefaultSizes() {
		t.Fatalf("Sizes = %+v, want the default %+v", ctx.Sizes, DefaultSizes())
	}
}

func TestCtxSetThemeUpdatesSizes(t *testing.T) {
	ctx, _ := NewCtx(testTheme(), testDevice{})
	large := testTheme()
	large.FontSize, large.SmallSize = 20, 16
	if err := ctx.SetTheme(large); err != nil {
		t.Fatal(err)
	}
	if ctx.Sizes.Line <= DefaultSizes().Line || ctx.Sizes.Row <= DefaultSizes().Row || ctx.Sizes.SmallLine <= DefaultSizes().SmallLine {
		t.Fatalf("Sizes did not grow with the theme: %+v", ctx.Sizes)
	}
	want := SizesFor(ctx.LineH(ctx.F.UI), ctx.LineH(ctx.F.Small), Metrics{})
	if ctx.Sizes != want {
		t.Fatalf("Sizes = %+v, want %+v", ctx.Sizes, want)
	}
}

// The embedded Go fonts share their metrics, so no test font has another line
// height; the theme sizes are changed behind SetFonts' back instead, and the
// call must measure the faces again.
func TestCtxSetFontsUpdatesSizes(t *testing.T) {
	ctx, _ := NewCtx(testTheme(), testDevice{})
	ctx.Th.FontSize, ctx.Th.SmallSize = 20, 16
	if err := ctx.SetFonts(FontData{Regular: gomono.TTF, Bold: gomonobold.TTF}, FontData{}); err != nil {
		t.Fatal(err)
	}
	want := SizesFor(ctx.LineH(ctx.F.UI), ctx.LineH(ctx.F.Small), Metrics{})
	if ctx.Sizes != want || ctx.Sizes == DefaultSizes() {
		t.Fatalf("Sizes = %+v, want %+v from the new faces", ctx.Sizes, want)
	}
	before := ctx.Sizes
	if err := ctx.SetFonts(FontData{Regular: []byte("nope")}, FontData{}); err == nil {
		t.Fatal("want an error for unreadable fonts")
	}
	if ctx.Sizes != before {
		t.Fatal("unreadable fonts changed Sizes")
	}
}

func TestCtxSetScaleKeepsSizes(t *testing.T) {
	for _, scale := range []float64{1.1, 1.5, 2} {
		ctx, _ := NewCtx(testTheme(), testDevice{})
		before := ctx.Sizes
		if err := ctx.SetScale(scale); err != nil {
			t.Fatal(err)
		}
		if ctx.Sizes != before {
			t.Errorf("scale %v: Sizes = %+v, want %+v", scale, ctx.Sizes, before)
		}
	}
}
