package view

import (
	"strings"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/ebitenrender"
	"github.com/gascoigne-ai/uitoolkit/render"
	"golang.org/x/image/font/gofont/gomono"
)

func monoSource(t *testing.T) render.FontSource {
	t.Helper()
	source, err := ebitenrender.NewDevice().NewFontSource(gomono.TTF)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestTextWithAFontSourceIsMeasuredInThatFont(t *testing.T) {
	c := testCtx(t)
	plainText := Text("iiiiii")
	mono := Text("iiiiii", FontSource(monoSource(t)))
	Layout(c, Row(plainText, mono), 600, 30)
	if mono.rect.W <= plainText.rect.W {
		t.Fatalf("a monospace \"i\" is wider than a proportional one: %d vs %d", mono.rect.W, plainText.rect.W)
	}
}

// A font list keeps a nil pointer for a file it could not read. In the
// interface it is not equal to nil, and measuring through it would crash.
func TestANilPointerAsAFontSourceKeepsTheInterfaceFont(t *testing.T) {
	unread := Text("iiiiii", FontSource((*ebitenrender.FontSource)(nil)))
	if unread.fontSource != nil {
		t.Fatalf("fontSource = %#v, want nil: the interface font", unread.fontSource)
	}
	c := testCtx(t)
	plainText := Text("iiiiii")
	root := Row(plainText, unread)
	Layout(c, root, 600, 30)
	if unread.rect.W != plainText.rect.W {
		t.Errorf("width %d, want the interface font's %d", unread.rect.W, plainText.rect.W)
	}
	if dump := Dump(root); strings.Contains(dump, "font=") {
		t.Errorf("dump marks another font:\n%s", dump)
	}
}

func TestADumpMarksTextDrawnInAnotherFont(t *testing.T) {
	dump := Dump(Row(Text("a"), Text("b", FontSource(monoSource(t)))))
	if strings.Count(dump, "font=") != 1 {
		t.Fatalf("dump:\n%s", dump)
	}
}
