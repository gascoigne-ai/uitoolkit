package core

// Sizes are the sizes every component shares, in logical pixels: heights,
// widths and spacings for one pair of UI line heights under one theme's
// Metrics (ui-guidelines.md, Sizes). It follows the faces the Ctx really
// draws with, so a text size the user picks never clips a row. Ctx.Sizes
// holds the value; widgets read it every frame and nothing writes it but
// NewCtx, SetTheme and SetFonts.
//
// Why a struct of fields: layout reads it every frame and it changes only
// with the theme or fonts. Rejected: a map keyed by scope name (untyped,
// and a typo compiles). Revisit past about 80 fields, or when a pane needs
// sizes per face (rows in the mono face).
type Sizes struct {
	// Line and SmallLine are the measured line heights of the UI and the
	// small face, without Leading.
	Line, SmallLine int
	// Leading is the theme's extra leading, see Metrics.Leading.
	Leading int
	// Steps are the spacing steps in pixels.
	Steps Steps

	// Height classes, so neighbours line up: Row for list and table rows and
	// standard controls, Compact for chips, icon buttons and inline fields
	// that sit in a row of that height, Slim for a pill and a button in the
	// status bar, Label for a section caption and Note for a one-line note in
	// the small face.
	Row, Compact, Slim, Label, Note int

	// Bars, built from Row and a spacing role: a bar of controls over a pane,
	// a strip of tabs (with TabGap above them), and a dialog's title bar and
	// the row of controls under it.
	Toolbar, TabBar, DialogTitle, DialogToolbar int

	IconBtnW int // an icon button: square room around a compact control
	WordBtnW int // a button that holds one short word (Log, Raw, Reset)
	Checkbox int // the side of a checkbox: one small line

	// Spacing roles. Gutter is the inset between a container and its content,
	// and the gap between its sections. Gaps between neighbours come in three
	// steps: NearGap for items that read as one (a label and its count),
	// TightGap for controls side by side (a field and its button, a row of
	// chips), LooseGap for the blocks of a pane stacked in a column and for
	// rows whose items need more air.
	TabLead      int // room before the first tab of a strip
	EdgeW        int // the width of an open tab's outline
	CaretW       int // the thickness of a bar or underline caret
	NearGap      int
	ShadowOffset int // how far a shadow is shifted down and right
	TightGap     int
	TabGap       int // space above tabs, so their outline never touches the banner or bar above
	RuleInset    int // how far a menu's separator line stays off the menu's sides
	EdgeGrab     int // how far inside its border a press grabs an edge (resize, or no drag)
	LooseGap     int
	CellPad      int // the inset of a table cell and of a header cell
	LabelPad     int // the side inset of the label of a pill or a tab
	TabSepInset  int // how far the line between two tabs stays off the strip's top and bottom
	Gutter       int
	DialogPad    int // the inset of a dialog's title bar, toolbar, footer and message body

	DialogW          int // a message dialog: wide enough for a sentence and two buttons
	DialogButtonMinW int // the least width of a dialog button, so Yes and No match
	// DialogMargin and DialogMaxH are for app layouts that size a dialog to
	// the window; no toolkit widget applies them.
	DialogMargin int // kept free between a dialog and the window's edge
	DialogMaxH   int // the tallest a dialog grows by itself
}

// The UI line height is limited so a huge font cannot blow up every row.
// A theme cannot move these: a theme file is untrusted input.
const (
	MinLine = 13
	MaxLine = 28
)

// The line heights of the default faces (Go regular at 15 and 13 px).
const (
	baseLine      = 17
	baseSmallLine = 15
)

// sizeID is a row of sizeRules.
type sizeID int

// sizeRule is one named size: the name templates and theme files use, the
// field of Sizes that holds it, the rule that derives it, and the bounds a
// theme may pin it within. The bounds keep the reach of a shadow and the
// integer arithmetic on an untrusted theme file small.
type sizeRule struct {
	name     string
	field    func(sizes *Sizes) *int
	derive   func(base sizeBase, sizes *Sizes) int
	min, max int
}

// sizeBase is what the rules start from besides the lines, the leading and
// the steps, which Sizes holds itself.
type sizeBase struct {
	unit int
}

// lineBox and smallBox are the lines the height classes are built on: the
// measured line plus the theme's leading.
func lineBox(sizes *Sizes) int  { return sizes.Line + sizes.Leading }
func smallBox(sizes *Sizes) int { return sizes.SmallLine + sizes.Leading }

// step is the rule of a size that is one spacing step.
func step(pick func(steps Steps) int) func(sizeBase, *Sizes) int {
	return func(_ sizeBase, sizes *Sizes) int { return pick(sizes.Steps) }
}

func hairStep(steps Steps) int   { return steps.Hair }
func nearStep(steps Steps) int   { return steps.Near }
func tightStep(steps Steps) int  { return steps.Tight }
func looseStep(steps Steps) int  { return steps.Loose }
func labelStep(steps Steps) int  { return steps.Label }
func gutterStep(steps Steps) int { return steps.Gutter }
func dialogStep(steps Steps) int { return steps.Dialog }

// linePlus is a height class: a line box plus units grid steps of padding.
func linePlus(units int) func(sizeBase, *Sizes) int {
	return func(base sizeBase, sizes *Sizes) int { return lineBox(sizes) + units*base.unit }
}

// rowPlus is a bar: a row with a spacing role above and below.
func rowPlus(pad func(sizes *Sizes) int) func(sizeBase, *Sizes) int {
	return func(_ sizeBase, sizes *Sizes) int { return sizes.Row + 2*pad(sizes) }
}

func fixed(pixels int) func(sizeBase, *Sizes) int {
	return func(sizeBase, *Sizes) int { return pixels }
}

// sizeRules are all named sizes, in dependency order: a rule reads only the
// knobs and the rows above it, so a pinned size moves what is derived from
// it. Resolution and Scope scan it; a name is looked up only when a theme
// is parsed. Rejected: a map by name, which loses the order. Revisit past
// 150 rows, or when something looks a name up every frame.
var sizeRules = [...]sizeRule{
	{"tabLead", func(sizes *Sizes) *int { return &sizes.TabLead }, step(hairStep), 0, 32},
	{"edgeW", func(sizes *Sizes) *int { return &sizes.EdgeW }, step(hairStep), 1, 8},
	{"caretW", func(sizes *Sizes) *int { return &sizes.CaretW }, step(hairStep), 1, 8},
	{"nearGap", func(sizes *Sizes) *int { return &sizes.NearGap }, step(nearStep), 0, 64},
	{"shadowOffset", func(sizes *Sizes) *int { return &sizes.ShadowOffset }, step(nearStep), 0, 16},
	{"tightGap", func(sizes *Sizes) *int { return &sizes.TightGap }, step(tightStep), 0, 64},
	{"tabGap", func(sizes *Sizes) *int { return &sizes.TabGap }, step(tightStep), 0, 64},
	{"ruleInset", func(sizes *Sizes) *int { return &sizes.RuleInset }, step(tightStep), 0, 64},
	{"edgeGrab", func(sizes *Sizes) *int { return &sizes.EdgeGrab }, step(tightStep), 1, 32},
	{"looseGap", func(sizes *Sizes) *int { return &sizes.LooseGap }, step(looseStep), 0, 64},
	{"cellPad", func(sizes *Sizes) *int { return &sizes.CellPad }, step(looseStep), 0, 64},
	{"labelPad", func(sizes *Sizes) *int { return &sizes.LabelPad }, step(labelStep), 0, 64},
	{"tabSepInset", func(sizes *Sizes) *int { return &sizes.TabSepInset }, step(labelStep), 0, 64},
	{"gutter", func(sizes *Sizes) *int { return &sizes.Gutter }, step(gutterStep), 0, 64},
	{"dialogPad", func(sizes *Sizes) *int { return &sizes.DialogPad }, step(dialogStep), 0, 64},

	{"rowH", func(sizes *Sizes) *int { return &sizes.Row }, linePlus(6), 8, 160},
	{"compactH", func(sizes *Sizes) *int { return &sizes.Compact }, linePlus(4), 8, 160},
	{"slimH", func(sizes *Sizes) *int { return &sizes.Slim }, linePlus(2), 8, 160},
	{"labelH", func(sizes *Sizes) *int { return &sizes.Label }, linePlus(3), 8, 160},
	{"noteH", func(sizes *Sizes) *int { return &sizes.Note },
		func(base sizeBase, sizes *Sizes) int { return smallBox(sizes) + 2*base.unit }, 8, 160},

	{"toolbarH", func(sizes *Sizes) *int { return &sizes.Toolbar },
		rowPlus(func(sizes *Sizes) int { return sizes.NearGap }), 8, 320},
	{"tabBarH", func(sizes *Sizes) *int { return &sizes.TabBar },
		rowPlus(func(sizes *Sizes) int { return sizes.TabGap }), 8, 320},
	{"dialogToolbarH", func(sizes *Sizes) *int { return &sizes.DialogToolbar },
		rowPlus(func(sizes *Sizes) int { return sizes.TightGap }), 8, 320},
	{"dialogTitleH", func(sizes *Sizes) *int { return &sizes.DialogTitle },
		rowPlus(func(sizes *Sizes) int { return sizes.LooseGap }), 8, 320},
	{"iconBtnW", func(sizes *Sizes) *int { return &sizes.IconBtnW },
		func(_ sizeBase, sizes *Sizes) int { return sizes.Compact + 2*sizes.NearGap }, 8, 320},
	{"wordBtnW", func(sizes *Sizes) *int { return &sizes.WordBtnW },
		func(_ sizeBase, sizes *Sizes) int { return sizes.IconBtnW + sizes.Gutter }, 8, 320},
	{"checkboxSize", func(sizes *Sizes) *int { return &sizes.Checkbox },
		func(_ sizeBase, sizes *Sizes) int { return sizes.SmallLine }, 4, 64},

	{"dialogW", func(sizes *Sizes) *int { return &sizes.DialogW },
		func(base sizeBase, _ *Sizes) int { return 190 * base.unit }, 120, 1600},
	{"dialogButtonMinW", func(sizes *Sizes) *int { return &sizes.DialogButtonMinW },
		func(_ sizeBase, sizes *Sizes) int { return 5 * sizes.Steps.Dialog }, 0, 400},
	{"dialogMargin", func(sizes *Sizes) *int { return &sizes.DialogMargin }, fixed(40), 0, 200},
	{"dialogMaxH", func(sizes *Sizes) *int { return &sizes.DialogMaxH }, fixed(560), 120, 4000},
}

// SizesFor applies the size rules of metrics to a UI line height, clamped to
// MinLine..MaxLine, and a small line height, clamped to 1..line. It never
// fails: knobs outside their bounds, which only a Go literal can hold, are
// clamped too.
func SizesFor(line, small int, metrics Metrics) Sizes {
	line = min(max(line, MinLine), MaxLine)
	sizes := Sizes{Line: line, SmallLine: min(max(small, 1), line)}
	sizes.Leading = min(max(metrics.Leading, MinLeading), MaxLeading)
	sizes.Steps = metrics.stepPixels()
	base := sizeBase{unit: metrics.unit()}
	for id, rule := range sizeRules {
		pixels, pinned := metrics.pins[sizeID(id)]
		if !pinned {
			pixels = rule.derive(base, &sizes)
		}
		*rule.field(&sizes) = pixels
	}
	return sizes
}

// defaultSizes is computed once: HOf and every other size option reads it
// on every build.
var defaultSizes = SizesFor(baseLine, baseSmallLine, Metrics{})

// DefaultSizes are the sizes of the default faces and the default Metrics.
func DefaultSizes() Sizes { return defaultSizes }

// Scope is the sizes and spacings under the names templates use, which are
// also the names a theme file pins them by.
func (s Sizes) Scope() map[string]any {
	scope := make(map[string]any, len(sizeRules))
	for _, rule := range sizeRules {
		scope[rule.name] = *rule.field(&s)
	}
	return scope
}

// Size roles: one function per named size of Sizes, so a view names the size
// it wants (view.HOf(core.RowOf)) and Layout reads it from the context.
func RowOf(sizes Sizes) int              { return sizes.Row }
func CompactOf(sizes Sizes) int          { return sizes.Compact }
func SlimOf(sizes Sizes) int             { return sizes.Slim }
func LabelOf(sizes Sizes) int            { return sizes.Label }
func NoteOf(sizes Sizes) int             { return sizes.Note }
func ToolbarOf(sizes Sizes) int          { return sizes.Toolbar }
func TabBarOf(sizes Sizes) int           { return sizes.TabBar }
func DialogTitleOf(sizes Sizes) int      { return sizes.DialogTitle }
func DialogToolbarOf(sizes Sizes) int    { return sizes.DialogToolbar }
func IconBtnOf(sizes Sizes) int          { return sizes.IconBtnW }
func WordBtnOf(sizes Sizes) int          { return sizes.WordBtnW }
func CheckboxOf(sizes Sizes) int         { return sizes.Checkbox }
func TabLeadOf(sizes Sizes) int          { return sizes.TabLead }
func EdgeWOf(sizes Sizes) int            { return sizes.EdgeW }
func CaretWOf(sizes Sizes) int           { return sizes.CaretW }
func NearGapOf(sizes Sizes) int          { return sizes.NearGap }
func ShadowOffsetOf(sizes Sizes) int     { return sizes.ShadowOffset }
func TightGapOf(sizes Sizes) int         { return sizes.TightGap }
func TabGapOf(sizes Sizes) int           { return sizes.TabGap }
func RuleInsetOf(sizes Sizes) int        { return sizes.RuleInset }
func EdgeGrabOf(sizes Sizes) int         { return sizes.EdgeGrab }
func LooseGapOf(sizes Sizes) int         { return sizes.LooseGap }
func CellPadOf(sizes Sizes) int          { return sizes.CellPad }
func LabelPadOf(sizes Sizes) int         { return sizes.LabelPad }
func TabSepInsetOf(sizes Sizes) int      { return sizes.TabSepInset }
func GutterOf(sizes Sizes) int           { return sizes.Gutter }
func DialogPadOf(sizes Sizes) int        { return sizes.DialogPad }
func DialogWOf(sizes Sizes) int          { return sizes.DialogW }
func DialogButtonMinWOf(sizes Sizes) int { return sizes.DialogButtonMinW }
func DialogMarginOf(sizes Sizes) int     { return sizes.DialogMargin }
func DialogMaxHOf(sizes Sizes) int       { return sizes.DialogMaxH }
