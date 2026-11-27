package core

// CaretShape is how the text caret is drawn. A value that names no shape
// draws as CaretBlock.
type CaretShape uint8

// The caret shapes.
const (
	CaretBlock     CaretShape = iota // a solid cell over the character, which is redrawn on it
	CaretBar                         // a thin upright line before the character
	CaretUnderline                   // a thin line under the character
)

// Caret is how fields and text areas show where typing goes. The zero value
// is a blinking block. It is a choice of the app for all its widgets, held by
// Ctx.Caret: it is not part of a theme.
type Caret struct {
	Shape CaretShape
	// Steady keeps the caret shown. Otherwise it hides and shows every
	// Feel.CaretHalfPeriod of the theme from the last RestartCaret; until the first
	// RestartCaret it stays shown. Call RestartCaret when Input.RestartsCaret
	// and when Steady is turned off.
	Steady bool
}
