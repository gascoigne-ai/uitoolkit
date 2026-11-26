package sections

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/gascoigne-ai/uitoolkit/view"
)

func TestCaretChoicesSetTheContext(t *testing.T) {
	for index, form := range Forms {
		t.Run(form.Label, func(t *testing.T) {
			showcase, ctx, _ := newTestShowcase(t)
			showcase.SetForm(index)
			open(t, showcase, ctx, "text")
			frame(t, showcase, ctx)
			if showcase.Caret != (core.Caret{}) || ctx.Caret != (core.Caret{}) {
				t.Fatalf("at the start: showcase %+v, context %+v, want the zero caret", showcase.Caret, ctx.Caret)
			}
			if !blinkChecked(t, showcase, ctx) {
				t.Fatal("at the start: Blink is not checked")
			}
			steps := []struct {
				click string
				want  core.Caret
			}{
				{`"Underline"`, core.Caret{Shape: core.CaretUnderline}},
				{`"Bar"`, core.Caret{Shape: core.CaretBar}},
				{`"Blink"`, core.Caret{Shape: core.CaretBar, Steady: true}},
				{`"Block"`, core.Caret{Shape: core.CaretBlock, Steady: true}},
				{`"Blink"`, core.Caret{}},
			}
			for _, step := range steps {
				clickOn(t, showcase, ctx, step.click)
				if showcase.Caret != step.want || ctx.Caret != step.want {
					t.Fatalf("after %s: showcase %+v, context %+v, want %+v", step.click, showcase.Caret, ctx.Caret, step.want)
				}
			}
		})
	}
}

// The toolkit restarts no blink by itself, so the app does.
func TestTypingRestartsTheCaret(t *testing.T) {
	showcase, ctx, clock := newTestShowcase(t)
	open(t, showcase, ctx, "text")
	later := func(in core.Input) {
		clock.at = clock.at.Add(700 * time.Millisecond)
		input(t, showcase, ctx, in)
	}

	showcase.Caret.Steady = true
	later(core.Input{})
	showcase.Caret.Steady = false
	later(core.Input{})
	if !ctx.CaretSince.Equal(clock.at) {
		t.Errorf("turning the blink on: CaretSince = %v, want %v", ctx.CaretSince, clock.at)
	}
	turnedOn := clock.at
	later(core.Input{X: 300, Y: 300}) // a pointer that only moves
	if !ctx.CaretSince.Equal(turnedOn) {
		t.Errorf("a moving pointer restarted the caret: CaretSince = %v", ctx.CaretSince)
	}
	for name, in := range map[string]core.Input{
		"typing":         {Typed: []rune("a")},
		"an editing key": {Edits: []core.Edit{core.EditLeft}},
		"a key":          {Keys: []render.Key{render.KeyArrowLeft}},
	} {
		later(in)
		if !ctx.CaretSince.Equal(clock.at) {
			t.Errorf("%s: CaretSince = %v, want %v", name, ctx.CaretSince, clock.at)
		}
	}
}

// blinkChecked reports whether the checkbox of the caret controls, the only
// one in the text section, is drawn checked: it fills its box with Accent.
func blinkChecked(t *testing.T, showcase *Showcase, ctx *core.Ctx) bool {
	t.Helper()
	root := showcase.SectionView(ctx)
	view.Layout(ctx, root, contentW, contentH)
	found, checked := 0, false
	walk(root, func(element *view.El) {
		behavior := element.Behavior()
		if behavior == nil || !strings.HasSuffix(reflect.TypeOf(behavior).String(), "checkB") {
			return
		}
		found++
		canvas := rendertest.NewCanvas(contentW, contentH)
		view.Draw(ctx, element, canvas)
		accent := ctx.Color(core.Accent)
		for _, op := range canvas.Ops() {
			switch fill := op.(type) {
			case rendertest.FillRect:
				checked = checked || fill.Color == accent
			case rendertest.FillRoundRect:
				checked = checked || fill.Top == accent
			}
		}
	})
	if found != 1 {
		t.Fatalf("%d checkboxes in the text section, want 1", found)
	}
	return checked
}
