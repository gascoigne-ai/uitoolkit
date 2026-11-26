package core

import (
	"strings"
	"testing"
	"time"
)

func TestTheCaretBlinksOnAWallClockAndRestartsOnActivity(t *testing.T) {
	ctx, err := NewCtx(testTheme(), testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	ctx.Now = func() time.Time { return now }
	ctx.RestartCaret()
	steps := []struct {
		after time.Duration
		want  bool
	}{
		{0, true}, {100 * time.Millisecond, true}, {499 * time.Millisecond, true},
		{500 * time.Millisecond, false}, {900 * time.Millisecond, false},
		{1000 * time.Millisecond, true}, {1500 * time.Millisecond, false},
	}
	start := now
	for _, step := range steps {
		now = start.Add(step.after)
		if got := ctx.CaretVisible(); got != step.want {
			t.Errorf("after %v: visible = %v, want %v", step.after, got, step.want)
		}
	}
	now = start.Add(700 * time.Millisecond) // in an off phase
	ctx.RestartCaret()
	if !ctx.CaretVisible() {
		t.Fatal("activity must show the caret again")
	}
}

func TestTheBlinkRateDoesNotDependOnTheUpdateRate(t *testing.T) {
	// the phase is a function of time only: the same instant gives the same
	// answer however many ticks have passed
	ctx, _ := NewCtx(testTheme(), testDevice{})
	start := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	ctx.Now = func() time.Time { return start.Add(600 * time.Millisecond) }
	ctx.CaretSince = start
	ctx.Tick = 7
	first := ctx.CaretVisible()
	ctx.Tick = 7000
	if ctx.CaretVisible() != first {
		t.Fatal("the caret phase changed with the tick count")
	}
}

// CursorAt measured every prefix of the text from its start, so a click in
// a long single line took time quadratic in its length: minutes for a
// paste of a few hundred KB.
func TestCursorAtOnALongLineIsFastAndStillExact(t *testing.T) {
	ctx, err := NewCtx(testTheme(), testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Repeat("abcdefghij", 4000) // 40k runes
	start := time.Now()
	got := ctx.CursorAt(text, ctx.Width(ctx.F.UI, text[:20000])+1)
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("CursorAt took %v on %d runes", elapsed, len(text))
	}
	if got != 20000 {
		t.Errorf("CursorAt = %d, want 20000", got)
	}
	short := "héllo wörld"
	for i := range len([]rune(short)) + 1 {
		x := ctx.Width(ctx.F.UI, string([]rune(short)[:i]))
		if got := ctx.CursorAt(short, x); got != i {
			t.Errorf("CursorAt at the edge of rune %d = %d", i, got)
		}
	}
	if got := ctx.CursorAt(short, -50); got != 0 {
		t.Errorf("left of the text: %d", got)
	}
	if got := ctx.CursorAt(short, 99999); got != len([]rune(short)) {
		t.Errorf("right of the text: %d", got)
	}
}

// A drawing kept across frames needs to know whether it shows the blink
// phase; CaretReads counts the reads, whatever they answer.
func TestCaretVisibleCountsReads(t *testing.T) {
	ctx, err := NewCtx(testTheme(), testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.CaretReads != 0 {
		t.Fatalf("CaretReads = %d before any read", ctx.CaretReads)
	}
	ctx.CaretVisible() // CaretSince is zero: shown without reading the clock
	ctx.RestartCaret()
	ctx.CaretVisible()
	if ctx.CaretReads != 2 {
		t.Fatalf("CaretReads = %d after two reads, want 2", ctx.CaretReads)
	}
}

func TestAZeroCaretIsABlinkingBlock(t *testing.T) {
	ctx, err := NewCtx(testTheme(), testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Caret != (Caret{}) {
		t.Fatalf("Caret = %+v, want the zero value", ctx.Caret)
	}
	start := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	now := start
	ctx.Now = func() time.Time { return now }
	ctx.RestartCaret()
	steps := []struct {
		after time.Duration
		want  bool
	}{
		{0, true}, {499 * time.Millisecond, true},
		{500 * time.Millisecond, false}, {900 * time.Millisecond, false},
		{1000 * time.Millisecond, true}, {1500 * time.Millisecond, false},
	}
	for _, step := range steps {
		now = start.Add(step.after)
		if got := ctx.CaretVisible(); got != step.want {
			t.Errorf("after %v: visible = %v, want %v", step.after, got, step.want)
		}
	}
}

// A kept drawing redraws for the blink only when CaretReads moved while it
// was drawn, so a steady caret must not count.
func TestASteadyCaretCountsNoReads(t *testing.T) {
	ctx, err := NewCtx(testTheme(), testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	ctx.Caret.Steady = true
	start := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	now := start
	ctx.Now = func() time.Time { return now }
	ctx.CaretVisible()
	ctx.RestartCaret()
	for _, after := range []time.Duration{0, 500 * time.Millisecond, 900 * time.Millisecond, 1500 * time.Millisecond} {
		now = start.Add(after)
		if !ctx.CaretVisible() {
			t.Errorf("after %v: a steady caret is hidden", after)
		}
	}
	if ctx.CaretReads != 0 {
		t.Fatalf("CaretReads = %d for a steady caret, want 0", ctx.CaretReads)
	}
}
