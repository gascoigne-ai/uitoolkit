package sections

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/gascoigne-ai/uitoolkit/view"
)

var update = flag.Bool("update", false, "rewrite the pinned section dumps")

// contentW and contentH are the room of a section beside the nav and under
// the top bar.
const contentW, contentH = 844, 684

func layoutDump(ctx *core.Ctx, node view.Node) string {
	view.Layout(ctx, node, contentW, contentH)
	return view.Dump(node)
}

func TestSectionDumps(t *testing.T) {
	if len(Sections) == 0 {
		t.Fatal("no sections")
	}
	for _, section := range Sections {
		t.Run(section.ID, func(t *testing.T) {
			showcase, ctx, _ := newTestShowcase(t)
			open(t, showcase, ctx, section.ID)
			got := layoutDump(ctx, showcase.SectionView(ctx))
			path := filepath.Join("testdata", "dump", section.ID+".txt")
			if *update {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Errorf("dump differs from %s (rerun with -update):\n%s", path, got)
			}
		})
	}
}

// slotGaps are the places where a markup form embeds a Go node, because
// markup has no element for the widget.
var slotGaps = []struct {
	id     string
	source string
	slots  []string
}{
	{"buttons", buttonsSource, []string{"keys"}},
	{"text", textSource, nil},
	{"toggles", togglesSource, []string{"autosave", "disabled", "offline"}},
	{"lists", listsSource, nil},
	{"layout", layoutSource, []string{"meter", "resizable"}},
	{"overlays", overlaysSource, nil},
	{"code", codeSource, nil},
	{"images", imagesSource, nil},
	{"typography", typographySource, []string{"paragraph"}},
	{"theming", themingSource, []string{"switch"}},
}

var slotName = regexp.MustCompile(`<slot name="([^"]+)"`)

// painted is what a laid-out node paints: it shows what Dump cannot, such
// as the value of a checkbox, a slider or a switch.
func painted(ctx *core.Ctx, node view.Node) string {
	canvas := rendertest.NewCanvas(contentW, contentH)
	view.Draw(ctx, node, canvas)
	var ops strings.Builder
	for _, op := range canvas.Ops() {
		fmt.Fprintf(&ops, "%T %+v\n", op, op)
	}
	return ops.String()
}

func TestGoAndMarkupFormsDumpAlike(t *testing.T) {
	if len(slotGaps) != len(Sections) {
		t.Fatalf("%d sections, %d rows of slot gaps", len(Sections), len(slotGaps))
	}
	for _, gap := range slotGaps {
		t.Run(gap.id, func(t *testing.T) {
			var used []string
			for _, match := range slotName.FindAllStringSubmatch(gap.source, -1) {
				used = append(used, match[1])
			}
			if got := slices.Sorted(slices.Values(used)); !slices.Equal(got, gap.slots) {
				t.Errorf("slots in the template = %v, recorded %v", got, gap.slots)
			}
			showcase, ctx, _ := newTestShowcase(t)
			open(t, showcase, ctx, gap.id)
			section := Sections[showcase.Section]
			goNode := section.Go(showcase, ctx)
			goDump, goOps := layoutDump(ctx, goNode), painted(ctx, goNode)
			markupNode := section.Markup(showcase, ctx)
			markupDump, markupOps := layoutDump(ctx, markupNode), painted(ctx, markupNode)
			if goDump != markupDump {
				t.Errorf("Go:\n%s\nmarkup:\n%s", goDump, markupDump)
			} else if goOps != markupOps {
				t.Errorf("same dump, different painting\nGo:\n%s\nmarkup:\n%s", goOps, markupOps)
			}
		})
	}
}

// The dialogs, the toasts and the popup menu are layers of the window, so their two forms
// are compared on their own.
func TestGoAndMarkupLayersDumpAlike(t *testing.T) {
	showcase, ctx, _ := newTestShowcase(t)
	open(t, showcase, ctx, "overlays")
	tests := []struct {
		name, shows string
		show        func()
	}{
		{"menu", `"New file"`, func() { showcase.openMenu(ctx, 40, 60) }},
		{"dialog", `"Delete project?"`, func() { showcase.openDialog() }},
		{"panel", `"Preferences"`, showcase.openPanel},
		{"resizable", "resizableB", showcase.openSized},
		// last: closeLayers leaves the toasts, which would show in later rows
		{"toasts", `"Export failed"`, func() { postBothToasts(showcase) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			showcase.closeLayers(ctx)
			test.show()
			var dumps []string
			for form := range Forms {
				showcase.Form = form
				dumps = append(dumps, layoutDump(ctx, view.Stack(showcase.layers(ctx, contentW, contentH))))
			}
			if len(dumps) != 2 || dumps[0] != dumps[1] || !strings.Contains(dumps[0], test.shows) {
				t.Errorf("layers of the two forms differ or miss %s:\n%s", test.shows, strings.Join(dumps, "\n---\n"))
			}
		})
	}
}
