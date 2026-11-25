package render

// Graphics is what a running renderer, its window system and GPU tell, for the
// system report. A renderer fills it from the library it draws with; the
// toolkit only shows it.
type Graphics struct {
	Renderer        string // the renderer's name, "Ebitengine"
	RendererVersion string // the version of the library behind it, "v2.10.5"
	Library         string // the graphics API: "Metal", "OpenGL", "DirectX"
	GPUImageBytes   int64  // memory taken by images on the GPU
	Vsync           bool
	FPS, TPS        float64 // actual frames and ticks per second
	Monitor         string
	MonitorW        int
	MonitorH        int
	DeviceScale     float64
	WindowW         int
	WindowH         int
}
