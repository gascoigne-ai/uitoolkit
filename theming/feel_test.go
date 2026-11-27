package theming

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
)

const ms = time.Millisecond

func TestParseThemeFeelBlock(t *testing.T) {
	data, err := os.ReadFile("testdata/feel.json")
	if err != nil {
		t.Fatal(err)
	}
	dark := Dark()
	theme, err := ParseTheme(data, &dark)
	if err != nil {
		t.Fatal(err)
	}
	want := core.Feel{
		CaretHalfPeriod: 700 * ms,
		KeyRepeat:       core.Repeat{Delay: 600 * ms, Interval: 40 * ms},
		HoldRepeat:      core.Repeat{Delay: 450 * ms, Interval: 90 * ms},
		MultiClickTime:  650 * ms, MultiClickDistance: 8,
		TooltipDelay: 250 * ms, ToastTime: 8 * time.Second,
		PanSlop: 12, WheelStep: 80, WheelLines: 5, WheelColumns: 9,
	}
	if theme.Feel != want {
		t.Fatalf("Feel = %+v, want %+v", theme.Feel, want)
	}
	for _, block := range []string{`null`, `{}`} {
		theme, err := ParseTheme([]byte(`{"feel":`+block+`}`), nil)
		if err != nil || theme.Feel != (core.Feel{}) {
			t.Errorf("feel %s: %+v, %v, want nothing stated", block, theme.Feel, err)
		}
	}
}

func TestParseThemeFeelErrors(t *testing.T) {
	tests := []struct{ name, block, want string }{
		{"zero", `{"tooltipDelay":0}`, `feel: "tooltipDelay" must be within 100 and 1000`},
		{"negative", `{"wheelStep":-48}`, `feel: "wheelStep" must be within 8 and 400`},
		{"huge", `{"caretHalfPeriod":2000000000}`, `feel: "caretHalfPeriod" must be within 200 and 2000`},
		{"overflowing a duration", `{"toastTime":9223372036854775807}`, `feel: "toastTime" must be within 1000 and 60000`},
		{"beyond any number", `{"toastTime":1e30}`, "cannot unmarshal number"},
		{"null value", `{"holdInterval":null}`, `feel: "holdInterval" needs a number`},
		{"fraction", `{"panSlop":1.5}`, "cannot unmarshal number"},
		{"string", `{"wheelLines":"3"}`, "cannot unmarshal string"},
		{"unknown key", `{"blinkRate":500}`, `feel: unknown value "blinkRate"`},
		{"first bad key by name", `{"wheelStep":0,"holdDelay":0}`, `"holdDelay"`},
		{"not an object", `[500]`, "cannot unmarshal array"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dark := Dark()
			_, err := ParseTheme([]byte(`{"feel":`+test.block+`}`), &dark)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("err = %v, want it to contain %q", err, test.want)
			}
		})
	}
}

func TestChildThemeMergesFeel(t *testing.T) {
	parent, err := ParseTheme([]byte(`{"feel":{"wheelStep":100,"holdDelay":500}}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	child, err := ParseTheme([]byte(`{"feel":{"wheelStep":200,"panSlop":9}}`), &parent)
	if err != nil {
		t.Fatal(err)
	}
	want := core.Feel{WheelStep: 200, PanSlop: 9, HoldRepeat: core.Repeat{Delay: 500 * ms}}
	if child.Feel != want {
		t.Errorf("child Feel = %+v, want %+v", child.Feel, want)
	}
	if parent.Feel.WheelStep != 100 || parent.Feel.PanSlop != 0 {
		t.Errorf("the parent changed: %+v", parent.Feel)
	}
	plain, err := ParseTheme([]byte(`{"name":"plain"}`), &parent)
	if err != nil || plain.Feel != parent.Feel {
		t.Errorf("a child with no block: %+v, %v", plain.Feel, err)
	}
}

func TestBuiltinThemesStateNoFeel(t *testing.T) {
	for name, theme := range map[string]core.Theme{"dark": Dark(), "light": Light()} {
		if theme.Feel != (core.Feel{}) {
			t.Errorf("%s states %+v", name, theme.Feel)
		}
	}
}

// Every way a catalog builds a theme keeps its feel.
func TestCatalogKeepsFeel(t *testing.T) {
	dir := t.TempDir()
	file := `{"name":"child","extends":"slow","feel":{"panSlop":9}}`
	if err := os.WriteFile(filepath.Join(dir, "child.json"), []byte(file), 0o600); err != nil {
		t.Fatal(err)
	}
	slow := func() core.Theme {
		theme := Dark()
		theme.Feel.WheelStep = 100
		return theme
	}
	catalog := NewCatalog(append([]ThemeChoice{{"slow", "Slow", slow}}, Builtins()...))
	themes, problems := catalog.LoadUser(dir)
	if len(problems) != 0 {
		t.Fatal(problems)
	}
	catalog.SetUser(themes)
	want := core.Feel{WheelStep: 100, PanSlop: 9}
	prefs := Prefs{Theme: "child", Corners: "round", Contrast: "high", Flat: true,
		Fonts: FontSizes{UI: 20}, Colors: map[string]map[string]string{"child": {"selection": "#000000"}}}
	for name, theme := range map[string]core.Theme{
		"LoadUser": themes["child"].Build(),
		"For":      catalog.For("child", FontSizes{UI: 20}),
		"Theme":    catalog.Theme(prefs),
	} {
		if theme.Feel != want {
			t.Errorf("%s: Feel = %+v, want %+v", name, theme.Feel, want)
		}
	}
}
