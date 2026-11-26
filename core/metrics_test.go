package core

import (
	"reflect"
	"strings"
	"testing"

	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
)

func pinned(t *testing.T, metrics Metrics, name string, pixels int) Metrics {
	t.Helper()
	metrics, err := metrics.Pin(name, pixels)
	if err != nil {
		t.Fatal(err)
	}
	return metrics
}

func TestSizesFollowMetrics(t *testing.T) {
	tests := []struct {
		name    string
		metrics func(t *testing.T) Metrics
		change  func(want *Sizes)
	}{
		{"unit 3", func(*testing.T) Metrics { return Metrics{Unit: 3} }, func(want *Sizes) {
			want.Steps = Steps{Hair: 3, Near: 6, Tight: 9, Loose: 12, Label: 15, Gutter: 18, Dialog: 24}
			want.TabLead, want.EdgeW, want.CaretW, want.NearGap, want.ShadowOffset = 3, 3, 3, 6, 6
			want.TightGap, want.TabGap, want.RuleInset, want.EdgeGrab = 9, 9, 9, 9
			want.LooseGap, want.CellPad, want.LabelPad, want.TabSepInset, want.Gutter, want.DialogPad = 12, 12, 15, 15, 18, 24
			want.Row, want.Compact, want.Slim, want.Label, want.Note = 35, 29, 23, 26, 21
			want.Toolbar, want.TabBar, want.DialogToolbar, want.DialogTitle = 47, 53, 53, 59
			want.IconBtnW, want.WordBtnW = 41, 59
			want.DialogW, want.DialogButtonMinW = 570, 120
		}},
		{"step tight 4", func(*testing.T) Metrics { return Metrics{Steps: Steps{Tight: 4}} }, func(want *Sizes) {
			want.Steps.Tight = 8
			want.TightGap, want.TabGap, want.RuleInset, want.EdgeGrab = 8, 8, 8, 8
			want.TabBar, want.DialogToolbar = 45, 45
		}},
		{"leading 2", func(*testing.T) Metrics { return Metrics{Leading: 2} }, func(want *Sizes) {
			want.Leading = 2
			want.Row, want.Compact, want.Slim, want.Label, want.Note = 31, 27, 23, 25, 21
			want.Toolbar, want.TabBar, want.DialogToolbar, want.DialogTitle = 39, 43, 43, 47
			want.IconBtnW, want.WordBtnW = 35, 47
		}},
		{"pinned gutter", func(t *testing.T) Metrics { return pinned(t, Metrics{}, "gutter", 14) }, func(want *Sizes) {
			want.Gutter, want.WordBtnW = 14, 47
		}},
		{"pinned rowH moves the bars", func(t *testing.T) Metrics { return pinned(t, Metrics{}, "rowH", 32) }, func(want *Sizes) {
			want.Row = 32
			want.Toolbar, want.TabBar, want.DialogToolbar, want.DialogTitle = 40, 44, 44, 48
		}},
		{"pinned tabGap moves the tab bar", func(t *testing.T) Metrics { return pinned(t, Metrics{}, "tabGap", 10) }, func(want *Sizes) {
			want.TabGap, want.TabBar = 10, 49
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			want := DefaultSizes()
			test.change(&want)
			if got := SizesFor(baseLine, baseSmallLine, test.metrics(t)); got != want {
				t.Errorf("sizes =\n%+v, want\n%+v", got, want)
			}
		})
	}
}

func TestPinSurvivesKnobChange(t *testing.T) {
	metrics := pinned(t, Metrics{}, "gutter", 14)
	metrics.Unit = 3
	sizes := SizesFor(baseLine, baseSmallLine, metrics)
	if sizes.Gutter != 14 || sizes.LooseGap != 12 {
		t.Errorf("gutter %d looseGap %d, want the pin 14 and the new unit's 12", sizes.Gutter, sizes.LooseGap)
	}
}

// A copy of a Theme shares its Metrics, so a pin on the copy must not show
// in the original.
func TestPinLeavesTheOriginalAlone(t *testing.T) {
	first := pinned(t, Metrics{}, "gutter", 14)
	second := pinned(t, first, "gutter", 20)
	if got := SizesFor(baseLine, baseSmallLine, first).Gutter; got != 14 {
		t.Errorf("gutter of the original = %d, want 14", got)
	}
	if got := SizesFor(baseLine, baseSmallLine, second).Gutter; got != 20 {
		t.Errorf("gutter of the copy = %d, want 20", got)
	}
}

func TestPinBounds(t *testing.T) {
	for _, rule := range sizeRules {
		t.Run(rule.name, func(t *testing.T) {
			for _, pixels := range []int{rule.min - 1, rule.max + 1} {
				_, err := (Metrics{}).Pin(rule.name, pixels)
				if err == nil || !strings.Contains(err.Error(), `size "`+rule.name+`" must be within`) {
					t.Errorf("Pin(%q, %d) error = %v, want a bounds error", rule.name, pixels, err)
				}
			}
			for _, pixels := range []int{rule.min, rule.max} {
				if got := pinned(t, Metrics{}, rule.name, pixels); SizesFor(baseLine, baseSmallLine, got).Scope()[rule.name] != pixels {
					t.Errorf("Pin(%q, %d) does not resolve to it", rule.name, pixels)
				}
			}
		})
	}
	if _, err := (Metrics{}).Pin("gutterr", 1); err == nil || !strings.Contains(err.Error(), `unknown size "gutterr"`) {
		t.Errorf("error = %v, want unknown size", err)
	}
}

func TestSizesForClampsGoMetrics(t *testing.T) {
	tests := []struct {
		name      string
		got, same Metrics
	}{
		{"unit 0 is the default", Metrics{}, Metrics{Unit: 2}},
		{"unit too large", Metrics{Unit: 99}, Metrics{Unit: 4}},
		{"unit negative", Metrics{Unit: -1}, Metrics{Unit: 1}},
		{"leading negative", Metrics{Leading: -3}, Metrics{}},
		{"leading too large", Metrics{Leading: 99}, Metrics{Leading: 8}},
		{"step too large", Metrics{Steps: Steps{Gutter: 99}}, Metrics{Steps: Steps{Gutter: 16}}},
		{"step negative", Metrics{Steps: Steps{Gutter: -2}}, Metrics{Steps: Steps{Gutter: 1}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, want := SizesFor(baseLine, baseSmallLine, test.got), SizesFor(baseLine, baseSmallLine, test.same)
			if got != want {
				t.Errorf("sizes =\n%+v, want\n%+v", got, want)
			}
			if !reflect.DeepEqual(test.same, Metrics{}) && test.same.Unit != 2 && want == DefaultSizes() {
				t.Errorf("the knob changed nothing")
			}
		})
	}
}

func TestSizesScopeNamesComeFromTheRules(t *testing.T) {
	scope := SizesFor(baseLine, baseSmallLine, pinned(t, Metrics{}, "dialogMaxH", 600)).Scope()
	if len(scope) != len(sizeRules) {
		t.Errorf("%d names for %d rules", len(scope), len(sizeRules))
	}
	for _, rule := range sizeRules {
		if _, ok := scope[rule.name]; !ok {
			t.Errorf("no scope name %q", rule.name)
		}
	}
	if scope["dialogMaxH"] != 600 {
		t.Errorf("dialogMaxH = %v, want the pinned 600", scope["dialogMaxH"])
	}
}

func TestSetThemeAndSetFontsApplyMetrics(t *testing.T) {
	ctx, err := NewCtx(testTheme(), testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	themed := testTheme()
	themed.Metrics = pinned(t, Metrics{Leading: 2}, "gutter", 14)
	if err := ctx.SetTheme(themed); err != nil {
		t.Fatal(err)
	}
	if ctx.Sizes.Gutter != 14 || ctx.Sizes.Row != DefaultSizes().Row+2 {
		t.Errorf("after SetTheme: gutter %d row %d", ctx.Sizes.Gutter, ctx.Sizes.Row)
	}
	ctx.Sizes = Sizes{}
	if err := ctx.SetFonts(FontData{Regular: gomono.TTF, Bold: gomonobold.TTF}, FontData{}); err != nil {
		t.Fatal(err)
	}
	if ctx.Sizes.Gutter != 14 || ctx.Sizes.Leading != 2 {
		t.Errorf("after SetFonts: gutter %d leading %d", ctx.Sizes.Gutter, ctx.Sizes.Leading)
	}
	fresh, err := NewCtx(themed, testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	if fresh.Sizes.Gutter != 14 {
		t.Errorf("NewCtx: gutter %d, want 14", fresh.Sizes.Gutter)
	}
}
