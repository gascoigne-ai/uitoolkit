package ebitenrender_test

import (
	_ "embed"
	"math"
	"strings"
	"testing"

	ebitentext "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/gascoigne-ai/uitoolkit/render/text"
)

// The conformance test lives here and not in render/text because it needs
// Ebitengine's text/v2, and the root module must not depend on Ebitengine.
// It runs inside the run-loop TestMain of this package, so it needs a display
// (xvfb in CI).

// liberationSans is vendored (SIL OFL 1.1, testdata/LiberationSans-LICENSE.txt)
// because the Go fonts have no kerning pairs.
//
//go:embed testdata/LiberationSans-Regular.ttf
var liberationSans []byte

var testFonts = []struct {
	name string
	data []byte
}{
	{"goregular", goregular.TTF},
	{"gobold", gobold.TTF},
	{"gomono", gomono.TTF},
	{"liberationsans", liberationSans},
}

var testSizes = []float64{9, 12, 13.333, 16, 24.5, 48}

// corpus holds the inputs the conformance test measures. The cut cases land
// on every rule of the line chunker: sentence terminators, a lowercase
// continuation, bidi level changes and the 4 KiB fallback.
func corpus() []string {
	return []string{
		"", " ", "a", "Hello, world", "AV", "To", "Te Wa", "fi ffl", "   leading and trailing   ",
		"tab\there", "line one\nline two is longer", "a\r\nbb\rccc\vd\fe\u0085ff g hhh", "\n\n", "x\n",
		"Hi. Then? Very!", ". T", "? V", "Dr. smith went", "Wait... Over", "What?! Yes",
		strings.Repeat("word ", 1200),
		strings.Repeat("x", 5000),
		// " A", " T" and " Y" kern in LiberationSans, so a fallback cut after the
		// space (which drops that pair) changes the width.
		strings.Repeat("x", 4100) + " AT Y A",
		strings.Repeat("A ", 3000),
		strings.Repeat("Hi. Then? Very! A. T ", 300),
		"é café é", "Привет, мир. То AV",
		"日本語のテキスト。次の文", "abc שלום. עולם def",
		"مرحبا بالعالم! hello? العربية",
		"Aא. Bב? C", "\U0001F600 smile \U0001F44D", "\U0001F468‍\U0001F469‍\U0001F467 family",
		"\xff", "a\xffb. C", "\xe2\x82", "‍. T", "Hi.́ There",
	}
}

func TestMeasureAndMetricsMatchEbitenText(t *testing.T) {
	for _, font := range testFonts {
		t.Run(font.name, func(t *testing.T) {
			ours, err := text.NewFontSource(font.data)
			if err != nil {
				t.Fatalf("NewFontSource: %v", err)
			}
			theirs, err := ebitentext.NewGoTextFaceSource(strings.NewReader(string(font.data)))
			if err != nil {
				t.Fatalf("ebiten source: %v", err)
			}
			if font.name == "liberationsans" {
				// Guard: without kerning the whole test would pass on a
				// shaper that ignores it.
				pair, a, v := ours.Measure("AV", 48), ours.Measure("A", 48), ours.Measure("V", 48)
				ebitenFace := &ebitentext.GoTextFace{Source: theirs, Size: 48}
				ep, _ := ebitentext.Measure("AV", ebitenFace, 0)
				ea, _ := ebitentext.Measure("A", ebitenFace, 0)
				ev, _ := ebitentext.Measure("V", ebitenFace, 0)
				if !(ep < ea+ev) {
					t.Fatalf("ebiten has no kerning: AV=%v, A+V=%v", ep, ea+ev)
				}
				if !(pair < a+v) {
					t.Fatalf("no kerning: Measure(AV)=%v, Measure(A)+Measure(V)=%v", pair, a+v)
				}
			}
			for _, size := range testSizes {
				face := &ebitentext.GoTextFace{Source: theirs, Size: size}
				m := face.Metrics()
				got := ours.Metrics(size)
				if math.Abs(got.Ascent-m.HAscent) > 0.01 || math.Abs(got.Descent-m.HDescent) > 0.01 {
					t.Errorf("size %v: Metrics = %+v, ebiten ascent %v descent %v", size, got, m.HAscent, m.HDescent)
				}
				for _, content := range corpus() {
					want, _ := ebitentext.Measure(content, face, 0)
					have := ours.Measure(content, size)
					if math.Abs(have-want) > 0.01 {
						t.Errorf("size %v: Measure(%.40q) = %v, ebiten %v", size, content, have, want)
					}
				}
			}
		})
	}
}
