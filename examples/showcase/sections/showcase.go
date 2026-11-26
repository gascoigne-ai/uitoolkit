// Package sections is the user interface of the showcase: one state, a table
// of sections that shows every widget, and the frame around them. It draws
// through render only, so it runs, and is tested, without a display.
//
// Each section is built twice, with the Go DSL and from a markup template,
// and the Form control switches between the two. A template is in a const at
// the end of its section's file. Widgets markup has no element for (switch,
// disabled-switch, paragraph) are Go nodes put into the template with slot.
//
// State is in memory only: nothing is saved.
package sections

import (
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/theming"
	"github.com/gascoigne-ai/uitoolkit/widget"
)

// ScaleSteps are the choices of the scale control. A step multiplies the
// display's own device scale factor.
var ScaleSteps = []float64{1, 1.25, 1.5, 2}

// Showcase is the state of the program; the view is a function of it. Make
// one with New. It is used by one goroutine.
type Showcase struct {
	// Section is an index into Sections, Form into Forms and Scale into
	// ScaleSteps. Select, SetForm and SetScale keep them in range.
	Section, Form, Scale int
	// Prefs are the chosen theme and contrast; SetThemeID and SetContrast
	// change them.
	Prefs theming.Prefs
	// Overlay shows the frame time and the draw counts; F3 toggles it.
	Overlay bool

	// ThemeJSON is a theme file the user edits, laid over the chosen theme;
	// ApplyThemeJSON reads it. ThemeErr says why the last edit was refused,
	// "" when it was not.
	ThemeJSON widget.TextArea
	ThemeErr  string

	catalog *theming.Catalog
	// lastGood is the theme in effect: the chosen one with the last edit that
	// could be read laid over it.
	lastGood     core.Theme
	themeChanged bool // lastGood is not on the context yet

	widgets widgets
	// dialog is the open dialog, nil for none. The open popup menu is the
	// context's (core.Ctx.Menu).
	dialog *widget.DialogState

	now        func() time.Time
	frameStart time.Time
	frameTime  time.Duration
	// stats are the draw counts of the last frame, for the overlay.
	stats core.DrawStats
}

// New returns a showcase at its first section in the dark theme. now is the
// clock the frame time is measured by.
func New(now func() time.Time) *Showcase {
	showcase := &Showcase{
		Prefs:     theming.Prefs{Theme: "dark", Contrast: "normal"},
		ThemeJSON: widget.NewTextArea(startThemeJSON),
		catalog:   theming.NewCatalog(theming.Builtins()),
		widgets:   newWidgets(),
		now:       now,
	}
	showcase.retheme()
	return showcase
}

// Select shows the section at index, kept inside the table. It closes the
// popup menu and the dialog, which would otherwise cover the new section.
func (s *Showcase) Select(ctx *core.Ctx, index int) {
	s.Section = clampIndex(index, len(Sections))
	s.closeLayers(ctx)
}

func (s *Showcase) closeLayers(ctx *core.Ctx) {
	widget.CloseMenu(ctx)
	s.dialog = nil
}

// SetForm picks how sections are built: an index into Forms.
func (s *Showcase) SetForm(index int) { s.Form = clampIndex(index, len(Forms)) }

// SetScale picks a step of ScaleSteps, kept inside the table.
func (s *Showcase) SetScale(index int) { s.Scale = clampIndex(index, len(ScaleSteps)) }

// ScaleFactor is the chosen step.
func (s *Showcase) ScaleFactor() float64 { return ScaleSteps[s.Scale] }

func clampIndex(index, count int) int { return min(max(index, 0), count-1) }
