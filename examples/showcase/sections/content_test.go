package sections

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/view"
	"github.com/gascoigne-ai/uitoolkit/widget"
)

// walk calls visit for node and every element under it.
func walk(node *view.El, visit func(*view.El)) {
	visit(node)
	for _, child := range node.Children() {
		walk(child, visit)
	}
}

func TestIconsSectionShowsEveryIcon(t *testing.T) {
	showcase, ctx, _ := newTestShowcase(t)
	open(t, showcase, ctx, "images")
	for form := range []int{0, 1} {
		showcase.Form = form
		node := showcase.SectionView(ctx)
		view.Layout(ctx, node, contentW, contentH)
		var drawn []string
		walk(node, func(element *view.El) {
			behavior := element.Behavior()
			if behavior == nil || !strings.HasSuffix(reflect.TypeOf(behavior).String(), "iconB") {
				return
			}
			drawn = append(drawn, reflect.ValueOf(behavior).FieldByName("name").String())
		})
		if got, want := slices.Sorted(slices.Values(drawn)), widget.IconNames(); !slices.Equal(got, want) {
			t.Errorf("form %d draws icons %v, want %v", form, got, want)
		}
		dump := view.Dump(node)
		for _, name := range widget.IconNames() {
			if !strings.Contains(dump, fmt.Sprintf("%q", name)) {
				t.Errorf("form %d has no label for icon %q", form, name)
			}
		}
	}
}

func TestCodeSectionIsLong(t *testing.T) {
	first, second := demoSource(), demoSource()
	if len(first) < 2000 {
		t.Fatalf("the demo source has %d lines, want thousands", len(first))
	}
	if !reflect.DeepEqual(first, second) {
		t.Error("the demo source is not deterministic")
	}
	showcase, ctx, _ := newTestShowcase(t)
	open(t, showcase, ctx, "code")
	frame(t, showcase, ctx)
	if got := len(showcase.widgets.code.Lines); got != len(first) {
		t.Errorf("the code view holds %d lines, want %d", got, len(first))
	}
}
