package sections

import (
	"image"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// Fixed sizes of the text section.
const (
	formLabelW  = 90
	formFieldW  = 320
	notesH      = 220
	searchHint  = "Type to search..."
	startNotes  = "Select with the mouse or with Shift and the arrows.\nUndo, redo, copy and paste work.\n\n\tTab indents.\nA long line scrolls sideways unless the area wraps."
	notesRemark = "A text area keeps Tab to indent; Escape and a click elsewhere leave it."
)

var textPoses = []Pose{{Apply: func(s *Showcase, _ *core.Ctx) image.Point {
	s.widgets.name.Focused = true
	return OffCanvas
}}}

func textGo(s *Showcase, _ *core.Ctx) Node {
	state := &s.widgets
	row := func(label string, field *TextField, opts ...Option) Node {
		return FormRow(label, formLabelW, Field(field, append([]Option{W(formFieldW), Bg(core.Sunken)}, opts...)...))
	}
	return Scroll(&state.textPage, Pad(core.Gutter), Gap(core.Gutter),
		Caption("Text fields"),
		row("Name", &state.name),
		row("Password", &state.secret),
		row("Search", &state.search, Placeholder(searchHint)),
		Caption("Text area"),
		Area(&state.notes, H(notesH), Bg(core.Sunken), Control("field")),
		Text(notesRemark, Small(), Muted()),
	)
}

func textMarkup(s *Showcase, ctx *core.Ctx) Node {
	state := &s.widgets
	return fill(ctx, textTemplate, markup.Scope{
		"page": &state.textPage, "name": &state.name, "secret": &state.secret, "search": &state.search, "notes": &state.notes,
		"labelW": formLabelW, "fieldW": formFieldW, "notesH": notesH, "searchHint": searchHint, "remark": notesRemark,
	})
}

var textTemplate = markup.Must("text.vml", textSource)

const textSource = `
<scroll state="page" pad="{{gutter}}" gap="{{gutter}}">
  <caption>Text fields</caption>
  <form-row label="Name" label-width="{{labelW}}">
    <field state="name" w="{{fieldW}}" bg="sunken"/>
  </form-row>
  <form-row label="Password" label-width="{{labelW}}">
    <field state="secret" w="{{fieldW}}" bg="sunken"/>
  </form-row>
  <form-row label="Search" label-width="{{labelW}}">
    <field state="search" w="{{fieldW}}" bg="sunken" placeholder="{{searchHint}}"/>
  </form-row>
  <caption>Text area</caption>
  <area state="notes" h="{{notesH}}" bg="sunken" control="field"/>
  <text small muted>{{remark}}</text>
</scroll>`
