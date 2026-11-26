package sections

import (
	"image"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	"github.com/gascoigne-ai/uitoolkit/view"
)

// builder makes one part of the view from the state.
type builder = func(*Showcase, *core.Ctx) view.Node

// Section is one page of the showcase, built with the Go DSL and from
// markup.
type Section struct {
	ID, Label  string
	Go, Markup builder
	// Poses are the states a golden screenshot is taken in besides the plain
	// one of a section that has none. A menu and a dialog are modal, which
	// turns hover off under them, so a menu, a tooltip and a dialog cannot be
	// seen at once and each has its own pose.
	Poses []Pose
}

// Pose puts a section that has been drawn once into a state worth a
// screenshot and returns where the pointer rests, in logical pixels. Poses
// are for the Go form: a pose that needs the place of an element does nothing
// when only the markup form has been drawn.
type Pose struct {
	Name  string // "" for the only pose of a section
	Apply func(*Showcase, *core.Ctx) image.Point
}

// OffCanvas is a pointer that is over nothing.
var OffCanvas = image.Pt(-1, -1)

// Sections are the pages in the order of the nav. A new page is a new row.
var Sections = []Section{
	{ID: "buttons", Label: "Buttons and chips", Go: buttonsGo, Markup: buttonsMarkup},
	{ID: "text", Label: "Text input", Go: textGo, Markup: textMarkup, Poses: textPoses},
	{ID: "toggles", Label: "Toggles and sliders", Go: togglesGo, Markup: togglesMarkup},
	{ID: "lists", Label: "Lists and panes", Go: listsGo, Markup: listsMarkup, Poses: listsPoses},
	{ID: "layout", Label: "Layout and custom", Go: layoutGo, Markup: layoutMarkup},
	{ID: "overlays", Label: "Menus and dialogs", Go: overlaysGo, Markup: overlaysMarkup, Poses: overlaysPoses},
	{ID: "code", Label: "Code view", Go: codeGo, Markup: codeMarkup, Poses: codePoses},
	{ID: "images", Label: "Images and icons", Go: imagesGo, Markup: imagesMarkup},
	{ID: "typography", Label: "Typography", Go: typographyGo, Markup: typographyMarkup},
	{ID: "theming", Label: "Theming", Go: themingGo, Markup: themingMarkup},
}

// Form is one way a view is built.
type Form struct {
	ID, Label string
	section   func(Section) builder
	dialog    func(dialogView) builder
	toasts    func(showcase *Showcase, ctx *core.Ctx, width, height int) view.Node
	menu      builder
}

// Forms are the choices of the Form control.
var Forms = []Form{
	{ID: "go", Label: "Go", section: func(section Section) builder { return section.Go }, dialog: func(view dialogView) builder { return view.Go }, toasts: toastsGo, menu: menuGo},
	{ID: "markup", Label: "Markup", section: func(section Section) builder { return section.Markup }, dialog: func(view dialogView) builder { return view.Markup }, toasts: toastsMarkup, menu: menuMarkup},
}

// fill builds template for scope. The sizes and spacings of the context are
// the fallback names ({{rowH}}, {{gutter}}), so no template holds a size.
func fill(ctx *core.Ctx, template *markup.Template, scope markup.Scope) view.Node {
	return template.MustBuild(scope, ctx.SizesScope())
}
