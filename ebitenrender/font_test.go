package ebitenrender_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/ebitenrender"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

func newFont(t *testing.T) *ebitenrender.FontSource {
	t.Helper()
	source, err := ebitenrender.NewDevice().NewFontSource(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func textV2Face(t *testing.T, size float64) *text.GoTextFace {
	t.Helper()
	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		t.Fatal(err)
	}
	return &text.GoTextFace{Source: source, Size: size}
}

func TestFontSourceMetricsAndMeasureMatchTextV2(t *testing.T) {
	source := newFont(t)
	for _, size := range []float64{11, 13, 19.5} {
		direct := textV2Face(t, size)
		metrics := direct.Metrics()
		want := render.Metrics{Ascent: metrics.HAscent, Descent: metrics.HDescent}
		face := render.Face{Source: source, Size: size}
		if got := face.Metrics(); got != want {
			t.Errorf("size %v: Metrics = %+v, want %+v", size, got, want)
		}
		for _, content := range []string{"Hello, gap", "two\nlines of text"} {
			width, _ := text.Measure(content, direct, 0)
			if got := face.Measure(content); got != width || got == 0 {
				t.Errorf("size %v: Measure(%q) = %v, want %v", size, content, got, width)
			}
		}
	}
}

func TestNewFontSourceRejectsDataThatIsNoFont(t *testing.T) {
	_, err := ebitenrender.NewDevice().NewFontSource([]byte("not a font"))
	if err == nil || !strings.HasPrefix(err.Error(), "load font: ") {
		t.Fatalf("NewFontSource = %v, want an error that starts with \"load font: \"", err)
	}
}

func TestFaceWithFontSourceIsComparable(t *testing.T) {
	first, second := newFont(t), newFont(t)
	cache := map[render.Face]int{{Source: first, Size: 13}: 1}
	if cache[render.Face{Source: first, Size: 13}] != 1 {
		t.Error("two faces of one font and size differ")
	}
	if _, ok := cache[render.Face{Source: second, Size: 13}]; ok {
		t.Error("faces of two loaded fonts are equal")
	}
}
