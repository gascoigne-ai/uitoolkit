package ebitenrender

import (
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// Run runs host in a window until it returns an error or the window closes. It
// draws with its own device and turns off ebiten's clearing of the screen
// before each frame, so a frame the host does not draw keeps the previous one.
// A nil options takes ebiten's defaults.
//
// A program with its own ebiten.Game uses Device.NewCanvas instead.
func Run(host render.Host, options *ebiten.RunGameOptions) error {
	ebiten.SetScreenClearedEveryFrame(false)
	return ebiten.RunGameWithOptions(&game{host: host, screen: NewDevice().NewCanvas(nil)}, options)
}

// game is a render.Host as an ebiten.Game. screen is one canvas, pointed at
// the screen ebiten hands over in each frame.
type game struct {
	host   render.Host
	screen *Canvas
}

var _ ebiten.Game = (*game)(nil)

// Layout implements ebiten.Game.
func (g *game) Layout(width, height int) (int, int) {
	return g.host.Layout(width, height, ebiten.Monitor().DeviceScaleFactor())
}

// Update implements ebiten.Game.
func (g *game) Update() error {
	if err := g.host.Update(); err != nil {
		return err
	}
	ebiten.SetCursorShape(cursorShapes[g.host.Cursor()])
	return nil
}

// Draw implements ebiten.Game. What the host reports is not needed: the screen
// is not cleared, so a frame it skips keeps the previous one.
func (g *game) Draw(screen *ebiten.Image) {
	g.screen.dst = screen
	g.host.Draw(g.screen)
}
