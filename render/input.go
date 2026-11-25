package render

// MouseButtons is a set of mouse buttons.
type MouseButtons uint8

// The mouse buttons.
const (
	MouseLeft MouseButtons = 1 << iota
	MouseMiddle
	MouseRight
)

// InputState is what a window reports of its devices in one tick: states and
// durations only, so that what they mean (clicks, edits, shortcuts) is decided
// once, outside the renderer.
type InputState struct {
	// X and Y are the pointer position in canvas pixels.
	X, Y           int
	WheelX, WheelY float64
	// Typed are the characters entered in this tick.
	Typed []rune
	// Buttons are held now; ButtonsPressed went down in this tick.
	Buttons, ButtonsPressed MouseButtons
	// Down are the keys that are down, in Key order, the aggregates
	// (KeyControl...) among them.
	Down []KeyDown
	// TPS is the number of ticks a second.
	TPS int
}

// KeyDown is a key that is down. Ticks is how many ticks it has been down for:
// 1 on the tick it goes down, which is how "just pressed" is told, even for a
// key released again in that tick. An aggregate stands for either of two keys
// that may have gone down at different ticks, so its Ticks is 0: it is never
// just pressed.
type KeyDown struct {
	Key   Key
	Ticks int
}
