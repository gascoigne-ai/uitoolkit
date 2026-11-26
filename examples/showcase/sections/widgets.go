package sections

import (
	"image"

	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/view"
	"github.com/gascoigne-ai/uitoolkit/widget"
)

// widgets is the state the widgets edit. It is one field per thing and
// outlives the section that shows it, so switching sections loses nothing.
type widgets struct {
	tooltip widget.TooltipState
	topBar  widget.ScrollState

	// buttons
	buttonsPage      widget.ScrollState
	clicks           int
	chip, viewChosen int

	// text
	textPage             widget.ScrollState
	name, secret, search widget.TextField
	notes                widget.TextArea

	// toggles
	togglesPage                        widget.ScrollState
	wrapLines, showHidden, termsAgreed bool
	autosave, offline                  bool
	volume, zoom                       int
	volumeDrag, zoomDrag               widget.SliderState

	// lists
	tabStrip       widget.ScrollState
	tab, detailTab int
	split          widget.SplitState
	list, detail   widget.ScrollState
	picked         int
	table          widget.TableState

	// overlays
	overlaysPage              widget.ScrollState
	lastPick, lastAnswer      string
	confirm                   widget.TextField
	dialogMove                view.MoveState
	menuButton, tooltipAnchor *view.El // where the poses open the menu and rest the pointer

	// code
	code     widget.CodeView
	wrapCode bool

	// images
	imagesPage widget.ScrollState
	demo       deviceImage

	// typography
	typographyPage widget.ScrollState

	// theming
	themingPage   widget.ScrollState
	sample        widget.TextField
	sampleChecked bool
	sampleOn      bool
}

// deviceImage is the demo picture as the device that draws it made it: a
// renderer draws only its own images.
type deviceImage struct {
	device  render.Device
	picture render.Image
	source  *image.RGBA
}

func newWidgets() widgets {
	state := widgets{
		name:   widget.NewTextField("Ada Lovelace"),
		secret: widget.NewTextField("correct horse"),
		search: widget.NewTextField(""),
		notes:  widget.NewTextArea(startNotes),

		wrapLines: true, termsAgreed: true, autosave: true,
		volume: 65, zoom: 100,

		lastPick: "nothing yet", lastAnswer: "not asked yet",
		confirm: widget.NewTextField(""),

		sample: widget.NewTextField("Sample text"), sampleChecked: true, sampleOn: true,
	}
	state.secret.Masked = true
	state.code.SetLines(demoSource())
	return state
}
