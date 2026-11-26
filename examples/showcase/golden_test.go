package main

import (
	"errors"
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/examples/showcase/sections"
	"github.com/gascoigne-ai/uitoolkit/render"
)

var update = flag.Bool("update", false, "rewrite the golden screenshots")

// goldenTicks are the ticks run after a pose: one more than the tooltip
// waits, which counts from the first frame that sees the pointer rest.
const goldenTicks = 31

// shot is one golden screenshot.
type shot struct {
	name, section   string
	pose            sections.Pose
	theme, contrast string
	device          float64 // the display's scale factor
}

// goldenShots is every section, in each of its poses, in the dark theme at
// scale 1, then the theming section in the other looks and the typography
// section on a display of scale 2.
func goldenShots(t *testing.T) []shot {
	var shots []shot
	for _, section := range sections.Sections {
		poses := section.Poses
		if len(poses) == 0 {
			poses = []sections.Pose{{}}
		}
		for _, pose := range poses {
			name := section.ID
			if pose.Name != "" {
				name += "-" + pose.Name
			}
			shots = append(shots, shot{name, section.ID, pose, "dark", "normal", 1})
		}
	}
	if len(shots) == 0 {
		t.Fatal("no sections to shoot")
	}
	return append(shots,
		shot{"theming-light", "theming", sections.Pose{}, "light", "normal", 1},
		shot{"theming-high-contrast", "theming", sections.Pose{}, "dark", "high", 1},
		shot{"typography-2x", "typography", sections.Pose{}, "dark", "normal", 2},
	)
}

type memoryClipboard struct{ text string }

func (c *memoryClipboard) Read() (string, error)   { return c.text, nil }
func (c *memoryClipboard) Write(text string) error { c.text = text; return nil }

// render draws one shot with everything that moves pinned: the clocks, the
// platform, the clipboard, the pointer and the number of ticks.
func (taken shot) render(t *testing.T, opened backend) *image.RGBA {
	t.Helper()
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	clock := func() time.Time { return at }
	showcase := sections.New(clock)
	showcase.SetThemeID(taken.theme)
	showcase.SetContrast(taken.contrast)
	ctx, err := core.NewCtx(showcase.Theme(), opened.Device)
	if err != nil {
		t.Fatal(err)
	}
	ctx.Now, ctx.HostOS, ctx.Clipboard = clock, "linux", &memoryClipboard{}
	pointer := sections.OffCanvas
	window := newHost(showcase, ctx, func() render.InputState {
		return render.InputState{X: int(float64(pointer.X) * ctx.Scale), Y: int(float64(pointer.Y) * ctx.Scale), TPS: 60}
	})
	window.sampler = core.Sampler{Now: clock, HostOS: "linux"}
	index := -1
	for i, section := range sections.Sections {
		if section.ID == taken.section {
			index = i
		}
	}
	if index < 0 {
		t.Fatalf("no section %q", taken.section)
	}
	showcase.Select(ctx, index)
	width, height := window.Layout(windowW, windowH, taken.device)
	screen := opened.Device.NewOffscreen(width, height)
	defer screen.Release()
	tick := func() {
		if err := window.Update(); err != nil {
			t.Fatal(err)
		}
		screen.Clear()
		if !window.Draw(screen) {
			t.Fatal("the host drew nothing")
		}
	}
	tick()
	if taken.pose.Apply != nil {
		pointer = taken.pose.Apply(showcase, ctx)
	}
	for range goldenTicks {
		tick()
	}
	return screen.ReadPixels()
}

// difference is how two pictures of one size differ.
type difference struct {
	pixels int
	worst  uint8
	first  image.Point
	diff   *image.RGBA // the differing pixels in red on black
}

func compare(got, want *image.RGBA) difference {
	found := difference{diff: image.NewRGBA(got.Bounds())}
	for y := got.Bounds().Min.Y; y < got.Bounds().Max.Y; y++ {
		for x := got.Bounds().Min.X; x < got.Bounds().Max.X; x++ {
			a, b := got.RGBAAt(x, y), want.RGBAAt(x, y)
			delta := max(gap(a.R, b.R), gap(a.G, b.G), gap(a.B, b.B), gap(a.A, b.A))
			found.diff.SetRGBA(x, y, color.RGBA{A: 255})
			if delta == 0 {
				continue
			}
			if found.pixels == 0 {
				found.first = image.Pt(x, y)
			}
			found.pixels++
			found.worst = max(found.worst, delta)
			found.diff.SetRGBA(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	return found
}

func gap(a, b uint8) uint8 { return max(a, b) - min(a, b) }

func (d difference) within(allowed tolerance) bool {
	return d.pixels == 0 || (d.pixels <= allowed.MaxPixels && d.worst <= allowed.MaxDelta)
}

func TestCompareImages(t *testing.T) {
	want := image.NewRGBA(image.Rect(0, 0, 3, 2))
	got := image.NewRGBA(want.Bounds())
	got.SetRGBA(2, 0, color.RGBA{G: 3, A: 0})
	got.SetRGBA(1, 1, color.RGBA{B: 9})
	found := compare(got, want)
	if found.pixels != 2 || found.worst != 9 || found.first != image.Pt(2, 0) {
		t.Fatalf("difference = %d pixels, worst %d, first %v", found.pixels, found.worst, found.first)
	}
	tests := []struct {
		name    string
		allowed tolerance
		want    bool
	}{
		{"exact", tolerance{}, false},
		{"enough of both", tolerance{MaxDelta: 9, MaxPixels: 2}, true},
		{"too small a delta", tolerance{MaxDelta: 8, MaxPixels: 2}, false},
		{"too few pixels", tolerance{MaxDelta: 9, MaxPixels: 1}, false},
	}
	for _, test := range tests {
		if got := found.within(test.allowed); got != test.want {
			t.Errorf("%s: within = %v, want %v", test.name, got, test.want)
		}
	}
}

func readPNG(path string) (*image.RGBA, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoded, err := png.Decode(file)
	if err != nil {
		return nil, err
	}
	picture := image.NewRGBA(decoded.Bounds())
	draw.Draw(picture, picture.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
	return picture, nil
}

func writePNG(t *testing.T, path string, picture image.Image) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := errors.Join(png.Encode(file, picture), file.Close()); err != nil {
		t.Fatal(err)
	}
}

func TestGolden(t *testing.T) {
	opened, err := openRenderer("ebiten")
	if err != nil {
		t.Fatal(err)
	}
	for _, taken := range goldenShots(t) {
		t.Run(taken.name, func(t *testing.T) {
			got := taken.render(t, opened)
			path := filepath.Join("testdata", "golden", taken.name+".png")
			if *update {
				writePNG(t, path, got)
			}
			want, err := readPNG(path)
			if errors.Is(err, fs.ErrNotExist) {
				writePNG(t, filepath.Join("testdata", "failed", taken.name+".got.png"), got)
				t.Fatalf("no golden %s (rerun with -update)", path)
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Bounds() != want.Bounds() {
				t.Fatalf("size %v, golden %v", got.Bounds(), want.Bounds())
			}
			found := compare(got, want)
			if found.within(opened.Tolerance) {
				return
			}
			gotPath := filepath.Join("testdata", "failed", taken.name+".got.png")
			diffPath := filepath.Join("testdata", "failed", taken.name+".diff.png")
			writePNG(t, gotPath, got)
			writePNG(t, diffPath, found.diff)
			t.Errorf("%d pixels differ, worst channel delta %d, first at %v; see %s and %s",
				found.pixels, found.worst, found.first, gotPath, diffPath)
		})
	}
}
