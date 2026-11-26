package sections

import (
	"slices"
	"strings"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/view"
)

var wantSections = []string{"buttons", "text", "toggles", "lists", "overlays", "code", "images", "typography", "theming"}

func TestSectionsTable(t *testing.T) {
	var ids, labels []string
	for _, section := range Sections {
		ids = append(ids, section.ID)
		labels = append(labels, section.Label)
		if section.ID == "" || section.Label == "" {
			t.Errorf("section %+v has no ID or label", section)
		}
		if section.Go == nil || section.Markup == nil {
			t.Errorf("section %q lacks a Go or a markup form", section.ID)
		}
	}
	if !slices.Equal(ids, wantSections) {
		t.Fatalf("section ids = %v, want %v", ids, wantSections)
	}
	if unique := slices.Compact(slices.Sorted(slices.Values(labels))); len(unique) != len(labels) {
		t.Errorf("labels repeat: %v", labels)
	}
	showcase, ctx, _ := newTestShowcase(t)
	nav := showcase.Nav(ctx)
	view.Layout(ctx, nav, 180, testH)
	dump, at := view.Dump(nav), 0
	for _, label := range labels {
		found := strings.Index(dump[at:], `"`+label+`"`)
		if found < 0 {
			t.Fatalf("nav does not list %q in order:\n%s", label, dump)
		}
		at += found
	}
}
