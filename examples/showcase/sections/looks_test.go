package sections

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/gascoigne-ai/uitoolkit/view"
)

// The sample field of the theming section was laid out with no height, and
// its buttons with the height of their label: in a row that centers its
// items each takes the height it measures, and a field measures none.
func TestControlsHaveTheirHeight(t *testing.T) {
	for _, section := range Sections {
		for form := range Forms {
			showcase, ctx, _ := newTestShowcase(t)
			open(t, showcase, ctx, section.ID)
			showcase.SetForm(form)
			node := showcase.SectionView(ctx)
			view.Layout(ctx, node, contentW, contentH)
			walk(node, func(element *view.El) {
				switch element.ControlName() {
				case "button", "primary", "icon-button", "field":
					if rect := element.Rect(); rect.H < ctx.Sizes.Compact || rect.W == 0 {
						t.Errorf("%s (%s): a %s is %dx%d", section.ID, Forms[form].ID, element.ControlName(), rect.W, rect.H)
					}
				}
			})
		}
	}
}

// The paragraph of the typography section was drawn over the caption under
// it: with MaxW its height was measured at a wider width than it wraps at.
func TestParagraphStaysInItsBox(t *testing.T) {
	showcase, ctx, _ := newTestShowcase(t)
	open(t, showcase, ctx, "typography")
	node := showcase.SectionView(ctx)
	view.Layout(ctx, node, contentW, contentH)
	found := false
	walk(node, func(element *view.El) {
		behavior := element.Behavior()
		if behavior == nil || !strings.HasSuffix(reflect.TypeOf(behavior).String(), "paragraphB") {
			return
		}
		found = true
		canvas := rendertest.NewCanvas(contentW, contentH)
		view.Draw(ctx, element, canvas)
		bottom := float64(element.Rect().Y + element.Rect().H)
		for _, op := range canvas.Ops() {
			if drawn, ok := op.(rendertest.DrawText); ok && drawn.Top+float64(ctx.LineH(drawn.Face)) > bottom {
				t.Errorf("line %q at %v ends below the box, which ends at %v", drawn.Text, drawn.Top, bottom)
			}
		}
	})
	if !found {
		t.Fatal("no paragraph in the typography section")
	}
}

// The posed code view scrolled by a guessed line height and left the folded
// block off the screen.
func TestCodePoseShowsAFoldedBlock(t *testing.T) {
	showcase, ctx, _ := newTestShowcase(t)
	open(t, showcase, ctx, "code")
	frame(t, showcase, ctx)
	codePoses[0].Apply(showcase, ctx)
	frame(t, showcase, ctx)
	if drawn := drawnText(frame(t, showcase, ctx)); !strings.Contains(drawn, "... }") {
		t.Errorf("no folded block on screen:\n%s", drawn)
	}
	// The if block is its first line, one inner line and the closing brace.
	block := headerLines + posedFunc*funcLines + branchLine
	if visible := showcase.widgets.code.Visible(); slices.Contains(visible, block+1) || slices.Contains(visible, block+2) {
		t.Error("folding the block does not hide all of it")
	}
}

// Only the Go form keeps the elements the poses open the menu and rest the
// pointer at, so a pose applied after drawing the markup form dereferenced
// nil.
func TestOverlayPosesDoNotPanicInEitherForm(t *testing.T) {
	for form := range Forms {
		for _, pose := range overlaysPoses {
			showcase, ctx, _ := newTestShowcase(t)
			open(t, showcase, ctx, "overlays")
			showcase.SetForm(form)
			frame(t, showcase, ctx)
			pose.Apply(showcase, ctx)
		}
	}
}
