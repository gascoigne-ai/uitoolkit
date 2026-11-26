package sections

import (
	"slices"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/theming"
)

func TestThemeChoices(t *testing.T) {
	catalog := theming.NewCatalog(theming.Builtins())
	for _, choice := range catalog.All() {
		built := map[string]core.Theme{}
		for _, contrast := range theming.ContrastChoices() {
			t.Run(choice.ID+" "+contrast.ID, func(t *testing.T) {
				showcase, ctx, _ := newTestShowcase(t)
				showcase.SetThemeID(choice.ID)
				showcase.SetContrast(contrast.ID)
				frame(t, showcase, ctx)
				theme := showcase.Theme()
				want := catalog.Theme(theming.Prefs{Theme: choice.ID, Contrast: contrast.ID})
				for _, token := range []core.Token{core.Bg, core.Text, core.Border, core.Accent} {
					if theme.Color(token) != want.Color(token) {
						t.Errorf("%s = %v, want %v", token, theme.Color(token), want.Color(token))
					}
					if ctx.Color(token) != want.Color(token) {
						t.Errorf("the context draws %s as %v, want %v", token, ctx.Color(token), want.Color(token))
					}
				}
				built[contrast.ID] = theme
			})
		}
		for _, token := range []core.Token{core.Text, core.Border} {
			if built["normal"].Color(token) == built["high"].Color(token) {
				t.Errorf("%s: high contrast does not change %s", choice.ID, token)
			}
		}
	}
}

func TestScaleSteps(t *testing.T) {
	if want := []float64{1, 1.25, 1.5, 2}; !slices.Equal(ScaleSteps, want) {
		t.Fatalf("ScaleSteps = %v, want %v", ScaleSteps, want)
	}
	tests := []struct {
		name  string
		index int
		want  float64
	}{
		{"first", 0, 1},
		{"third", 2, 1.5},
		{"below the range", -3, 1},
		{"past the range", 99, 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			showcase, _, _ := newTestShowcase(t)
			showcase.SetScale(test.index)
			if got := showcase.ScaleFactor(); got != test.want {
				t.Errorf("ScaleFactor = %v, want %v", got, test.want)
			}
		})
	}
}
