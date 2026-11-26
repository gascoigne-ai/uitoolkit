package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/gascoigne-ai/uitoolkit/render"
	ebitenrender "github.com/gascoigne-ai/uitoolkit/renderer/ebiten"
)

// The window the showcase opens, in device-independent pixels. The goldens
// are drawn at the same size.
const (
	windowW     = 1024
	windowH     = 720
	windowTitle = "Toolkit showcase"
)

// backend is what one renderer gives the showcase: the device that makes its
// fonts and images, the input of a tick, the window loop, and how far its
// pixels may differ from a golden.
type backend struct {
	Device    render.Device
	ReadInput func() render.InputState
	Run       func(host render.Host) error
	Tolerance tolerance
}

// tolerance is what a golden comparison forgives: at most MaxPixels pixels
// may differ, each channel by at most MaxDelta.
type tolerance struct {
	MaxDelta  uint8
	MaxPixels int
}

// rendererRow is one choice of the -renderer flag.
type rendererRow struct {
	Name string
	Open func() (backend, error)
}

// renderers are the choices of the -renderer flag, in the order the errors
// list them. A new renderer is a new row.
var renderers = []rendererRow{
	{Name: "ebiten", Open: openEbiten},
	{Name: "wgpu", Open: openWgpu},
}

func openRenderer(name string) (backend, error) {
	names := make([]string, len(renderers))
	for i, row := range renderers {
		if row.Name == name {
			return row.Open()
		}
		names[i] = row.Name
	}
	return backend{}, fmt.Errorf("unknown renderer %q: the renderers are %s", name, strings.Join(names, ", "))
}

func openEbiten() (backend, error) {
	return backend{
		Device:    ebitenrender.NewDevice(),
		ReadInput: ebitenrender.ReadInput,
		Run: func(host render.Host) error {
			ebiten.SetWindowSize(windowW, windowH)
			ebiten.SetWindowTitle(windowTitle)
			ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
			return ebitenrender.Run(host, nil)
		},
	}, nil
}

// openWgpu says the wgpu renderer is not there yet; task 016 replaces it
// with a build behind the wgpu tag.
func openWgpu() (backend, error) {
	return backend{}, errors.New(`renderer "wgpu" not built: build with -tags wgpu (task 016)`)
}
