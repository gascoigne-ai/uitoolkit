package core

import (
	"fmt"
	"time"
)

// Feel are the timings and pointer distances a theme sets. The zero value of
// a field is its built-in value, so a Theme that says nothing about them
// behaves as the built-in themes do. A theme file states them in its "feel"
// block (see theming.ParseTheme), durations in whole milliseconds, within the
// bounds given at each field. The bounds keep a theme, which may be
// untrusted, from freezing or spinning the interface. Safety limits, such as
// how many toasts are kept, are not part of it.
//
// Widgets read the values through Ctx.Feel, so a new theme applies from the
// next tick. The values are counted in the ticks of Input.TPS or timed by a
// clock, as each field's reader needs.
type Feel struct {
	// CaretHalfPeriod is how long a blinking caret is shown, and then
	// hidden: 200ms to 2s, built in 500ms. Caret.Steady turns the blink off
	// whatever it says.
	CaretHalfPeriod time.Duration

	// KeyRepeat is how a held editing key repeats when the system's own
	// timing is not known: the desktop's, in Sampler.Repeat, wins, since
	// users set it for their own reach. Delay is 150ms to 1s, built in
	// 417ms (25 ticks at 60 TPS); Interval is 10ms to 1s, built in 50ms.
	KeyRepeat Repeat

	// HoldRepeat is how a held button, such as an arrow of an HScroll,
	// repeats. Delay is 100ms to 1s, built in 333ms (20 ticks at 60 TPS);
	// Interval is 10ms to 1s, built in 50ms.
	HoldRepeat Repeat

	// MultiClickTime is the longest pause between the clicks of a double or
	// triple click: 200ms to 1s, built in 400ms. A long window would turn a
	// quick second drag into a double click, so it is bounded tightly.
	MultiClickTime time.Duration
	// MultiClickDistance is how far apart, on each axis, those clicks may
	// be, in device pixels, since the Sampler reads the pointer before it is
	// scaled: 1 to 16, built in 4.
	MultiClickDistance int

	// TooltipDelay is how long the pointer rests on an element before its
	// tooltip shows: 100ms to 1s, built in 500ms.
	TooltipDelay time.Duration
	// ToastTime is how long an info toast stays: 1s to 60s, built in 3s. A
	// toast keeps the time it was shown with.
	ToastTime time.Duration

	// PanSlop is how far, in logical pixels, a pressed pointer moves before
	// it drags an HScroll: 1 to 32, built in 4.
	PanSlop int
	// WheelStep is how many logical pixels one wheel notch scrolls a Scroll
	// or an HScroll: 8 to 400, built in 48.
	WheelStep int
	// WheelLines is how many rows one wheel notch scrolls a CodeView or a
	// TextArea: 1 to 20, built in 3.
	WheelLines int
	// WheelColumns is how many characters one wheel notch scrolls them
	// sideways: 1 to 40, built in 6.
	WheelColumns int
}

// feelRule is one value of Feel: its name in theme files, its bounds in the
// file's units (milliseconds, pixels or a count) and how it is written.
type feelRule struct {
	name      string
	low, high int
	set       func(*Feel, int)
	resolve   func(*Feel) // the built-in for zero, else within the bounds
}

// millisRule is a duration, stated in theme files in milliseconds.
func millisRule(name string, builtin, low, high int, field func(*Feel) *time.Duration) feelRule {
	span := func(millis int) time.Duration { return time.Duration(millis) * time.Millisecond }
	return feelRule{name, low, high,
		func(feel *Feel, value int) { *field(feel) = span(value) },
		func(feel *Feel) { resolveFeel(field(feel), span(builtin), span(low), span(high)) }}
}

// countRule is a number of pixels, rows or characters.
func countRule(name string, builtin, low, high int, field func(*Feel) *int) feelRule {
	return feelRule{name, low, high,
		func(feel *Feel, value int) { *field(feel) = value },
		func(feel *Feel) { resolveFeel(field(feel), builtin, low, high) }}
}

func resolveFeel[T int | time.Duration](value *T, builtin, low, high T) {
	if *value == 0 {
		*value = builtin
		return
	}
	*value = min(max(*value, low), high)
}

// feelRules is every value of Feel, in the order of its fields.
var feelRules = [...]feelRule{
	millisRule("caretHalfPeriod", 500, 200, 2000, func(f *Feel) *time.Duration { return &f.CaretHalfPeriod }),
	millisRule("keyRepeatDelay", 417, 150, 1000, func(f *Feel) *time.Duration { return &f.KeyRepeat.Delay }),
	millisRule("keyRepeatInterval", 50, 10, 1000, func(f *Feel) *time.Duration { return &f.KeyRepeat.Interval }),
	millisRule("holdDelay", 333, 100, 1000, func(f *Feel) *time.Duration { return &f.HoldRepeat.Delay }),
	millisRule("holdInterval", 50, 10, 1000, func(f *Feel) *time.Duration { return &f.HoldRepeat.Interval }),
	millisRule("multiClickTime", 400, 200, 1000, func(f *Feel) *time.Duration { return &f.MultiClickTime }),
	countRule("multiClickDistance", 4, 1, 16, func(f *Feel) *int { return &f.MultiClickDistance }),
	millisRule("tooltipDelay", 500, 100, 1000, func(f *Feel) *time.Duration { return &f.TooltipDelay }),
	millisRule("toastTime", 3000, 1000, 60000, func(f *Feel) *time.Duration { return &f.ToastTime }),
	countRule("panSlop", 4, 1, 32, func(f *Feel) *int { return &f.PanSlop }),
	countRule("wheelStep", 48, 8, 400, func(f *Feel) *int { return &f.WheelStep }),
	countRule("wheelLines", 3, 1, 20, func(f *Feel) *int { return &f.WheelLines }),
	countRule("wheelColumns", 6, 1, 40, func(f *Feel) *int { return &f.WheelColumns }),
}

// Resolved is feel ready to use: a zero field is its built-in value and any
// other is brought into its bounds.
func (feel Feel) Resolved() Feel {
	for _, rule := range feelRules {
		rule.resolve(&feel)
	}
	return feel
}

// Set is a copy of feel with the named value set, in the units of a theme
// file: milliseconds for a duration, else pixels or a count. The names are
// the keys of the "feel" block, such as "tooltipDelay". An unknown name or a
// value outside the bounds, zero and negative ones among them, is an error and
// leaves the copy as feel.
func (feel Feel) Set(name string, value int) (Feel, error) {
	for _, rule := range feelRules {
		if rule.name != name {
			continue
		}
		if value < rule.low || value > rule.high {
			return feel, fmt.Errorf("%q must be within %d and %d", name, rule.low, rule.high)
		}
		rule.set(&feel, value)
		return feel, nil
	}
	return feel, fmt.Errorf("unknown value %q", name)
}
