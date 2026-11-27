package widget

import (
	"slices"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// setFeel switches ctx to its theme with one value of the feel set, as a theme
// file would state it.
func setFeel(t *testing.T, ctx *core.Ctx, name string, value int) {
	t.Helper()
	theme := ctx.Th
	feel, err := theme.Feel.Set(name, value)
	if err != nil {
		t.Fatal(err)
	}
	theme.Feel = feel
	if err := ctx.SetTheme(theme); err != nil {
		t.Fatal(err)
	}
}

// tipShownAfter reports whether a tooltip rested on for the given ticks at tps
// is shown.
func tipShownAfter(t *testing.T, ctx *core.Ctx, tps, ticks int) bool {
	t.Helper()
	state := &TooltipState{}
	root := Stack(Row(W(100), Tooltip("tip")), TooltipLayer(state))
	ctx.In, ctx.Tick = core.Input{X: 50, Y: 10, TPS: tps}, 100
	drawn(t, ctx, root, 200, 50)
	ctx.Tick += ticks
	drawn(t, ctx, root, 200, 50)
	return state.Shown()
}

func TestThemeSetsTooltipDelay(t *testing.T) {
	ctx := testCtx(t)
	setFeel(t, ctx, "tooltipDelay", 200)
	for _, test := range []struct {
		tps, ticks int
		want       bool
	}{{60, 11, false}, {60, 12, true}, {30, 5, false}, {30, 6, true}, {0, 11, false}, {0, 12, true}, {-1, 12, true}} {
		if got := tipShownAfter(t, ctx, test.tps, test.ticks); got != test.want {
			t.Errorf("%d ticks at %d TPS: shown = %v, want %v", test.ticks, test.tps, got, test.want)
		}
	}
	// A tooltip waiting across a theme switch is measured by the new delay.
	setFeel(t, ctx, "tooltipDelay", 1000)
	state := &TooltipState{}
	root := Stack(Row(W(100), Tooltip("tip")), TooltipLayer(state))
	ctx.In, ctx.Tick = core.Input{X: 50, Y: 10, TPS: 60}, 100
	drawn(t, ctx, root, 200, 50)
	ctx.Tick = 140
	drawn(t, ctx, root, 200, 50)
	if state.Shown() {
		t.Fatal("shown before 1s")
	}
	setFeel(t, ctx, "tooltipDelay", 500)
	drawn(t, ctx, root, 200, 50)
	if !state.Shown() {
		t.Fatal("the shorter delay has passed")
	}
}

func TestThemeSetsHoldDelay(t *testing.T) {
	ctx := testCtx(t)
	setFeel(t, ctx, "holdDelay", 500) // 30 ticks
	if got, want := holdSteps(t, ctx, 36), []int{30, 33, 36}; !slices.Equal(got, want) {
		t.Fatalf("hold stepped at %v, want %v", got, want)
	}
}

func TestThemeSetsHoldInterval(t *testing.T) {
	ctx := testCtx(t)
	setFeel(t, ctx, "holdInterval", 100) // 6 ticks
	if got, want := holdSteps(t, ctx, 32), []int{20, 26, 32}; !slices.Equal(got, want) {
		t.Fatalf("hold stepped at %v, want %v", got, want)
	}
}

// A repeat shorter than a tick, from a theme built in Go, is clamped and
// steps no more than once a tick.
func TestAnOutOfRangeHoldStepsAtMostOnceATick(t *testing.T) {
	ctx := testCtx(t)
	ctx.Th.Feel.HoldRepeat = core.Repeat{Delay: 1, Interval: 1}
	got := holdSteps(t, ctx, 9)
	if want := []int{6, 7, 8, 9}; !slices.Equal(got, want) { // delay clamped to 100ms, interval to 10ms
		t.Fatalf("hold stepped at %v, want %v", got, want)
	}
}

func TestThemeSetsPanSlop(t *testing.T) {
	ctx := testCtx(t)
	setFeel(t, ctx, "panSlop", 12)
	if got := panAfter(t, ctx, 11); got != 0 {
		t.Errorf("11px panned %d", got)
	}
	if got := panAfter(t, ctx, 12); got != 12 {
		t.Errorf("12px panned %d, want 12", got)
	}
}

func TestThemeSetsWheelStep(t *testing.T) {
	ctx := testCtx(t)
	setFeel(t, ctx, "wheelStep", 100)
	if down, sideways := scrollNotch(t, ctx); down != 100 || sideways != 100 {
		t.Fatalf("one notch moved %d and %d px, want 100 and 100", down, sideways)
	}
}

func TestThemeSetsWheelLines(t *testing.T) {
	ctx := testCtx(t)
	setFeel(t, ctx, "wheelLines", 5)
	wheelRows(t, ctx, 5, 6)
}

func TestThemeSetsWheelColumns(t *testing.T) {
	ctx := testCtx(t)
	setFeel(t, ctx, "wheelColumns", 9)
	wheelRows(t, ctx, 3, 9)
}

func TestThemeSetsToastTime(t *testing.T) {
	ctx := testCtx(t)
	setFeel(t, ctx, "toastTime", 8000)
	var stack ToastStack
	start := time.Unix(100, 0)
	stack.Show("first", ToastInfo, start, ctx.Feel())
	setFeel(t, ctx, "toastTime", 1000)
	stack.Show("second", ToastInfo, start, ctx.Feel())
	if stack.Prune(start.Add(time.Second - 1)) {
		t.Fatal("gone before its time")
	}
	if !stack.Prune(start.Add(time.Second)) || len(stack) != 1 || stack[0].Text != "first" {
		t.Fatalf("after 1s: %+v, want only the toast shown under 8s", stack)
	}
	if stack.Prune(start.Add(8*time.Second-1)) || !stack.Prune(start.Add(8*time.Second)) {
		t.Fatal("a shown toast keeps the 8s it was shown with")
	}
}

func TestAToastShownWithNoFeelOrAWildOneStaysInBounds(t *testing.T) {
	start := time.Unix(100, 0)
	for _, test := range []struct {
		given, want time.Duration
	}{{0, 3 * time.Second}, {-time.Hour, time.Second}, {1, time.Second}, {24 * time.Hour, time.Minute}} {
		var stack ToastStack
		stack.Show("saved", ToastInfo, start, core.Feel{ToastTime: test.given})
		if got := stack[0].Expires.Sub(start); got != test.want {
			t.Errorf("ToastTime %v: stays %v, want %v", test.given, got, test.want)
		}
	}
}
