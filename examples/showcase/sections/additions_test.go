package sections

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render/rendertest"
	"github.com/gascoigne-ai/uitoolkit/view"
	"github.com/gascoigne-ai/uitoolkit/widget"
)

// pageDump is the dump of the section and the layers over it.
func pageDump(showcase *Showcase, ctx *core.Ctx) string {
	return layoutDump(ctx, showcase.SectionView(ctx)) +
		layoutDump(ctx, view.Stack(showcase.layers(ctx, contentW, contentH)))
}

func postBothToasts(showcase *Showcase) {
	showcase.widgets.toasts.Show("Settings saved", widget.ToastInfo, showcase.now())
	showcase.widgets.toasts.Show("Export failed", widget.ToastError, showcase.now())
}

func TestSectionsShowTheWidgets(t *testing.T) {
	tests := []struct {
		name, section string
		open          func(*Showcase)
		needles       []string
	}{
		{"key chips", "buttons", nil, []string{`"Ctrl"`, `"Shift"`, `"P"`, `","`, `"+"`}},
		{"text menus", "text", nil, []string{
			"fieldB w=320 bg=sunken control=field onRightClick=true rect=114,39,",
			"fieldB w=320 bg=sunken control=field onRightClick=true rect=114,121,",
		}},
		{"dialog parts", "overlays", (*Showcase).openPanel, []string{
			`"Preferences"`, `"Editor"`, `"Modified"`, "tip=Close (Esc)", "borderB=border-soft", "focusScope=true",
		}},
		{"resizable dialog", "overlays", (*Showcase).openSized, []string{"resizableB", "w=480 h=320", `"Size: 480 × 320 (default)"`}},
		{"toasts", "overlays", postBothToasts, []string{`"Settings saved"`, `"Export failed"`, `"2 shown"`}},
		{"stack", "layout", nil, []string{" stack h=72 bg=sunken control=table", "float=true"}},
		{"hscroll", "layout", nil, []string{"hscrollB", `"Tag 1"`, `"Tag 24"`}},
		{"inset rule", "layout", nil, []string{"flow w=1 h=29 padT=6 padB=6"}},
		{"lines", "layout", nil, []string{`"Lines draws one text per line,"`, `"so a short note keeps its breaks"`, `"without a paragraph."`}},
		{"custom", "layout", nil, []string{"meterB meter=65"}},
		{"resizable", "layout", nil, []string{"resizableB w=240 h=80", `"240 × 80"`}},
	}
	for _, test := range tests {
		for form := range Forms {
			t.Run(test.name+"/"+Forms[form].ID, func(t *testing.T) {
				showcase, ctx, _ := newTestShowcase(t)
				open(t, showcase, ctx, test.section)
				showcase.SetForm(form)
				if test.open != nil {
					test.open(showcase)
				}
				dump := pageDump(showcase, ctx)
				for _, needle := range test.needles {
					if !strings.Contains(dump, needle) {
						t.Errorf("no %s in:\n%s", needle, dump)
					}
				}
			})
		}
	}
}

var dumpedRect = regexp.MustCompile(`rect=(-?\d+),(-?\d+),(\d+),(\d+)`)

// windowRect is the rect of the first element of the whole window whose
// dump line holds needle.
func windowRect(t *testing.T, showcase *Showcase, ctx *core.Ctx, needle string) core.Rect {
	t.Helper()
	root := showcase.View(ctx, testW, testH)
	view.Layout(ctx, root, testW, testH)
	dump := view.Dump(root)
	for _, line := range strings.Split(dump, "\n") {
		if found := dumpedRect.FindStringSubmatch(line); found != nil && strings.Contains(line, needle) {
			var numbers [4]int
			for i := range numbers {
				numbers[i], _ = strconv.Atoi(found[i+1])
			}
			return core.Rect{X: numbers[0], Y: numbers[1], W: numbers[2], H: numbers[3]}
		}
	}
	t.Fatalf("no element with %s:\n%s", needle, dump)
	return core.Rect{}
}

// input runs one frame with in.
func input(t *testing.T, showcase *Showcase, ctx *core.Ctx, in core.Input) *rendertest.Canvas {
	t.Helper()
	ctx.In = in
	canvas := frame(t, showcase, ctx)
	ctx.In = core.Input{}
	return canvas
}

func click(t *testing.T, showcase *Showcase, ctx *core.Ctx, posX, posY int) {
	t.Helper()
	input(t, showcase, ctx, core.Input{X: posX, Y: posY, Click: true, Held: true})
}

func clickOn(t *testing.T, showcase *Showcase, ctx *core.Ctx, needle string) {
	t.Helper()
	rect := windowRect(t, showcase, ctx, needle)
	click(t, showcase, ctx, rect.X+rect.W/2, rect.Y+rect.H/2)
}

var dialogOpeners = []struct {
	name string
	kind dialogKind
	open func(*Showcase)
}{
	{"message", messageDialog, (*Showcase).openDialog},
	{"panel", panelDialog, (*Showcase).openPanel},
	{"sized", sizedDialog, (*Showcase).openSized},
}

func TestEscapeClosesTheMenuThenTheDialog(t *testing.T) {
	for _, test := range dialogOpeners {
		t.Run(test.name, func(t *testing.T) {
			showcase, ctx, _ := newTestShowcase(t)
			open(t, showcase, ctx, "overlays")
			test.open(showcase)
			showcase.openMenu(ctx, 40, 60)
			if showcase.dialog != test.kind {
				t.Fatalf("dialog = %v, want %v", showcase.dialog, test.kind)
			}
			input(t, showcase, ctx, core.Input{Escape: true})
			if widget.MenuOpen(ctx) || showcase.dialog != test.kind {
				t.Fatalf("after one Escape: menu open = %v, dialog = %v", widget.MenuOpen(ctx), showcase.dialog)
			}
			input(t, showcase, ctx, core.Input{Escape: true})
			if showcase.dialog != noDialog {
				t.Errorf("after two Escapes the dialog is %v", showcase.dialog)
			}
		})
	}
}

func TestSwitchingSectionsClosesEveryDialog(t *testing.T) {
	for _, test := range dialogOpeners {
		t.Run(test.name, func(t *testing.T) {
			showcase, ctx, _ := newTestShowcase(t)
			open(t, showcase, ctx, "overlays")
			test.open(showcase)
			open(t, showcase, ctx, "buttons")
			if showcase.dialog != noDialog {
				t.Errorf("dialog = %v after switching", showcase.dialog)
			}
		})
	}
}

// The toolkit's dialogs keep open on a click outside them
// (view.KeepOnOutsideClick); the close button of the title bar closes them.
func TestOnlyTheTitleButtonClosesADialogByPointer(t *testing.T) {
	for _, test := range dialogOpeners {
		for form := range Forms {
			t.Run(test.name+"/"+Forms[form].ID, func(t *testing.T) {
				showcase, ctx, _ := newTestShowcase(t)
				open(t, showcase, ctx, "overlays")
				showcase.SetForm(form)
				test.open(showcase)
				frame(t, showcase, ctx)
				click(t, showcase, ctx, 5, testH-5)
				if showcase.dialog != test.kind {
					t.Fatalf("a click outside left dialog %v, want %v", showcase.dialog, test.kind)
				}
				clickOn(t, showcase, ctx, "tip=Close (Esc)")
				if showcase.dialog != noDialog {
					t.Errorf("the close button left dialog %v", showcase.dialog)
				}
			})
		}
	}
}

func TestAnsweringTheMessageDialogClosesIt(t *testing.T) {
	showcase, ctx, _ := newTestShowcase(t)
	open(t, showcase, ctx, "overlays")
	showcase.openDialog()
	frame(t, showcase, ctx)
	clickOn(t, showcase, ctx, `text="Cancel"`)
	if showcase.dialog != noDialog || showcase.widgets.lastAnswer != "cancelled" {
		t.Errorf("dialog = %v, answer %q", showcase.dialog, showcase.widgets.lastAnswer)
	}
}

func TestToastLifetime(t *testing.T) {
	showcase, ctx, clock := newTestShowcase(t)
	open(t, showcase, ctx, "overlays")
	frame(t, showcase, ctx)
	clickOn(t, showcase, ctx, `text="Info toast"`)
	clickOn(t, showcase, ctx, `text="Error toast"`)
	if len(showcase.widgets.toasts) != 2 {
		t.Fatalf("the buttons posted %+v", showcase.widgets.toasts)
	}
	open(t, showcase, ctx, "buttons")
	input(t, showcase, ctx, core.Input{Escape: true})
	clock.at = clock.at.Add(widget.ToastInfoFor - time.Millisecond)
	frame(t, showcase, ctx)
	if len(showcase.widgets.toasts) != 2 {
		t.Fatalf("a section switch or Escape or an early prune lost a toast: %+v", showcase.widgets.toasts)
	}
	clock.at = clock.at.Add(time.Millisecond)
	if drawn := drawnText(input(t, showcase, ctx, core.Input{})); !strings.Contains(drawn, errorToast+"\n") ||
		strings.Contains(drawn, infoToast+"\n") {
		t.Fatalf("after its time only the error is drawn:\n%s", drawn)
	}
	clock.at = clock.at.Add(time.Hour)
	frame(t, showcase, ctx)
	if len(showcase.widgets.toasts) != 1 || showcase.widgets.toasts[0].Kind != widget.ToastError {
		t.Fatalf("the error toast stays: %+v", showcase.widgets.toasts)
	}
	clickOn(t, showcase, ctx, `text="`+errorToast+`"`)
	if len(showcase.widgets.toasts) != 0 {
		t.Errorf("a click left %+v", showcase.widgets.toasts)
	}
}

// edgeDrag presses on the right edge of rect and moves the pointer by delta.
func edgeDrag(t *testing.T, showcase *Showcase, ctx *core.Ctx, rect core.Rect, delta int) {
	t.Helper()
	posX, posY := rect.X+rect.W-2, rect.Y+rect.H/2
	click(t, showcase, ctx, posX, posY)
	input(t, showcase, ctx, core.Input{X: posX + delta, Y: posY, Held: true})
	input(t, showcase, ctx, core.Input{X: posX + delta, Y: posY})
}

func TestDraggingTheBoxEdgeResizesIt(t *testing.T) {
	tests := []struct {
		name         string
		delta, wantW int
	}{
		{"outward, twice the movement", 20, 280},
		// The pointer stays inside the page: once it leaves a clipping parent,
		// the toolkit hands the dragged element a pointer far off the window,
		// and a drag of the right edge snaps to the minimum.
		{"up to the maximum", 250, 600},
		{"down to the minimum", -200, 160},
	}
	for _, test := range tests {
		for form := range Forms {
			t.Run(test.name+"/"+Forms[form].ID, func(t *testing.T) {
				showcase, ctx, _ := newTestShowcase(t)
				open(t, showcase, ctx, "layout")
				showcase.SetForm(form)
				frame(t, showcase, ctx)
				edgeDrag(t, showcase, ctx, windowRect(t, showcase, ctx, "resizableB"), test.delta)
				rect := windowRect(t, showcase, ctx, "resizableB")
				if showcase.widgets.box.W != test.wantW || rect.W != test.wantW {
					t.Fatalf("state %d, drawn %d, want %d", showcase.widgets.box.W, rect.W, test.wantW)
				}
				input(t, showcase, ctx, core.Input{X: rect.X + rect.W - 2, Y: rect.Y + rect.H/2, Click: true, Held: true, Clicks: 2})
				if showcase.widgets.box.W != 0 || windowRect(t, showcase, ctx, "resizableB").W != boxW {
					t.Errorf("a double click left width %d", showcase.widgets.box.W)
				}
			})
		}
	}
}

func TestResizingTheDialogUpdatesTheSizeText(t *testing.T) {
	showcase, ctx, _ := newTestShowcase(t)
	open(t, showcase, ctx, "overlays")
	showcase.openSized()
	frame(t, showcase, ctx)
	edgeDrag(t, showcase, ctx, windowRect(t, showcase, ctx, "resizableB"), 20)
	if rect := windowRect(t, showcase, ctx, "resizableB"); rect.W != 520 || rect.H != 320 {
		t.Errorf("the dialog is %dx%d, want 520x320", rect.W, rect.H)
	}
	if dump := layoutDump(ctx, showcase.SectionView(ctx)); !strings.Contains(dump, `"Size: 520 × 320"`) {
		t.Errorf("the size text did not follow:\n%s", dump)
	}
}
