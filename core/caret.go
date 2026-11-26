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
// is a steady block. It is a choice of the app for all its widgets, held by
// Ctx.Caret: it is not part of a theme.
type Caret struct {
	Shape CaretShape
	// Blink hides and shows the caret every CaretHalfPeriod, counted from the
	// last RestartCaret. An app that never calls it gets a caret that does not
	// blink: call it when Input.RestartsCaret and when turning Blink on.
	Blink bool
}
