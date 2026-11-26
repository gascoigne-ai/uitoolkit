package widget

import (
	"strings"
	"testing"

	. "github.com/gascoigne-ai/uitoolkit/view"
)

const paragraphText = "Words that are long enough to need several lines when the room is narrow, and one when it is wide."

// A paragraph breaks at the width its parent gives it: the narrower, the
// taller, and never wider than the room.
func TestAParagraphWrapsToTheWidthItIsGiven(t *testing.T) {
	c := testCtx(t)
	heights := map[int]int{}
	for _, width := range []int{120, 240, 2000} {
		paragraph := Paragraph(paragraphText)
		Layout(c, Col(AlignItems(AlignStart), paragraph), width, 400)
		rect := paragraph.Rect()
		if rect.W > width {
			t.Fatalf("width %d: paragraph is %d wide", width, rect.W)
		}
		heights[width] = rect.H
	}
	if !(heights[120] > heights[240] && heights[240] > heights[2000]) {
		t.Fatalf("heights by width %v: a narrower room must give a taller paragraph", heights)
	}
}

// The height was measured at the offered width but the text wrapped at the
// narrower MaxW, so the extra lines drew over the element below.
func TestAParagraphWithMaxWIsAsTallAsItsLines(t *testing.T) {
	for _, tt := range []struct {
		name  string
		bound Option
	}{{"MaxW", MaxW(120)}, {"W", W(120)}} {
		t.Run(tt.name, func(t *testing.T) {
			c := testCtx(t)
			paragraph, next := Paragraph(paragraphText, tt.bound), Text("next")
			Layout(c, Col(AlignItems(AlignStart), paragraph, next), 2000, 400)
			lines := len(wrapParagraph(c, c.F.Small, paragraphText, 120))
			if lines < 2 {
				t.Fatalf("%d lines at 120: the text must wrap", lines)
			}
			if got, want := paragraph.Rect().H, lines*c.LineH(c.F.Small); got != want {
				t.Errorf("height %d, want %d (%d lines)", got, want, lines)
			}
			if bottom := paragraph.Rect().Y + paragraph.Rect().H; next.Rect().Y < bottom {
				t.Errorf("next starts at %d, above the paragraph's end at %d", next.Rect().Y, bottom)
			}
		})
	}
}

func TestAParagraphHoldsAllItsWords(t *testing.T) {
	c := testCtx(t)
	paragraph := Paragraph(paragraphText)
	Layout(c, Col(paragraph), 150, 400)
	got := strings.Join(wrapParagraph(c, c.F.Small, paragraphText, 150), " ")
	if got != paragraphText {
		t.Fatalf("lines joined %q, want %q", got, paragraphText)
	}
}

func TestAParagraphWithNoWordsHasNoHeight(t *testing.T) {
	c := testCtx(t)
	paragraph := Paragraph("")
	Layout(c, Col(AlignItems(AlignStart), paragraph), 150, 400)
	if paragraph.Rect().H != 0 {
		t.Fatalf("empty paragraph is %d high", paragraph.Rect().H)
	}
}

func TestADumpShowsTheTextOfAParagraph(t *testing.T) {
	if got := props(Paragraph(paragraphText)); !strings.Contains(got, `text="`+paragraphText+`"`) {
		t.Fatalf("dump %q does not show the text", got)
	}
}
