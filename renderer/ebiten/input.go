package ebitenrender

import (
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// mouseButtons maps ebiten's buttons to the render bits.
var mouseButtons = []struct {
	button ebiten.MouseButton
	bit    render.MouseButtons
}{
	{ebiten.MouseButtonLeft, render.MouseLeft},
	{ebiten.MouseButtonMiddle, render.MouseMiddle},
	{ebiten.MouseButtonRight, render.MouseRight},
}

// ReadInput samples the devices for this tick. Call it from Update. A key with
// no render name is left out.
func ReadInput() render.InputState {
	x, y := ebiten.CursorPosition()
	wheelX, wheelY := ebiten.Wheel()
	state := render.InputState{
		X: x, Y: y,
		WheelX: wheelX, WheelY: wheelY,
		Typed: ebiten.AppendInputChars(nil),
		TPS:   ebiten.TPS(),
	}
	for _, mouse := range mouseButtons {
		if ebiten.IsMouseButtonPressed(mouse.button) {
			state.Buttons |= mouse.bit
		}
		if inpututil.IsMouseButtonJustPressed(mouse.button) {
			state.ButtonsPressed |= mouse.bit
		}
	}
	for _, down := range inpututil.AppendPressedKeys(nil) {
		key := KeyFromEbiten(down)
		if key == render.KeyUnknown {
			continue
		}
		// A key pressed and released within one tick has no duration, but it
		// did go down. An aggregate has neither: ebiten does not time it.
		ticks := inpututil.KeyPressDuration(down)
		if inpututil.IsKeyJustPressed(down) {
			ticks = 1
		}
		state.Down = append(state.Down, render.KeyDown{Key: key, Ticks: ticks})
	}
	return state
}
