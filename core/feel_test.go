package core

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/render"
)

const ms = time.Millisecond

// feelCases is every value of Feel in theme-file units: its built-in and
// bounds, and how to read and write the field.
var feelCases = []struct {
	name              string
	builtin, min, max int
	get               func(Feel) int
	set               func(*Feel, int)
}{
	millisCase("caretHalfPeriod", 500, 200, 2000, func(f *Feel) *time.Duration { return &f.CaretHalfPeriod }),
	millisCase("keyRepeatDelay", 417, 150, 1000, func(f *Feel) *time.Duration { return &f.KeyRepeat.Delay }),
	millisCase("keyRepeatInterval", 50, 10, 1000, func(f *Feel) *time.Duration { return &f.KeyRepeat.Interval }),
	millisCase("holdDelay", 333, 100, 1000, func(f *Feel) *time.Duration { return &f.HoldRepeat.Delay }),
	millisCase("holdInterval", 50, 10, 1000, func(f *Feel) *time.Duration { return &f.HoldRepeat.Interval }),
	millisCase("multiClickTime", 400, 200, 1000, func(f *Feel) *time.Duration { return &f.MultiClickTime }),
	countCase("multiClickDistance", 4, 1, 16, func(f *Feel) *int { return &f.MultiClickDistance }),
	millisCase("tooltipDelay", 500, 100, 1000, func(f *Feel) *time.Duration { return &f.TooltipDelay }),
	millisCase("toastTime", 3000, 1000, 60000, func(f *Feel) *time.Duration { return &f.ToastTime }),
	countCase("panSlop", 4, 1, 32, func(f *Feel) *int { return &f.PanSlop }),
	countCase("wheelStep", 48, 8, 400, func(f *Feel) *int { return &f.WheelStep }),
	countCase("wheelLines", 3, 1, 20, func(f *Feel) *int { return &f.WheelLines }),
	countCase("wheelColumns", 6, 1, 40, func(f *Feel) *int { return &f.WheelColumns }),
}

type feelCase = struct {
	name              string
	builtin, min, max int
	get               func(Feel) int
	set               func(*Feel, int)
}

func millisCase(name string, builtin, low, high int, field func(*Feel) *time.Duration) feelCase {
	return feelCase{name, builtin, low, high,
		func(f Feel) int { return int(*field(&f) / ms) },
		func(f *Feel, value int) { *field(f) = time.Duration(value) * ms }}
}

func countCase(name string, builtin, low, high int, field func(*Feel) *int) feelCase {
	return feelCase{name, builtin, low, high,
		func(f Feel) int { return *field(&f) },
		func(f *Feel, value int) { *field(f) = value }}
}

func TestFeelResolvedFillsBuiltInsAndClamps(t *testing.T) {
	for _, row := range feelCases {
		for _, test := range []struct{ given, want int }{
			{0, row.builtin}, {-1, row.min}, {row.min - 1, row.min}, {row.min, row.min},
			{row.max, row.max}, {row.max + 1, row.max}, {math.MaxInt32, row.max},
		} {
			var feel Feel
			row.set(&feel, test.given)
			want := test.want
			if row.min-1 == 0 && test.given == 0 {
				want = row.builtin // 0 is unset, whatever the minimum
			}
			if got := row.get(feel.Resolved()); got != want {
				t.Errorf("%s %d resolves to %d, want %d", row.name, test.given, got, want)
			}
		}
	}
}

// A duration set in Go below a millisecond is stated, not unset.
func TestFeelResolvedClampsTinyAndExtremeDurations(t *testing.T) {
	tiny := Feel{CaretHalfPeriod: 1, ToastTime: math.MinInt64, TooltipDelay: math.MaxInt64}.Resolved()
	if tiny.CaretHalfPeriod != 200*ms || tiny.ToastTime != time.Second || tiny.TooltipDelay != time.Second {
		t.Fatalf("resolved %v %v %v", tiny.CaretHalfPeriod, tiny.ToastTime, tiny.TooltipDelay)
	}
}

func TestFeelBuiltInsKeepTodaysTicksAt60TPS(t *testing.T) {
	feel := Feel{}.Resolved()
	for _, test := range []struct {
		name     string
		duration time.Duration
		want     int
	}{
		{"key delay", feel.KeyRepeat.Delay, 25}, {"key interval", feel.KeyRepeat.Interval, 3},
		{"hold delay", feel.HoldRepeat.Delay, 20}, {"hold interval", feel.HoldRepeat.Interval, 3},
		{"tooltip", feel.TooltipDelay, 30},
	} {
		for _, tps := range []int{60, 0, -1} {
			if got := (Input{TPS: tps}).Ticks(test.duration); got != test.want {
				t.Errorf("%s at %d TPS: %d ticks, want %d", test.name, tps, got, test.want)
			}
		}
	}
	if got := (Input{TPS: 30}).Ticks(500 * ms); got != 15 {
		t.Errorf("500ms at 30 TPS: %d ticks, want 15", got)
	}
}

func TestTicksSaturatesAndNeverGoesNegative(t *testing.T) {
	for _, in := range []Input{{TPS: math.MaxInt}, {TPS: 60}} {
		if got := in.Ticks(math.MaxInt64); got <= 0 {
			t.Errorf("TPS %d: the longest duration is %d ticks", in.TPS, got)
		}
		if got := in.Ticks(-time.Second); got != 0 {
			t.Errorf("TPS %d: a negative duration is %d ticks, want 0", in.TPS, got)
		}
	}
}

func TestFeelSetRejectsZeroNegativeHugeAndUnknown(t *testing.T) {
	for _, row := range feelCases {
		for _, value := range []int{0, -1, row.min - 1, row.max + 1, math.MaxInt32, math.MaxInt, math.MinInt} {
			if got, err := (Feel{}).Set(row.name, value); err == nil || got != (Feel{}) {
				t.Errorf("%s = %d: got %+v, %v, want an error and no change", row.name, value, got, err)
			}
		}
		for _, value := range []int{row.min, row.max} {
			var original Feel
			got, err := original.Set(row.name, value)
			if err != nil || row.get(got) != value || row.get(got.Resolved()) != value {
				t.Errorf("%s = %d: got %d, %v", row.name, value, row.get(got), err)
			}
			if original != (Feel{}) {
				t.Errorf("%s: Set changed its receiver", row.name)
			}
		}
	}
	for _, name := range []string{"", "WheelStep", "wheelstep", "caretSteady"} {
		if _, err := (Feel{}).Set(name, 10); err == nil {
			t.Errorf("unknown name %q was accepted", name)
		}
	}
}

func TestCaretBlinksByTheThemesHalfPeriod(t *testing.T) {
	theme := testTheme()
	theme.Feel.CaretHalfPeriod = 800 * ms
	ctx, err := NewCtx(theme, testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(100, 0)
	now := start
	ctx.Now = func() time.Time { return now }
	ctx.RestartCaret()
	for after, want := range map[time.Duration]bool{500 * ms: true, 799 * ms: true, 800 * ms: false, 1599 * ms: false, 1600 * ms: true} {
		now = start.Add(after)
		if got := ctx.CaretVisible(); got != want {
			t.Errorf("after %v: visible = %v, want %v", after, got, want)
		}
	}
	zero := &Ctx{CaretSince: start, Now: func() time.Time { return start.Add(500 * ms) }}
	if zero.CaretVisible() {
		t.Error("a zero Ctx blinks by the built-in 500ms")
	}
}

func TestCtxFeelFollowsSetTheme(t *testing.T) {
	ctx, err := NewCtx(testTheme(), testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Feel() != (Feel{}).Resolved() || ctx.Feel().ToastTime != 3*time.Second {
		t.Fatalf("a theme with no feel gives the built-ins, got %+v", ctx.Feel())
	}
	theme := testTheme()
	theme.Feel.ToastTime = 9 * time.Second
	if err := ctx.SetTheme(theme); err != nil {
		t.Fatal(err)
	}
	if ctx.Feel().ToastTime != 9*time.Second {
		t.Fatalf("ToastTime = %v after SetTheme", ctx.Feel().ToastTime)
	}
	ctx.Th.Feel.ToastTime = time.Hour // a plain assignment is followed too, and clamped
	if ctx.Feel().ToastTime != time.Minute {
		t.Fatalf("ToastTime = %v, want the bound", ctx.Feel().ToastTime)
	}
}

func TestThemeCopiesKeepFeel(t *testing.T) {
	theme := testTheme()
	theme.Feel.WheelStep = 100
	for name, copied := range map[string]Theme{"With": theme.With(nil), "WithTextSize": theme.WithTextSize(20, 18, 19)} {
		if copied.Feel != theme.Feel {
			t.Errorf("%s lost Feel: %+v", name, copied.Feel)
		}
	}
}

func TestSamplerCountsClicksByFeel(t *testing.T) {
	feel := Feel{MultiClickTime: 800 * ms, MultiClickDistance: 10}
	for _, test := range []struct {
		name  string
		after time.Duration
		x     int
		want  int
	}{
		{"within the themed window", 800 * ms, 20, 2},
		{"too late", 801 * ms, 10, 1},
		{"too far", 100 * ms, 21, 1},
	} {
		now := time.Unix(100, 0)
		sampler := &Sampler{Now: func() time.Time { return now }, HostOS: "linux"}
		click := render.InputState{X: 10, Y: 10, Buttons: render.MouseLeft, ButtonsPressed: render.MouseLeft}
		sampler.Sample(click, feel)
		now, click.X = now.Add(test.after), test.x
		if got := sampler.Sample(click, feel).Clicks; got != test.want {
			t.Errorf("%s: Clicks = %d, want %d", test.name, got, test.want)
		}
	}
	// Out-of-range values of a Go-built theme are clamped.
	var sampler Sampler
	start := time.Unix(100, 0)
	wild := Feel{MultiClickTime: time.Hour, MultiClickDistance: 1000}
	sampler.Click(start, 0, 0, wild)
	if got := sampler.Click(start.Add(1001*ms), 0, 0, wild); got != 1 {
		t.Errorf("after 1001ms: count = %d, want the time clamped to 1s", got)
	}
	if got := sampler.Click(start.Add(1002*ms), 17, 0, wild); got != 1 {
		t.Errorf("17px off: count = %d, want the distance clamped to 16", got)
	}
}

// firedAt are the ticks of the first n, from 1, at which a held editing key
// acts as the sampler reads it.
func firedAt(sampler *Sampler, feel Feel, tps, n int) []int {
	var ticks []int
	for held := 1; held <= n; held++ {
		state := render.InputState{TPS: tps, Down: []render.KeyDown{{Key: render.KeyDelete, Ticks: held}}}
		if len(sampler.Sample(state, feel).Edits) > 0 {
			ticks = append(ticks, held)
		}
	}
	return ticks
}

func TestSamplerKeyRepeatThemeDelay(t *testing.T) {
	feel, err := Feel{}.Set("keyRepeatDelay", 600)
	if err != nil {
		t.Fatal(err)
	}
	got := firedAt(&Sampler{HostOS: "linux"}, feel, 60, 42)
	if want := []int{1, 36, 39, 42}; !slices.Equal(got, want) {
		t.Fatalf("fired at %v, want %v", got, want)
	}
}

func TestSamplerKeyRepeatThemeInterval(t *testing.T) {
	feel, err := Feel{}.Set("keyRepeatInterval", 200)
	if err != nil {
		t.Fatal(err)
	}
	got := firedAt(&Sampler{HostOS: "linux"}, feel, 60, 49)
	if want := []int{1, 25, 37, 49}; !slices.Equal(got, want) {
		t.Fatalf("fired at %v, want %v", got, want)
	}
}

func TestSamplerKeyRepeatFeelAndSystem(t *testing.T) {
	theme := Feel{KeyRepeat: Repeat{Delay: 200 * ms, Interval: 100 * ms}} // ticks 12, 6
	themed := []int{1, 12, 18, 24}
	for _, test := range []struct {
		name   string
		system Repeat
		want   []int
	}{
		{"unknown: the theme's", Repeat{}, themed},
		{"known: the desktop's wins", Repeat{Delay: 150 * ms, Interval: 150 * ms}, []int{1, 9, 18}},
		{"a delay alone: the theme's, both", Repeat{Delay: 500 * ms}, themed},
		{"an interval alone: the theme's, both", Repeat{Interval: 50 * ms}, themed},
		{"a negative delay: the theme's", Repeat{Delay: -1, Interval: 50 * ms}, themed},
		{"a negative interval: the theme's", Repeat{Delay: 50 * ms, Interval: -time.Second}, themed},
	} {
		got := firedAt(&Sampler{HostOS: "linux", Repeat: test.system}, theme, 60, 24)
		if !slices.Equal(got, test.want) {
			t.Errorf("%s: fired at %v, want %v", test.name, got, test.want)
		}
	}
	// Known but under a tick: every tick once, never more.
	sampler := &Sampler{HostOS: "linux", Repeat: Repeat{Delay: 1, Interval: 1}}
	if got := firedAt(sampler, theme, 60, 4); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Errorf("1ns repeat fired at %v", got)
	}
	state := render.InputState{TPS: 60, Down: []render.KeyDown{{Key: render.KeyDelete, Ticks: 3}}}
	if got := sampler.Sample(state, theme).Edits; len(got) != 1 {
		t.Errorf("one tick gave %d edits", len(got))
	}
}

func TestRepeatStaysBoundedOnAnyInput(t *testing.T) {
	repeats := []Repeat{{}, {Delay: -1, Interval: -1}, {Delay: time.Second}, {Interval: time.Second},
		{Delay: 1, Interval: 1}, {Delay: math.MaxInt64, Interval: math.MaxInt64}}
	for _, repeat := range repeats {
		for _, tps := range []int{math.MinInt, -1, 0, 1, 60, math.MaxInt} {
			for _, held := range []int{0, 1, 2, 25, math.MaxInt} {
				repeat.Fires(held, tps) // must not panic
				if repeat.Due(held, tps) && held <= 0 {
					t.Errorf("%+v is due for a key that is not down", repeat)
				}
			}
		}
	}
	longest := Repeat{Delay: math.MaxInt64, Interval: math.MaxInt64}
	if longest.Due(25, math.MaxInt) || longest.Fires(25, 60) {
		t.Error("the longest repeat must not fire early")
	}
}

func TestRepeatDueSkipsThePress(t *testing.T) {
	repeat := Repeat{Delay: 100 * ms, Interval: 50 * ms} // ticks 6, 3
	var got []int
	for held := range 13 {
		if repeat.Due(held, 60) {
			got = append(got, held)
		}
	}
	if want := []int{6, 9, 12}; !slices.Equal(got, want) {
		t.Fatalf("due at %v, want %v", got, want)
	}
	if !(Repeat{}).Due(25, 60) || (Repeat{}).Due(1, 60) {
		t.Error("an unknown repeat is the built-in key repeat")
	}
	if !repeat.Fires(1, 60) {
		t.Error("Fires includes the press")
	}
}

func TestSamplePassesTheTickRateOn(t *testing.T) {
	var sampler Sampler
	if got := sampler.Sample(render.InputState{TPS: 30}, Feel{}).Scaled(2); got.TPS != 30 {
		t.Fatalf("TPS = %d, want 30", got.TPS)
	}
}
