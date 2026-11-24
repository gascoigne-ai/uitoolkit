package rendertest_test

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

func TestFixedFontMeasuresTheLongestLineByRunes(t *testing.T) {
	font := &rendertest.FixedFont{AdvanceEm: 0.5}
	cases := []struct {
		name string
		text string
		want float64
	}{
		{"empty", "", 0},
		{"ascii", "abc", 15},
		{"runes, not bytes", "\u017c\u00f3\u0142\u0107", 20},
		{"longest line first", "abcd\nab", 20},
		{"longest line last", "ab\nabcd", 20},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := font.Measure(testCase.text, 10); got != testCase.want {
				t.Errorf("Measure(%q, 10) = %v, want %v", testCase.text, got, testCase.want)
			}
		})
	}
}

func TestFixedFontMetricsScaleWithTheSize(t *testing.T) {
	font := &rendertest.FixedFont{AscentEm: 0.75, DescentEm: 0.25}
	if got, want := font.Metrics(16), (render.Metrics{Ascent: 12, Descent: 4}); got != want {
		t.Errorf("Metrics(16) = %+v, want %+v", got, want)
	}
}
