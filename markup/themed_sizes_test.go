package markup

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/theming"
	"github.com/gascoigne-ai/uitoolkit/view"
)

// The size names of a template come from the sizes the Ctx draws with, so a
// themed size lays a template out as the Go options for it do.
func TestTemplateSizesComeFromThemedCtx(t *testing.T) {
	var metrics core.Metrics
	for name, pixels := range map[string]int{"gutter": 21, "rowH": 50, "looseGap": 14, "dialogW": 200} {
		pinned, err := metrics.Pin(name, pixels)
		if err != nil {
			t.Fatal(err)
		}
		metrics = pinned
	}
	theme := theming.Dark()
	theme.Metrics = metrics
	ctx, err := core.NewCtx(theme, testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	template := Must("t.vml", `<col pad="{{gutter}}" gap="{{looseGap}}"><row h="{{rowH}}" w="{{dialogW}}"/><row h="{{rowH}}"/></col>`)
	fromTemplate := template.MustBuild(Scope(ctx.SizesScope()))
	fromGo := view.Col(view.PadOf(core.GutterOf), view.GapOf(core.LooseGapOf),
		view.Row(view.HOf(core.RowOf), view.WOf(core.DialogWOf)), view.Row(view.HOf(core.RowOf)))
	view.Layout(ctx, fromTemplate, 400, 300)
	view.Layout(ctx, fromGo, 400, 300)
	got, want := view.Dump(fromTemplate), view.Dump(fromGo)
	if got != want {
		t.Errorf("template =\n%s\nGo =\n%s", got, want)
	}
	if first := fromGo.Children()[0].Rect(); first != (core.Rect{X: 21, Y: 21, W: 200, H: 50}) {
		t.Errorf("first row = %+v, want it at the themed gutter with the themed size", first)
	}
}
