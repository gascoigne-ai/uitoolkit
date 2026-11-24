package render_test

import (
	"reflect"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
)

func TestFaceIsAComparableMapKey(t *testing.T) {
	font := &rendertest.FixedFont{AdvanceEm: 0.5}
	twin := &rendertest.FixedFont{AdvanceEm: 0.5}
	face := render.Face{Source: font, Size: 12}
	cases := []struct {
		name  string
		other render.Face
		equal bool
	}{
		{"same literal", render.Face{Source: font, Size: 12}, true},
		{"other size", render.Face{Source: font, Size: 13}, false},
		{"other font with the same fields", render.Face{Source: twin, Size: 12}, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := face == testCase.other; got != testCase.equal {
				t.Errorf("equal = %v, want %v", got, testCase.equal)
			}
			cache := map[render.Face]int{face: 1}
			if _, found := cache[testCase.other]; found != testCase.equal {
				t.Errorf("found in map = %v, want %v", found, testCase.equal)
			}
		})
	}
	if !reflect.TypeFor[render.Face]().Comparable() {
		t.Error("render.Face is not comparable")
	}
}

func TestFaceMeasuresWithItsSourceAtItsSize(t *testing.T) {
	font := &rendertest.FixedFont{AdvanceEm: 0.5, AscentEm: 0.75, DescentEm: 0.25}
	cases := []struct {
		name    string
		size    float64
		text    string
		width   float64
		metrics render.Metrics
	}{
		{"empty text", 12, "", 0, render.Metrics{Ascent: 9, Descent: 3}},
		{"one line", 12, "abcd", 24, render.Metrics{Ascent: 9, Descent: 3}},
		{"twice the size", 24, "abcd", 48, render.Metrics{Ascent: 18, Descent: 6}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			face := render.Face{Source: font, Size: testCase.size}
			if got := face.Measure(testCase.text); got != testCase.width {
				t.Errorf("Measure(%q) = %v, want %v", testCase.text, got, testCase.width)
			}
			if got := face.Metrics(); got != testCase.metrics {
				t.Errorf("Metrics() = %+v, want %+v", got, testCase.metrics)
			}
		})
	}
}
