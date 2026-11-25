// Package text measures text with go-text/typesetting, without a GPU or a
// window. Layout that wraps, ellipsizes or places a caret gets the same
// widths from it on every renderer.
//
// A FontSource is a render.FontSource: render.Face{Source, Size} stays the
// only face type. Measure follows the Ebitengine text/v2 face byte for byte
// (same shaper, same chunk cuts, same sizes), so the two renderers agree
// within 0.01 px; the conformance test pins that against it.
package text

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/go-text/typesetting/bidi"
	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/language"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"

	"github.com/gascoigne-ai/uitoolkit/render"
)

// maxSize keeps size*64 inside Int26_6 (int32) with room to sum advances.
const maxSize = 1 << 16

// lineBreaks are the code points that end a line, as in Ebitengine's text.
const lineBreaks = "\n\v\f\r\u0085  "

// FontSource is a font parsed from one TrueType or OpenType file (not a
// collection), usable at any size. It is a pointer type, as render.FontSource
// asks, and safe for concurrent use: shaping state is shared and guarded by a
// mutex.
type FontSource struct {
	face    *font.Face
	upem    float64
	ascent  float64
	descent float64

	// mutex guards everything below; the shaper, the segmenter and the
	// buffers keep state between calls.
	mutex     sync.Mutex
	shaper    shaping.HarfbuzzShaper
	segment   shaping.Segmenter
	paragraph bidi.Paragraph
	runes     []rune
	cuts      []int
}

var _ render.FontSource = (*FontSource)(nil)

// NewFontSource parses data, one font file. A collection, a truncated or a
// corrupt file is an error, never a panic: the bytes come from outside.
func NewFontSource(data []byte) (source *FontSource, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			source, err = nil, fmt.Errorf("parse font: %v", recovered)
		}
	}()
	loader, err := opentype.NewLoader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}
	parsed, err := font.NewFont(loader)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}
	face := font.NewFace(parsed)
	if face.Upem() == 0 {
		return nil, fmt.Errorf("parse font: units per em is 0")
	}
	source = &FontSource{face: face, upem: float64(face.Upem())}
	if extents, ok := face.FontHExtents(); ok {
		source.ascent = float64(extents.Ascender)
		source.descent = float64(-extents.Descender)
	}
	source.shaper.SetFontCacheSize(1)
	return source, nil
}

// Metrics implements render.FontSource. A size that is not a positive finite
// number has no height.
func (f *FontSource) Metrics(size float64) render.Metrics {
	if !validSize(size) {
		return render.Metrics{}
	}
	scale := size / f.upem
	return render.Metrics{Ascent: f.ascent * scale, Descent: f.descent * scale}
}

// Measure implements render.FontSource: the width of the widest line of text,
// lines ending at any of the Unicode line breaks (CRLF is one). Invalid UTF-8
// counts as U+FFFD. Text is measured in logical order, left to right: bidi
// runs are not reordered, which cannot change a width.
func (f *FontSource) Measure(text string, size float64) float64 {
	if text == "" || !validSize(size) {
		return 0
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	fixedSize := fixed.Int26_6(size * 64)
	var widest fixed.Int26_6
	for text != "" {
		line, rest := splitLine(text)
		widest = max(widest, f.lineAdvance(line, fixedSize))
		text = rest
	}
	return float64(widest) / 64
}

func validSize(size float64) bool {
	return size > 0 && size <= maxSize && !math.IsNaN(size)
}

// splitLine cuts the first line off text, without its break.
func splitLine(text string) (line, rest string) {
	index := strings.IndexAny(text, lineBreaks)
	if index < 0 {
		return text, ""
	}
	_, width := utf8.DecodeRuneInString(text[index:])
	end := index + width
	if text[index] == '\r' && end < len(text) && text[end] == '\n' {
		end++
	}
	return text[:index], text[end:]
}

// lineAdvance is the width of one line: the sum of its chunks, each shaped on
// its own. Cutting at sentence ends bounds what one shaping call sees and,
// like Ebitengine, loses kerning across a cut; matching that is the point.
// Callers hold f.mutex.
func (f *FontSource) lineAdvance(line string, size fixed.Int26_6) fixed.Int26_6 {
	f.cuts = appendCuts(f.cuts[:0], line, &f.paragraph)
	var total fixed.Int26_6
	start := 0
	for _, end := range f.cuts {
		if end > start {
			total += f.chunkAdvance(line[start:end], size)
		}
		start = end
	}
	return total
}

// chunkAdvance shapes one chunk and sums its glyph advances. Script and
// language are what a zero language tag gives: the segmenter picks the real
// script of each run, and "und" is a language the shaper does not know.
// Callers hold f.mutex.
func (f *FontSource) chunkAdvance(chunk string, size fixed.Int26_6) fixed.Int26_6 {
	f.runes = f.runes[:0]
	for _, char := range chunk {
		f.runes = append(f.runes, char)
	}
	input := shaping.Input{
		Text:      f.runes,
		RunStart:  0,
		RunEnd:    len(f.runes),
		Direction: di.DirectionLTR,
		Face:      f.face,
		Size:      size,
		Script:    language.Unknown,
		Language:  "und",
	}
	var total fixed.Int26_6
	for _, run := range f.segment.Split(input, singleFont{f.face}) {
		total += f.shaper.ShapeNoExtents(run).Advance
	}
	return total
}

// singleFont maps every rune to one face: no fallback fonts.
type singleFont struct{ face *font.Face }

func (s singleFont) ResolveFace(rune) *font.Face { return s.face }
