package widget

import (
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// ToastMax is how many toasts are shown; older ones are dropped beyond it.
const ToastMax = 5

// ToastKind says how a toast behaves and looks: an info toast fades, an
// error toast stays until clicked. The zero value is ToastInfo.
type ToastKind uint8

// The kinds of toast.
const (
	ToastInfo ToastKind = iota
	ToastError
)

// Toast is a short message over the window.
type Toast struct {
	Text    string
	Kind    ToastKind
	Expires time.Time // zero: stays until clicked
}

// ToastStack is the toasts shown, oldest first.
type ToastStack []Toast

// Show adds a toast: an info one fades after feel's ToastTime (pass
// ctx.Feel(); the zero value is the built-in), fixed now so that a later theme
// does not change it, and an error stays
// until it is dismissed. It reports whether anything was added; an error
// with the text of one already shown is not, since the same failure on
// every frame or send is one message.
func (stack *ToastStack) Show(text string, kind ToastKind, now time.Time, feel core.Feel) bool {
	toast := Toast{Text: text, Kind: kind}
	if kind == ToastError {
		for _, shown := range *stack {
			if shown.Kind == ToastError && shown.Text == text {
				return false
			}
		}
	} else {
		toast.Expires = now.Add(feel.Resolved().ToastTime)
	}
	*stack = append(*stack, toast)
	if len(*stack) > ToastMax {
		*stack = (*stack)[len(*stack)-ToastMax:]
	}
	return true
}

// Prune drops the info toasts whose time is up and reports whether any went.
func (stack *ToastStack) Prune(now time.Time) bool {
	kept := (*stack)[:0]
	for _, toast := range *stack {
		if toast.Expires.IsZero() || now.Before(toast.Expires) {
			kept = append(kept, toast)
		}
	}
	removed := len(kept) != len(*stack)
	*stack = kept
	return removed
}

// Dismiss removes toast i.
func (stack *ToastStack) Dismiss(index int) { *stack = append((*stack)[:index], (*stack)[index+1:]...) }

// Toasts shows the stack in the bottom-right corner of a window of
// width by height, bottomInset above its lower edge (a status bar), oldest
// on top. A click dismisses the toast under it. Nil when there are none.
func Toasts(stack *ToastStack, width, height, bottomInset int) *El {
	if len(*stack) == 0 {
		return nil
	}
	count := len(*stack)
	return Col(WOf(core.ToastWOf), GapOf(core.ToastGapOf),
		FloatAtOf(
			func(sizes core.Sizes) int { return width - sizes.Gutter - sizes.ToastW },
			func(sizes core.Sizes) int {
				return height - bottomInset - sizes.ToastMargin - (count*sizes.ToastH + (count-1)*sizes.ToastGap)
			},
		),
		For(*stack, func(index int, toast Toast) Node {
			border, color := core.Border, core.Text
			if toast.Kind == ToastError {
				border, color = core.Error, core.Error
			}
			return Row(HOf(core.ToastHOf), PadXOf(core.GutterOf), Bg(core.Panel), Border(border), Shadow(), OnClick(func() { stack.Dismiss(index) }),
				Text(toast.Text, Ellipsis(), Color(color), Grow(1)),
			)
		}),
	)
}
