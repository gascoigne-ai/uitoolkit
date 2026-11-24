package render

// Verb is what a path segment does.
type Verb uint8

// The verbs.
const (
	VerbMove Verb = iota
	VerbLine
	VerbCubic
	VerbClose
)

// Segment is one step of a path.
type Segment struct {
	Verb Verb
	// Coords holds the points of the verb as x, y pairs: one point for
	// VerbMove and VerbLine, two control points and the end point for
	// VerbCubic, none for VerbClose. The rest is zero.
	Coords [6]float32
}

// Path is an outline made of lines and cubic curves, in device pixels. The
// zero value is an empty path.
type Path struct{ segments []Segment }

// MoveTo starts a new contour at (toX, toY).
func (p *Path) MoveTo(toX, toY float32) {
	p.add(VerbMove, [6]float32{toX, toY})
}

// LineTo adds a straight line to (toX, toY).
func (p *Path) LineTo(toX, toY float32) {
	p.add(VerbLine, [6]float32{toX, toY})
}

// CubicTo adds a cubic curve with control points (x1, y1) and (x2, y2) that
// ends at (x3, y3).
func (p *Path) CubicTo(x1, y1, x2, y2, x3, y3 float32) {
	p.add(VerbCubic, [6]float32{x1, y1, x2, y2, x3, y3})
}

// Close joins the current contour to its start.
func (p *Path) Close() {
	p.add(VerbClose, [6]float32{})
}

// Segments returns the steps in the order they were added. The caller must not
// modify the slice.
func (p *Path) Segments() []Segment { return p.segments }

func (p *Path) add(verb Verb, coords [6]float32) {
	p.segments = append(p.segments, Segment{Verb: verb, Coords: coords})
}

// LineCap is the shape of the two ends of an open stroke.
type LineCap uint8

// The line caps. The zero value is LineCapButt.
const (
	LineCapButt LineCap = iota
	LineCapRound
	LineCapSquare
)

// LineJoin is the shape of a corner of a stroke.
type LineJoin uint8

// The line joins. The zero value is LineJoinMiter.
const (
	LineJoinMiter LineJoin = iota
	LineJoinBevel
	LineJoinRound
)

// Stroke is how the outline of a path is drawn.
type Stroke struct {
	// Width is the thickness in device pixels.
	Width float32
	Cap   LineCap
	Join  LineJoin
	// MiterLimit is the longest a LineJoinMiter corner may be, as a multiple
	// of Width; a sharper corner is beveled. Zero, the zero value, means
	// every join is a bevel: it is what the toolkit draws today, and there is
	// no default limit that stands in for it.
	MiterLimit float32
}
