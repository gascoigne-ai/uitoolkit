package render

// Cursor is a mouse cursor shape; core.Cursor is an alias of it.
type Cursor uint8

// The cursor shapes.
const (
	CursorDefault    Cursor = iota
	CursorResizeH           // resize left/right, over a vertical divider
	CursorResizeV           // resize up/down, over a horizontal divider
	CursorResizeNWSE        // resize along a diagonal, over a top-left or bottom-right corner
	CursorResizeNESW        // resize along the other diagonal
	CursorMove              // something is held and being moved: a row, a tab
)

// Host is the program a renderer runs in a window. The renderer calls it from
// one goroutine.
type Host interface {
	// Layout takes the window size and the device scale factor and returns
	// the size of the canvas to draw on, in device pixels.
	Layout(width, height int, scale float64) (canvasW, canvasH int)
	// Update advances the program by one tick.
	Update() error
	// Draw paints the frame on screen and reports whether it drew. False
	// means there is nothing new: the renderer keeps the previous frame.
	Draw(screen Canvas) bool
	// Cursor is the mouse cursor shape to show.
	Cursor() Cursor
}
