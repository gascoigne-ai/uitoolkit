package ebitenrender

import (
	"errors"
	"image"
	"image/color"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// fakeHost records what the game asks of its host.
type fakeHost struct {
	width, height int
	scale         float64
	updates       int
	updateErr     error
	cursor        render.Cursor
	draw          func(screen render.Canvas)
}

func (h *fakeHost) Layout(width, height int, scale float64) (int, int) {
	h.width, h.height, h.scale = width, height, scale
	return 2 * width, 3 * height
}

func (h *fakeHost) Update() error {
	h.updates++
	return h.updateErr
}

func (h *fakeHost) Draw(screen render.Canvas) bool {
	if h.draw != nil {
		h.draw(screen)
	}
	return h.draw != nil
}

func (h *fakeHost) Cursor() render.Cursor { return h.cursor }

func TestGameLayoutPassesTheDeviceScale(t *testing.T) {
	host := &fakeHost{}
	loop := &game{host: host, screen: NewDevice().NewCanvas(nil)}
	width, height := loop.Layout(100, 50)
	if width != 200 || height != 150 {
		t.Errorf("Layout = %d, %d, want what the host returns: 200, 150", width, height)
	}
	want := ebiten.Monitor().DeviceScaleFactor()
	if host.width != 100 || host.height != 50 || host.scale != want {
		t.Errorf("host got %d, %d, %v, want 100, 50, %v", host.width, host.height, host.scale, want)
	}
}

func TestGameDrawHandsTheScreenAsCanvas(t *testing.T) {
	red := color.RGBA{255, 0, 0, 255}
	var bounds image.Rectangle
	host := &fakeHost{draw: func(screen render.Canvas) {
		bounds = screen.Bounds()
		screen.FillRect(render.Box{X0: 1, Y0: 1, X1: 3, Y1: 3}, red)
	}}
	loop := &game{host: host, screen: NewDevice().NewCanvas(nil)}
	for _, size := range []int{8, 12} {
		screen := ebiten.NewImage(size, size)
		loop.Draw(screen)
		if bounds != screen.Bounds() {
			t.Errorf("canvas bounds = %v, want %v", bounds, screen.Bounds())
		}
		if got := screen.At(2, 2); got != red {
			t.Errorf("pixel of the screen = %v, want %v", got, red)
		}
	}
}

func TestGameUpdateReturnsTheHostError(t *testing.T) {
	broken := errors.New("broken")
	tests := []struct {
		name  string
		err   error
		shape ebiten.CursorShapeType
	}{
		{"the cursor follows the host", nil, ebiten.CursorShapeEWResize},
		{"an error is returned before the cursor is read", broken, ebiten.CursorShapeDefault},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ebiten.SetCursorShape(ebiten.CursorShapeDefault)
			defer ebiten.SetCursorShape(ebiten.CursorShapeDefault)
			host := &fakeHost{updateErr: testCase.err, cursor: render.CursorResizeH}
			loop := &game{host: host, screen: NewDevice().NewCanvas(nil)}
			if err := loop.Update(); !errors.Is(err, testCase.err) {
				t.Errorf("Update = %v, want %v", err, testCase.err)
			}
			if host.updates != 1 {
				t.Errorf("host updated %d times, want 1", host.updates)
			}
			if got := ebiten.CursorShape(); got != testCase.shape {
				t.Errorf("cursor shape = %v, want %v", got, testCase.shape)
			}
		})
	}
}
