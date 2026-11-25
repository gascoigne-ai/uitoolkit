package ebitenrender

import (
	"runtime/debug"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/hajimehoshi/ebiten/v2"
)

const ebitenModule = "github.com/hajimehoshi/ebiten/v2"

// ReadGraphics is the graphics side of the running game: call it only while
// the run loop runs, where the graphics library is known. It is the
// `Graphics` source of systeminfo.Info.
func ReadGraphics() render.Graphics {
	var debugInfo ebiten.DebugInfo
	ebiten.ReadDebugInfo(&debugInfo)
	monitor := ebiten.Monitor()
	g := render.Graphics{
		Renderer:        "Ebitengine",
		RendererVersion: moduleVersion(ebitenModule),
		Library:         debugInfo.GraphicsLibrary.String(),
		GPUImageBytes:   debugInfo.TotalGPUImageMemoryUsageInBytes,
		Vsync:           ebiten.IsVsyncEnabled(),
		FPS:             ebiten.ActualFPS(),
		TPS:             ebiten.ActualTPS(),
	}
	if monitor != nil {
		g.Monitor, g.DeviceScale = monitor.Name(), monitor.DeviceScaleFactor()
		g.MonitorW, g.MonitorH = monitor.Size()
	}
	g.WindowW, g.WindowH = ebiten.WindowSize()
	return g
}

// moduleVersion is the version of a module the program was built with.
func moduleVersion(path string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, dep := range info.Deps {
		if dep.Path == path {
			return dep.Version
		}
	}
	return "unknown"
}
