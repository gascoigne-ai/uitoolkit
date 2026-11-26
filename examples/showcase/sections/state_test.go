package sections

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/widget"
)

func TestSectionStateSurvivesSwitching(t *testing.T) {
	showcase, ctx, _ := newTestShowcase(t)
	open(t, showcase, ctx, "text")
	frame(t, showcase, ctx)
	showcase.widgets.name.Insert(" typed")
	edited := showcase.widgets.name.Text
	open(t, showcase, ctx, "lists")
	frame(t, showcase, ctx)
	showcase.widgets.list.Y = 40
	showcase.widgets.picked = 3
	open(t, showcase, ctx, "buttons")
	frame(t, showcase, ctx)
	for _, id := range []string{"text", "lists"} {
		open(t, showcase, ctx, id)
		frame(t, showcase, ctx)
	}
	if showcase.widgets.name.Text != edited || edited == "" {
		t.Errorf("name = %q, want %q", showcase.widgets.name.Text, edited)
	}
	if showcase.widgets.list.Y != 40 || showcase.widgets.picked != 3 {
		t.Errorf("list scroll %d, picked %d, want 40 and 3", showcase.widgets.list.Y, showcase.widgets.picked)
	}
}

// A popup menu lives in the context, not in a section, so the next section
// would open under it.
func TestSwitchingSectionsClosesPopupsAndDialog(t *testing.T) {
	showcase, ctx, _ := newTestShowcase(t)
	open(t, showcase, ctx, "overlays")
	frame(t, showcase, ctx)
	showcase.openMenu(ctx, 10, 10)
	showcase.openDialog()
	open(t, showcase, ctx, "buttons")
	frame(t, showcase, ctx)
	if widget.MenuOpen(ctx) || showcase.dialog != nil {
		t.Errorf("menu open = %v, dialog = %v after switching, want neither", widget.MenuOpen(ctx), showcase.dialog)
	}
}
