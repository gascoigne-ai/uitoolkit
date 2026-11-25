package text_test

import (
	"go/build"
	"math"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"testing"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/text"
)

var _ render.FontSource = (*text.FontSource)(nil)

func TestMeasureIsTheWidestLine(t *testing.T) {
	source, err := text.NewFontSource(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	const size = 16
	short, long := source.Measure("ab", size), source.Measure("abcdefgh", size)
	cases := []struct {
		name, text string
		want       float64
	}{
		{"empty", "", 0},
		{"only breaks", "\n\r\n\v", 0},
		{"widest first", "abcdefgh\nab", long},
		{"widest last", "ab\nabcdefgh", long},
		{"crlf is one break", "ab\r\nab", short},
		{"cr then lf apart", "ab\r\n\nab", short},
		{"nel", "ab\u0085abcdefgh", long},
		{"line separator", "abcdefgh ab", long},
		{"paragraph separator", "ab abcdefgh", long},
		{"trailing break", "abcdefgh\n", long},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := source.Measure(testCase.text, size); math.Abs(got-testCase.want) > 1e-9 {
				t.Errorf("Measure(%q) = %v, want %v", testCase.text, got, testCase.want)
			}
		})
	}
	for _, size := range []float64{0, -3, math.NaN(), math.Inf(1)} {
		if got := source.Measure("abc", size); got != 0 {
			t.Errorf("Measure at size %v = %v, want 0", size, got)
		}
	}
}

func TestNewFontSourceRejectsBadData(t *testing.T) {
	random := make([]byte, 4096)
	rand.New(rand.NewSource(1)).Read(random)
	ttc := slices.Concat([]byte("ttcf\x00\x01\x00\x00\x00\x00\x00\x01\x00\x00\x00\x10"), make([]byte, 64))
	cases := []struct {
		name string
		data []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"truncated header", goregular.TTF[:8]},
		{"truncated tables", goregular.TTF[:len(goregular.TTF)/3]},
		{"random", random},
		{"collection", ttc},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			source, err := text.NewFontSource(testCase.data)
			if err == nil || source != nil {
				t.Errorf("NewFontSource = %v, %v; want nil and an error", source, err)
			}
		})
	}
}

func TestRenderTextImports(t *testing.T) {
	pkg, err := build.ImportDir(".", 0)
	if err != nil {
		t.Fatal(err)
	}
	allowed := []string{
		"github.com/gascoigne-ai/uitoolkit/render",
		"golang.org/x/image/math/fixed",
	}
	for _, path := range pkg.Imports {
		first, _, _ := strings.Cut(path, "/")
		if !strings.Contains(first, ".") {
			continue
		}
		if !slices.Contains(allowed, path) && !strings.HasPrefix(path, "github.com/go-text/typesetting/") {
			t.Errorf("render/text imports %s", path)
		}
	}
}

func TestFontSourceIsSafeForConcurrentUse(t *testing.T) {
	source, err := text.NewFontSource(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	inputs := []string{"Hello. World", "AV\nTo", "שלום abc", "x"}
	want := make([]float64, len(inputs))
	for i, in := range inputs {
		want[i] = source.Measure(in, 14)
	}
	var group sync.WaitGroup
	for worker := range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			for round := range 50 {
				i := (worker + round) % len(inputs)
				if got := source.Measure(inputs[i], 14); got != want[i] {
					t.Errorf("Measure(%q) = %v, want %v", inputs[i], got, want[i])
				}
				_ = source.Metrics(14)
			}
		}()
	}
	group.Wait()
}

func FuzzNewFontSource(f *testing.F) {
	f.Add(goregular.TTF[:200])
	f.Add([]byte("OTTO"))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		source, err := text.NewFontSource(data)
		if err != nil {
			return
		}
		source.Metrics(12)
		source.Measure("Hello. World\nAV", 12)
	})
}
