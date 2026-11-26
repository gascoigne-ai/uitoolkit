package core

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"runtime"
	"strings"
	"time"

	"github.com/gascoigne-ai/uitoolkit/render"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
)

// Fonts are the faces the toolkit draws with: UI and Small text, Mono for code, and bold forms of UI and Small.
type Fonts struct{ UI, Small, Mono, Bold, SmallBold render.Face }

// FontData is the content of a regular and a bold font file. The zero
// value means the embedded Go font.
type FontData struct{ Regular, Bold []byte }

// withDefault is d, or the embedded fonts when d is empty.
func (d FontData) withDefault(regular, bold []byte) FontData {
	if len(d.Regular) == 0 {
		return FontData{Regular: regular, Bold: bold}
	}
	if len(d.Bold) == 0 {
		d.Bold = d.Regular
	}
	return d
}

// fontSources are the fonts a Ctx draws with, as its device loaded them. A
// face is one of them at a size, so a change of scale or theme loads nothing.
type fontSources struct{ regular, bold, mono render.FontSource }

// loadFonts loads the chosen fonts (the embedded Go fonts where none was
// chosen) through device.
func loadFonts(device render.Device, chosen FontChoice) (fontSources, error) {
	ui := chosen.UI.withDefault(goregular.TTF, gobold.TTF)
	code := chosen.Code.withDefault(gomono.TTF, gobold.TTF)
	regular, err := device.NewFontSource(ui.Regular)
	if err != nil {
		return fontSources{}, fmt.Errorf("load regular font: %w", err)
	}
	bold, err := device.NewFontSource(ui.Bold)
	if err != nil {
		return fontSources{}, fmt.Errorf("load bold font: %w", err)
	}
	mono, err := device.NewFontSource(code.Regular)
	if err != nil {
		return fontSources{}, fmt.Errorf("load mono font: %w", err)
	}
	return fontSources{regular: regular, bold: bold, mono: mono}, nil
}

// facesOf are the faces of sources at the sizes of the theme, for a display
// with the given device scale factor.
func facesOf(th Theme, sources fontSources, scale float64) Fonts {
	return Fonts{
		UI:        render.Face{Source: sources.regular, Size: th.FontSize * scale},
		Small:     render.Face{Source: sources.regular, Size: th.SmallSize * scale},
		Mono:      render.Face{Source: sources.mono, Size: th.MonoSize * scale},
		Bold:      render.Face{Source: sources.bold, Size: th.FontSize * scale},
		SmallBold: render.Face{Source: sources.bold, Size: th.SmallSize * scale},
	}
}

// sizesOf measures the line heights of the theme's UI and small faces at
// scale 1, as LineH does, so the sizes never depend on the device scale.
func sizesOf(th Theme, regular render.FontSource) Sizes {
	line := func(size float64) int {
		metrics := regular.Metrics(size)
		return int(metrics.Ascent + metrics.Descent + 0.5)
	}
	return SizesFor(line(th.FontSize), line(th.SmallSize))
}

// FontChoice is the fonts picked for the interface and for code.
type FontChoice struct{ UI, Code FontData }

// Focuser is the state of a component that takes the keyboard: it holds
// the focus flag the window and the runtime read and write. It lives here
// so Ctx can name the focus owner by it; view.Focuser is the same type.
type Focuser interface {
	HasFocus() bool
	SetFocus(on bool)
}

// Ctx is what widgets draw and hit-test with: theme, fonts and this
// frame's input. Widgets work in logical pixels; Scale is the number of
// device pixels per logical pixel, applied only when drawing.
type Ctx struct {
	Th Theme
	F  Fonts
	// Device is the renderer's, which loads the fonts of F. NewCtx sets it and
	// nothing changes it after: F would keep the fonts of the old one.
	Device render.Device
	// sources are the loaded fonts behind F.
	sources fontSources
	// Sizes follow the line heights of F: set by NewCtx, SetTheme and
	// SetFonts, never by SetScale, since the scale multiplies at draw time.
	Sizes Sizes
	// Chosen are the fonts picked by the user; the zero value is the
	// embedded ones. They survive theme and scale changes.
	Chosen FontChoice
	In     Input
	// PointerX and PointerY are the pointer position in logical pixels as
	// sampled, set by view.Update before any container hides it from its
	// children; a drag that leaves its container keeps following the
	// pointer with them.
	PointerX, PointerY int
	Tick               int

	// Now is the clock the caret blinks by; nil is time.Now. CaretSince is
	// when the blink last started over (RestartCaret).
	Now        func() time.Time
	CaretSince time.Time
	// Caret is the shape of the text caret and whether it stays steady; the
	// zero value is a blinking block.
	Caret Caret
	// CaretReads counts the CaretVisible calls made while the caret blinks: a
	// drawing kept across frames shows the blink phase when the count moved
	// while it was drawn.
	CaretReads int
	Scale      float64

	// Clipboard is used by text widgets for copy, cut and paste.
	Clipboard Clipboard

	// FocusOwner is the one widget state (a *TextField, *TextArea or
	// *CodeView) that has the keyboard, nil for none. view.Update keeps it:
	// the app asks for focus by setting a state's Focused, and the owner
	// is whoever asked last; any other focused state, and one that is no
	// longer in the tree, is blurred.
	FocusOwner Focuser

	// Fired counts the handlers (OnClick, OnSelect, OnEnter, ...) that ran
	// since the app began counting. A handler changes app state, so the app
	// compares it before and after view.Update to know whether the tree it
	// built still shows that state. A widget that calls a handler adds one.
	Fired int

	// LayoutGen counts the passes of view.Layout and Laying is true inside
	// one: an element's measured size is remembered for one pass only, and
	// outside a pass (a size asked for by an Update handler) it is never
	// remembered, because the content may have changed since.
	LayoutGen int
	Laying    bool

	// Stats counts the draw calls of the frame when a benchmark or test sets
	// it; nil in the app, where counting costs one nil check.
	Stats *DrawStats

	// widths caches Width by face and text; see widthKey.
	widths map[widthKey]int
	// advances caches Advance by face; see faceKey. Cleared with widths.
	advances map[faceKey]float64

	// masks are the white masks drawn through Mask, by key.
	masks map[MaskKey]render.Offscreen

	// sizeScope is the memo of SizesScope.
	sizeScope sizeScope

	// TipText is the tooltip text of the element under the pointer, set by
	// view.Draw for widget.TooltipLayer to show.
	TipText string

	// Menu is the open popup menu; widget owns its type (*widget.MenuState) and
	// core only carries it, so that any element can open a menu without a
	// host. nil when none is open. It is any on purpose, an accepted exception
	// to toolkit rule 11: core cannot import widget, and a typed field would
	// need a widget type here. FocusOwner is typed because its interface
	// (Focuser) uses no widget type.
	Menu any

	// HostOS is the operating system the shortcut text and modifier follow:
	// runtime.GOOS from NewCtx; a test sets it to pin the text across machines.
	HostOS string

	// HoverOff turns hover feedback off for everything drawn while it is
	// set: elements under a modal overlay must not light up.
	HoverOff bool

	// Cursor is the mouse cursor shape widgets ask for this frame; the
	// application applies it. It is reset to CursorDefault every frame.
	Cursor Cursor
}

// Cursor is a mouse cursor shape.
type Cursor = render.Cursor

// The cursor shapes.
const (
	CursorDefault    = render.CursorDefault
	CursorResizeH    = render.CursorResizeH    // resize left/right, over a vertical divider
	CursorResizeV    = render.CursorResizeV    // resize up/down, over a horizontal divider
	CursorResizeNWSE = render.CursorResizeNWSE // resize along a diagonal, over a top-left or bottom-right corner
	CursorResizeNESW = render.CursorResizeNESW // resize along the other diagonal
	CursorMove       = render.CursorMove       // something is held and being moved: a row, a tab
)

// NewCtx returns a context at scale 1 that draws with the renderer device
// belongs to.
func NewCtx(th Theme, device render.Device) (*Ctx, error) {
	sources, err := loadFonts(device, FontChoice{})
	if err != nil {
		return nil, err
	}
	return &Ctx{
		Th: th, F: facesOf(th, sources, 1), Sizes: sizesOf(th, sources.regular), Device: device, sources: sources,
		Scale: 1, Clipboard: SystemClipboard(), HostOS: runtime.GOOS,
	}, nil
}

// sizeScope is the Sizes.Scope of one Sizes value, kept so a frame that
// builds many templates makes the map once.
type sizeScope struct {
	of    Sizes
	names map[string]any
}

// SizesScope is Sizes.Scope, remembered while Sizes does not change. The map
// is shared: read it, never write into it (Sizes.Scope gives a map to write
// into). The memo holds the Sizes it was made from and compares it on every
// read, so a plain assignment to the exported Sizes field cannot leave it
// stale; a zero Ctx works.
//
// Rejected: dropping the memo in SetTheme and SetFonts only, as the width
// cache is, since an assignment to Sizes would then read stale names. A
// package-level cache is global mutable state. Revisit when sizes become per
// face or per pane: then one entry is not enough.
func (c *Ctx) SizesScope() map[string]any {
	if c.sizeScope.names == nil || c.sizeScope.of != c.Sizes {
		c.sizeScope = sizeScope{of: c.Sizes, names: c.Sizes.Scope()}
	}
	return c.sizeScope.names
}

// CaretHalfPeriod is how long the caret is shown, and then hidden.
const CaretHalfPeriod = 500 * time.Millisecond

func (c *Ctx) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// RestartCaret makes a blinking caret visible now and starts its blink over,
// as an edit or a click does.
func (c *Ctx) RestartCaret() { c.CaretSince = c.now() }

// CaretVisible reports whether the caret is shown now: always while it is
// steady, else in the shown phase of its blink. The phase follows the clock,
// so it does not depend on how often the app updates or draws.
func (c *Ctx) CaretVisible() bool {
	if c.Caret.Steady {
		return true
	}
	c.CaretReads++
	if c.CaretSince.IsZero() {
		return true
	}
	return (c.now().Sub(c.CaretSince)/CaretHalfPeriod)%2 == 0
}

// SetScale resizes the faces when the device scale factor changes, for
// example when the window moves to another monitor. The fonts are loaded
// already, so the error is always nil.
func (c *Ctx) SetScale(scale float64) error {
	if scale <= 0 || scale == c.Scale {
		return nil
	}
	c.F, c.Scale = facesOf(c.Th, c.sources, scale), scale
	c.widths, c.advances = nil, nil
	c.dropMasks()
	return nil
}

// SetTheme switches theme and resizes the faces, whose sizes the theme sets.
// The fonts are loaded already, so the error is always nil.
func (c *Ctx) SetTheme(th Theme) error {
	c.Th, c.F, c.Sizes = th, facesOf(th, c.sources, c.Scale), sizesOf(th, c.sources.regular)
	c.widths, c.advances = nil, nil
	c.dropMasks()
	return nil
}

// SetFonts switches the interface and code fonts; empty data is the
// embedded font. Fonts that cannot be read change nothing.
func (c *Ctx) SetFonts(ui, code FontData) error {
	chosen := FontChoice{UI: ui, Code: code}
	sources, err := loadFonts(c.Device, chosen)
	if err != nil {
		return err
	}
	c.Chosen, c.sources = chosen, sources
	c.F, c.Sizes = facesOf(c.Th, sources, c.Scale), sizesOf(c.Th, sources.regular)
	c.widths, c.advances = nil, nil
	return nil
}

// Hovered reports whether the pointer is over r and hover feedback is on.
func (c *Ctx) Hovered(rect Rect) bool { return !c.HoverOff && c.In.Over(rect) }

// Color resolves a color token in the current theme.
func (c *Ctx) Color(tok Token) color.RGBA { return c.Th.Color(tok) }

func (c *Ctx) dev(value float32) float32 { return value * float32(c.Scale) }

// devRect is rect's corners in device pixels: left, top, right, bottom.
func (c *Ctx) devRect(rect Rect) (x0, y0, x1, y1 float32) {
	return c.dev(float32(rect.X)), c.dev(float32(rect.Y)), c.dev(float32(rect.X + rect.W)), c.dev(float32(rect.Y + rect.H))
}

// snapBox is rect in device pixels, each edge on a whole one.
func (c *Ctx) snapBox(rect Rect) render.Box {
	return render.Box{X0: c.snap(rect.X), Y0: c.snap(rect.Y), X1: c.snap(rect.X + rect.W), Y1: c.snap(rect.Y + rect.H)}
}

// snap is a logical coordinate as a device pixel position. Rounding the
// edges (not the sizes) lets neighbouring rectangles share an edge at any
// scale, with no gap or overlap.
func (c *Ctx) snap(pixels int) float32 { return float32(math.Round(float64(pixels) * c.Scale)) }

// LineWidth is the thickness in device pixels of the one-pixel lines and
// outlines: the scale rounded, never less than one, so a line is a whole
// number of device pixels and stays sharp at 125%, 150% and 175%.
func (c *Ctx) LineWidth() float32 { return float32(max(1, math.Round(c.Scale))) }

// Fill paints r with clr, its edges on device pixels.
func (c *Ctx) Fill(dst render.Canvas, rect Rect, clr color.RGBA) {
	dst.FillRect(c.snapBox(rect), clr)
}

// Outline strokes a one-pixel border just inside r.
func (c *Ctx) Outline(dst render.Canvas, rect Rect, clr color.RGBA) {
	dst.StrokeRect(c.snapBox(rect), c.LineWidth(), clr)
}

// HLine draws a one-pixel horizontal line.
func (c *Ctx) HLine(dst render.Canvas, posX, posY, width int, clr color.RGBA) {
	top := c.snap(posY)
	dst.FillRect(render.Box{X0: c.snap(posX), Y0: top, X1: c.snap(posX + width), Y1: top + c.LineWidth()}, clr)
}

// VLine draws a one-pixel vertical line.
func (c *Ctx) VLine(dst render.Canvas, posX, posY, height int, clr color.RGBA) {
	left := c.snap(posX)
	dst.FillRect(render.Box{X0: left, Y0: c.snap(posY), X1: left + c.LineWidth(), Y1: c.snap(posY + height)}, clr)
}

// Clip returns canvas limited to r, so drawing outside r is discarded.
func (c *Ctx) Clip(canvas render.Canvas, rect Rect) render.Canvas {
	left, top, right, bottom := c.devRect(rect)
	least := image.Pt(int(left), int(top))
	most := image.Pt(max(int(right+0.5), least.X), max(int(bottom+0.5), least.Y))
	return canvas.Sub(image.Rectangle{Min: least, Max: most})
}

// Width is the logical width of s in face, a tab as a gap of TabStop
// spaces (as Text draws it).
func (c *Ctx) Width(face render.Face, content string) int {
	key := widthKey{face: face, content: content}
	if width, ok := c.widths[key]; ok {
		return width
	}
	if c.Stats != nil {
		c.Stats.Widths++
	}
	width := int(face.Measure(tabsAsGaps(content))/c.Scale + 0.5)
	if len(c.widths) >= widthCacheLimit {
		clear(c.widths)
	}
	if c.widths == nil {
		c.widths = map[widthKey]int{}
	}
	c.widths[key] = width
	return width
}

// widthCacheLimit bounds Ctx.widths. A frame asks for the same few hundred
// labels again and again, so a bound only guards against a stream of new
// strings (a typed filter, a scrolled log); when it is reached the whole
// map is dropped, since an LRU would cost more than the refill. Revisit
// when a frame's hit rate falls under 80%.
const widthCacheLimit = 4096

// widthKey is what a width depends on. A face is its font source and size, so
// an element's own font (a new face value each call) still hits. The scale is
// not in the key: SetScale drops the cache.
type widthKey struct {
	face    render.Face
	content string
}

// faceKey is what a face's metrics depend on: the face.
type faceKey = render.Face

// Advance is the exact logical width of one character of a monospace face,
// not rounded: columns placed at whole multiples of the rounded Width drift
// away from the text, which is drawn with the exact advance. It is measured
// once per face and kept until the scale, theme or fonts change, because
// every mono widget asks for it several times a frame. The map holds one
// entry per face, so it needs no bound.
func (c *Ctx) Advance(face render.Face) float64 {
	if advance, ok := c.advances[face]; ok {
		return advance
	}
	if c.Stats != nil {
		c.Stats.Advances++
	}
	advance := face.Measure("0") / c.Scale
	if c.advances == nil {
		c.advances = map[faceKey]float64{}
	}
	c.advances[face] = advance
	return advance
}

// LineH is the logical line height of face.
func (c *Ctx) LineH(face render.Face) int {
	metrics := face.Metrics()
	return int((metrics.Ascent+metrics.Descent)/c.Scale + 0.5)
}

// Ascent is the logical distance from the top of a line of face to its baseline.
func (c *Ctx) Ascent(face render.Face) int { return int(face.Metrics().Ascent/c.Scale + 0.5) }

// StemBoost is how much heavier text is drawn than its plain coverage:
// Ctx.Text draws a second pass of the same glyphs at this alpha, so a
// partly covered edge pixel at coverage a ends at a + StemBoost*a*(1-a)
// while empty and full pixels stay as they are. It stands in for the stem
// darkening of macOS text: the Go fonts drawn unhinted with plain linear
// coverage look thin and grey, most of all light text on a dark panel.
// 0.5 lifts a half covered pixel to 0.625: slightly heavier, not bold
// (style.md, Typography).
const StemBoost = 0.5

// textPasses is the alpha of each pass Text draws the glyphs with, for a
// stem boost of strength: a full pass, then the same glyphs again at
// strength, which source-over blending turns into the coverage curve
// a + strength*a*(1-a).
func textPasses(strength float32) []float32 { return []float32{1, strength} }

// Text draws s with its top-left at x, y.
//
// Glyphs are unhinted outlines placed at fractional x positions (the
// Ebitengine renderer keeps up to 8 subpixel variants per glyph; y snaps to a
// whole device pixel), drawn twice at the same place so partly covered edge
// pixels get heavier by StemBoost, like the stem darkening of macOS text. The
// second pass takes the color scaled by StemBoost. Why a second pass rather
// than a coverage shader or a wider outline: the pass changes only partial
// pixels, never the outline or the advances, so Width, Advance and LineH
// measure exactly what is drawn (style.md, Typography).
func (c *Ctx) Text(dst render.Canvas, face render.Face, content string, posX, posY int, clr color.RGBA) {
	content = tabsAsGaps(content)
	left, top := float64(posX)*c.Scale, float64(posY)*c.Scale
	for _, alpha := range textPasses(StemBoost) {
		dst.DrawText(face, content, left, top, fade(clr, alpha))
	}
}

// fade is the premultiplied clr at alpha: every channel scaled, rounded to the
// nearest step.
func fade(clr color.RGBA, alpha float32) color.RGBA {
	scaled := func(value uint8) uint8 { return uint8(float32(value)*alpha + 0.5) }
	return color.RGBA{R: scaled(clr.R), G: scaled(clr.G), B: scaled(clr.B), A: scaled(clr.A)}
}

// TextMid draws s starting at x, vertically centered in r.
func (c *Ctx) TextMid(dst render.Canvas, face render.Face, content string, posX int, rect Rect, clr color.RGBA) {
	c.Text(dst, face, content, posX, rect.Y+(rect.H-c.LineH(face))/2, clr)
}

// TextCenter draws s centered in r.
func (c *Ctx) TextCenter(dst render.Canvas, face render.Face, content string, rect Rect, clr color.RGBA) {
	c.TextMid(dst, face, content, rect.X+(rect.W-c.Width(face, content))/2, rect, clr)
}

// Ellipsize shortens s with "..." until it fits maxW pixels. The longest
// prefix whose trimmed form plus the dots fits is found by binary search on
// the rune count, as CursorAt does, not by measuring every prefix.
func (c *Ctx) Ellipsize(face render.Face, content string, maxW int) string {
	if c.Width(face, content) <= maxW {
		return content
	}
	runes := []rune(content)
	candidate := func(count int) string { return strings.TrimRight(string(runes[:count]), " ") + "..." }
	fits := func(count int) bool { return c.Width(face, candidate(count)) <= maxW }
	low, high := 0, len(runes)-1 // the longest count that fits is in low-1..high
	for low < high {
		middle := (low + high + 1) / 2
		if fits(middle) {
			low = middle
		} else {
			high = middle - 1
		}
	}
	if !fits(low) {
		return ""
	}
	return candidate(low)
}

// CursorAt maps a pixel offset into s to the nearest rune boundary. The
// width of a prefix grows with its length, so the boundary is found by
// binary search: measuring every prefix would be quadratic in the length
// of the text.
func (c *Ctx) CursorAt(content string, posX int) int {
	runes := []rune(content)
	widthTo := func(count int) int { return c.Width(c.F.UI, string(runes[:count])) }
	low, high := 0, len(runes) // the first boundary whose width reaches x is in low..high
	for low < high {
		middle := (low + high) / 2
		if widthTo(middle) < posX {
			low = middle + 1
		} else {
			high = middle
		}
	}
	if low > 0 && abs(posX-widthTo(low-1)) <= abs(posX-widthTo(low)) {
		return low - 1
	}
	return low
}

func abs(count int) int {
	if count < 0 {
		return -count
	}
	return count
}

// Line draws an antialiased segment between logical points.
func (c *Ctx) Line(dst render.Canvas, x0, y0, x1, y1, width float32, clr color.RGBA) {
	dst.StrokeLine(c.dev(x0), c.dev(y0), c.dev(x1), c.dev(y1), c.dev(width), clr)
}

// Radius is the corner radius of the named control in this theme, after
// the user's corner setting: what every component draws its corners with.
func (c *Ctx) Radius(control string) float64 { return c.Th.Style(control).Radius }

// FillStyled paints r the way a control style says: a flat or vertical
// gradient fill from top to bottom, with rounded corners when the style
// has a radius. Square flat fills stay on the crisp rectangle path; a square
// gradient is not snapped to device pixels.
func (c *Ctx) FillStyled(dst render.Canvas, rect Rect, top, bottom color.RGBA, radius float64) {
	if radius*c.Scale >= 1 {
		c.fillRounded(dst, rect, top, bottom, radius, CornersAll)
		return
	}
	if top == bottom {
		c.Fill(dst, rect, top)
		return
	}
	left, upper, right, lower := c.devRect(rect)
	dst.FillRectGradient(render.Box{X0: left, Y0: upper, X1: right, Y1: lower}, top, bottom)
}

// fillRounded paints rect with the corners in the set rounded, its edges
// snapped to device pixels as Fill's are.
func (c *Ctx) fillRounded(dst render.Canvas, rect Rect, top, bottom color.RGBA, radius float64, corners Corners) {
	if c.Stats != nil {
		c.Stats.MaskFills++
	}
	dst.FillRoundRect(c.snapBox(rect), c.dev(float32(radius)), corners, top, bottom)
}

// DrawStats counts what one frame asks the renderer for, to pin and measure
// the batching of rounded shapes. Set Ctx.Stats to collect; the caller zeroes
// it between frames.
type DrawStats struct {
	MaskFills int // Canvas.FillRoundRect calls: rounded fills, which a renderer draws from a corner mask
	RingFills int // Canvas.StrokeRoundRect calls: rounded outlines, drawn from a ring mask
	Widths    int // Ctx.Width measurements
	Advances  int // Ctx.Advance measurements
	Elements  int // elements the view runtime drew; one skipped outside its clip is not counted
}

// Corners is a set of rectangle corners.
type Corners = render.Corners

// The corners of a rectangle.
const (
	CornerTL = render.CornerTL
	CornerTR = render.CornerTR
	CornerBR = render.CornerBR
	CornerBL = render.CornerBL

	CornersTop    = render.CornersTop
	CornersBottom = render.CornersBottom
	CornersAll    = render.CornersAll
)

// FillStyledCorners is FillStyled with only the given corners rounded, for
// a bar or a column that sits in the corner of a rounded panel: its corner
// follows the panel's, with anti-aliased (alpha) edges, instead of poking
// out past it.
func (c *Ctx) FillStyledCorners(dst render.Canvas, rect Rect, top, bottom color.RGBA, radius float64, corners Corners) {
	if radius*c.Scale < 1 || corners == 0 {
		c.FillStyled(dst, rect, top, bottom, 0)
		return
	}
	c.fillRounded(dst, rect, top, bottom, radius, corners)
}

// OutlineStyled strokes a border just inside r, LineWidth thick, with
// rounded corners when radius is set. The edges snap to device pixels as
// Outline's do. The width is LineWidth (the unrounded scale a stroke used to
// take would blur on a fractional scale and not match Outline's lines).
func (c *Ctx) OutlineStyled(dst render.Canvas, rect Rect, clr color.RGBA, radius float64) {
	if radius*c.Scale < 1 {
		c.Outline(dst, rect, clr)
		return
	}
	if c.Stats != nil {
		c.Stats.RingFills++
	}
	dst.StrokeRoundRect(c.snapBox(rect), c.dev(float32(radius)), c.LineWidth(), clr)
}
