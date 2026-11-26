package theming

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
)

func resolved(metrics core.Metrics) core.Sizes {
	base := core.DefaultSizes()
	return core.SizesFor(base.Line, base.SmallLine, metrics)
}

// wantSizes are the sizes a block must resolve to, by scope name, and the
// leading and steps beside them.
func checkScope(t *testing.T, sizes core.Sizes, want map[string]int) {
	t.Helper()
	scope := sizes.Scope()
	for name, pixels := range want {
		if scope[name] != pixels {
			t.Errorf("%s = %v, want %d", name, scope[name], pixels)
		}
	}
}

func parseMetrics(t *testing.T, block string, base *core.Theme) (core.Theme, error) {
	t.Helper()
	return ParseTheme([]byte(`{"metrics":`+block+`}`), base)
}

func TestParseThemeMetrics(t *testing.T) {
	tests := []struct {
		name    string
		block   string
		want    map[string]int
		leading int
	}{
		{"unit", `{"unit":3}`, map[string]int{"gutter": 18, "rowH": 35, "dialogW": 570}, 0},
		{"leading", `{"leading":2}`, map[string]int{"rowH": 31, "checkboxSize": 15}, 2},
		{"step hair", `{"steps":{"hair":2}}`, map[string]int{"tabLead": 4, "edgeW": 4, "caretW": 4}, 0},
		{"step near", `{"steps":{"near":3}}`, map[string]int{"nearGap": 6, "shadowOffset": 6, "toolbarH": 41}, 0},
		{"step tight", `{"steps":{"tight":4}}`, map[string]int{"tightGap": 8, "edgeGrab": 8, "tabBarH": 45}, 0},
		{"step loose", `{"steps":{"loose":5}}`, map[string]int{"looseGap": 10, "cellPad": 10}, 0},
		{"step label", `{"steps":{"label":6}}`, map[string]int{"labelPad": 12, "tabSepInset": 12}, 0},
		{"step gutter", `{"steps":{"gutter":7}}`, map[string]int{"gutter": 14, "wordBtnW": 47}, 0},
		{"step dialog", `{"steps":{"dialog":10}}`, map[string]int{"dialogPad": 20, "dialogButtonMinW": 100}, 0},
		{"named sizes", `{"sizes":{"gutter":14,"rowH":32}}`, map[string]int{"gutter": 14, "rowH": 32, "toolbarH": 40}, 0},
		{"null is absent", `null`, map[string]int{"gutter": 12}, 0},
		{"empty block", `{}`, map[string]int{"gutter": 12, "rowH": 29}, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			theme, err := parseMetrics(t, test.block, nil)
			if err != nil {
				t.Fatal(err)
			}
			sizes := resolved(theme.Metrics)
			checkScope(t, sizes, test.want)
			if sizes.Leading != test.leading {
				t.Errorf("leading = %d, want %d", sizes.Leading, test.leading)
			}
		})
	}
}

func TestParseThemeMetricsRejectsBadBlocks(t *testing.T) {
	tests := []struct{ name, block, want string }{
		{"unit 0", `{"unit":0}`, "metrics: unit must be within 1 and 4"},
		{"unit 5", `{"unit":5}`, "metrics: unit must be within 1 and 4"},
		{"leading -1", `{"leading":-1}`, "metrics: leading must be within 0 and 8"},
		{"leading 9", `{"leading":9}`, "metrics: leading must be within 0 and 8"},
		{"step 0", `{"steps":{"tight":0}}`, `metrics: step "tight" must be within 1 and 16`},
		{"step 17", `{"steps":{"dialog":17}}`, `metrics: step "dialog" must be within 1 and 16`},
		{"pin too large", `{"sizes":{"gutter":65}}`, `metrics: size "gutter" must be within 0 and 64`},
		{"pin negative", `{"sizes":{"rowH":-1}}`, `metrics: size "rowH" must be within 8 and 160`},
		{"first bad size by name", `{"sizes":{"rowH":-1,"gutter":65}}`, `size "gutter"`},
		{"null size", `{"sizes":{"gutter":null}}`, `metrics: size "gutter" needs a number`},
		{"unknown size", `{"sizes":{"gutterr":1}}`, `metrics: unknown size "gutterr"`},
		{"unknown key in the block", `{"units":2}`, `unknown field "units"`},
		{"unknown step", `{"steps":{"huge":2}}`, `unknown field "huge"`},
		{"fraction", `{"unit":1.5}`, "cannot unmarshal number"},
		{"overflow", `{"sizes":{"gutter":1e30}}`, "cannot unmarshal number"},
		{"huge integer", `{"leading":99999999999999999999}`, "cannot unmarshal number"},
		{"text", `{"steps":{"hair":"2"}}`, "cannot unmarshal string"},
		{"not an object", `3`, "cannot unmarshal number"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseMetrics(t, test.block, nil)
			if err == nil || !strings.Contains(err.Error(), test.want) || !strings.HasPrefix(err.Error(), "theme file: ") {
				t.Errorf("error = %v, want a theme file error with %q", err, test.want)
			}
		})
	}
}

// An unknown key at the top level of a file with a metrics block stays an
// error.
func TestParseThemeMetricsKeepsTheTopLevelStrict(t *testing.T) {
	_, err := ParseTheme([]byte(`{"metric":{}}`), nil)
	if err == nil || !strings.Contains(err.Error(), `unknown field "metric"`) {
		t.Errorf("error = %v, want unknown field", err)
	}
}

func TestThemeMetricsAreInherited(t *testing.T) {
	start := Dark()
	base, err := parseMetrics(t, `{"unit":2,"leading":2,"steps":{"tight":4},"sizes":{"gutter":14,"rowH":40}}`, &start)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		block   string
		want    map[string]int
		leading int
	}{
		{"nothing stated", `{}`, map[string]int{"gutter": 14, "rowH": 40, "tightGap": 8, "nearGap": 4}, 2},
		{"a knob change keeps the pins", `{"unit":3}`, map[string]int{"gutter": 14, "rowH": 40, "tightGap": 12, "nearGap": 6}, 2},
		{"an explicit 0 leading wins", `{"leading":0}`, map[string]int{"gutter": 14}, 0},
		{"one step leaves the others", `{"steps":{"near":4}}`, map[string]int{"nearGap": 8, "tightGap": 8}, 2},
		{"a child pin wins, pins merge", `{"sizes":{"gutter":20,"cellPad":3}}`, map[string]int{"gutter": 20, "cellPad": 3, "rowH": 40}, 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			child, err := parseMetrics(t, test.block, &base)
			if err != nil {
				t.Fatal(err)
			}
			sizes := resolved(child.Metrics)
			checkScope(t, sizes, test.want)
			if sizes.Leading != test.leading {
				t.Errorf("leading = %d, want %d", sizes.Leading, test.leading)
			}
			checkScope(t, resolved(base.Metrics), map[string]int{"gutter": 14, "rowH": 40, "cellPad": 8})
		})
	}
}

func TestParseThemeMetricsFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/metrics.json")
	if err != nil {
		t.Fatal(err)
	}
	base := Dark()
	theme, err := ParseTheme(data, &base)
	if err != nil {
		t.Fatal(err)
	}
	sizes := resolved(theme.Metrics)
	checkScope(t, sizes, map[string]int{"gutter": 14, "rowH": 32, "tightGap": 12, "toolbarH": 44, "dialogPad": 24})
	if sizes.Leading != 2 || fmt.Sprint(sizes.Steps) != "{3 6 12 12 15 18 24}" {
		t.Errorf("leading %d steps %v", sizes.Leading, sizes.Steps)
	}
}

func TestBuiltinThemesResolveToDefaultSizes(t *testing.T) {
	for _, choice := range Builtins() {
		t.Run(choice.ID, func(t *testing.T) {
			if got := resolved(choice.Build().Metrics); got != core.DefaultSizes() {
				t.Errorf("sizes = %+v, want the defaults", got)
			}
		})
	}
}
