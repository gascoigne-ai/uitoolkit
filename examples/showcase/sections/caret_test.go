package sections

import (
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
)

func TestCaretChoicesSetTheContext(t *testing.T) {
	for index, form := range Forms {
		t.Run(form.Label, func(t *testing.T) {
			showcase, ctx, _ := newTestShowcase(t)
			showcase.SetForm(index)
			open(t, showcase, ctx, "text")
			frame(t, showcase, ctx)
			steps := []struct {
				click string
				want  core.Caret
			}{
				{`"Underline"`, core.Caret{Shape: core.CaretUnderline}},
				{`"Bar"`, core.Caret{Shape: core.CaretBar}},
				{`"Blink"`, core.Caret{Shape: core.CaretBar, Blink: true}},
				{`"Block"`, core.Caret{Shape: core.CaretBlock, Blink: true}},
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

	showcase.Caret.Blink = true
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
