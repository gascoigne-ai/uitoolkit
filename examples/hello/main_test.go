package main

import (
	"strings"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/renderer/ebiten"
	"github.com/gascoigne-ai/uitoolkit/theming"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// The example's view builds, lays out and reacts to its state.
func TestHelloViewFollowsItsState(t *testing.T) {
	ctx, err := core.NewCtx(theming.Dark(), ebitenrender.NewDevice())
	if err != nil {
		t.Fatal(err)
	}
	g := &hello{ctx: ctx, name: NewTextField("")}
	root := g.view()
	Layout(ctx, root, 480, 320)
	if d := Dump(root); !strings.Contains(d, `"Nobody yet"`) {
		t.Fatalf("empty state:\n%s", d)
	}
	g.name = NewTextField("Ada")
	g.greet()
	if d := Dump(g.view()); !strings.Contains(d, `"Hello, Ada!"`) || strings.Contains(d, "Nobody yet") {
		t.Fatalf("after greeting:\n%s", d)
	}
}

// The window is laid out in logical pixels and the canvas is in device pixels,
// so on a HiDPI display the text is drawn larger and stays sharp.
func TestHelloLayoutScalesToDevicePixels(t *testing.T) {
	ctx, err := core.NewCtx(theming.Dark(), ebitenrender.NewDevice())
	if err != nil {
		t.Fatal(err)
	}
	g := &hello{ctx: ctx, name: NewTextField("")}
	width, height := g.Layout(480, 320, 1.5)
	if width != 720 || height != 480 {
		t.Errorf("canvas = %dx%d, want 720x480", width, height)
	}
	if ctx.Scale != 1.5 {
		t.Errorf("ctx.Scale = %v, want 1.5", ctx.Scale)
	}
	if g.w != 480 || g.h != 320 {
		t.Errorf("view area = %dx%d, want the logical 480x320", g.w, g.h)
	}
}
