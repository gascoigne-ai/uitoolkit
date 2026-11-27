package widget

import (
	"strings"
	"testing"

	. "github.com/gascoigne-ai/uitoolkit/view"
)

// unevenWords is a long text of words of many lengths, so that a few pixels
// less width break its lines in other places.
func unevenWords() string {
	words := make([]string, 600)
	for i := range words {
		words[i] = strings.Repeat("a", 1+(i*7)%11)
	}
	return strings.Join(words, " ")
}

// Regression: a Scroll measured its content at its full width and then laid it
// out beside the scrollbar, where a wrapping Paragraph is taller. The scroll
// range ended too early and the last lines could not be reached.
func TestAScrollReachesTheEndOfTextThatWrapsBesideItsScrollbar(t *testing.T) {
	tests := []struct {
		name  string
		width int
		pad   int
	}{
		{name: "narrow", width: 260},
		{name: "the width of a message dialog's body", width: 348},
		{name: "padded by less than the scrollbar", width: 352, pad: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := testCtx(t)
			state := &ScrollState{}
			build := func() *El {
				return Col(Scroll(state, Pad(test.pad), Paragraph(unevenWords()), Text("the end")))
			}
			root := build()
			Layout(ctx, root, test.width, 200)
			state.Y = 1 << 20
			Update(ctx, root)
			root = build()
			Layout(ctx, root, test.width, 200)
			scroll := root.Children()[0]
			last := scroll.Children()[1].Rect()
			bottom := scroll.Rect().Y + scroll.Rect().H - test.pad
			if got := last.Y + last.H; got != bottom {
				t.Errorf("scrolled to the end, the last line ends at %d, want %d", got, bottom)
			}
		})
	}
}

func TestATallMessageDialogReachesTheEndOfAWrappingParagraph(t *testing.T) {
	ctx := testCtx(t)
	state := &DialogState{Title: "T", Buttons: []DialogButton{{Label: "OK"}}}
	build := func() Node { return MessageDialog(state, Paragraph(unevenWords()), Text("the end")) }
	root, _ := inWindow(ctx, build(), 800, 600)
	state.Scroll.Y = 1 << 20
	Update(ctx, root)
	_, panel := inWindow(ctx, build(), 800, 600)
	scroll := panel.Children()[1].Children()[0]
	last := scroll.Children()[1].Rect()
	if !inside(last, scroll.Rect()) {
		t.Errorf("scrolled to the end, the last line %+v is not in view %+v", last, scroll.Rect())
	}
}
