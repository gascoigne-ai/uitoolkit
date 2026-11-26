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
	caretChipW  = 96
	notesH      = 220
	searchHint  = "Type to search..."
	startNotes  = "Select with the mouse or with Shift and the arrows.\nUndo, redo, copy and paste work.\n\n\tTab indents.\nA long line scrolls sideways unless the area wraps."
	notesRemark = "A text area keeps Tab to indent; Escape and a click elsewhere leave it."
)

func (s *Showcase) caretChoices() []choice {
	shapes := []struct {
		label string
		shape core.CaretShape
	}{{"Block", core.CaretBlock}, {"Bar", core.CaretBar}, {"Underline", core.CaretUnderline}}
	var choices []choice
	for _, entry := range shapes {
		choices = append(choices, choice{
			Label: entry.label, Selected: s.Caret.Shape == entry.shape,
			Pick: func() { s.Caret.Shape = entry.shape },
		})
	}
	return choices
}

var textPoses = []Pose{{Apply: func(s *Showcase, _ *core.Ctx) image.Point {
	s.widgets.name.Focused = true
	return OffCanvas
}}}

func textGo(s *Showcase, _ *core.Ctx) Node {
	state := &s.widgets
	row := func(label string, field *TextField, opts ...Option) Node {
		return FormRow(label, formLabelW, Field(field, append([]Option{W(formFieldW), Bg(core.Sunken)}, opts...)...))
	}
	return Scroll(&state.textPage, PadOf(core.GutterOf), GapOf(core.GutterOf),
		Caption("Text fields"),
		row("Name", &state.name, TextMenu(&state.name)),
		row("Password", &state.secret),
		row("Search", &state.search, Placeholder(searchHint), TextMenu(&state.search)),
		Caption("Text area"),
		Area(&state.notes, H(notesH), Bg(core.Sunken), Control("field")),
		Text(notesRemark, Small(), Muted()),
		Caption("Caret"),
		FormRow("Shape", formLabelW, Row(GapOf(core.TightGapOf),
			For(s.caretChoices(), func(_ int, entry choice) Node {
				return SegmentChip(entry.Label, entry.Selected, W(caretChipW), OnClick(entry.Pick))
			}),
		)),
		CheckRow(!s.Caret.Steady, OnClick(flip(&s.Caret.Steady)), AlignItems(AlignCenter), Text("Blink")),
	)
}

func textMarkup(s *Showcase, ctx *core.Ctx) Node {
	state := &s.widgets
	return fill(ctx, textTemplate, markup.Scope{
		"page": &state.textPage, "name": &state.name, "secret": &state.secret, "search": &state.search, "notes": &state.notes,
		"carets": s.caretChoices(), "blink": !s.Caret.Steady, "flipBlink": flip(&s.Caret.Steady), "caretChipW": caretChipW,
		"labelW": formLabelW, "fieldW": formFieldW, "notesH": notesH, "searchHint": searchHint, "remark": notesRemark,
	})
}

var textTemplate = markup.Must("text.vml", textSource)

const textSource = `
<scroll state="page" pad="{{gutter}}" gap="{{gutter}}">
  <caption>Text fields</caption>
  <form-row label="Name" label-width="{{labelW}}">
    <field state="name" w="{{fieldW}}" bg="sunken" text-menu="name"/>
  </form-row>
  <form-row label="Password" label-width="{{labelW}}">
    <field state="secret" w="{{fieldW}}" bg="sunken"/>
  </form-row>
  <form-row label="Search" label-width="{{labelW}}">
    <field state="search" w="{{fieldW}}" bg="sunken" placeholder="{{searchHint}}" text-menu="search"/>
  </form-row>
  <caption>Text area</caption>
  <area state="notes" h="{{notesH}}" bg="sunken" control="field"/>
  <text small muted>{{remark}}</text>
  <caption>Caret</caption>
  <form-row label="Shape" label-width="{{labelW}}">
    <row gap="{{tightGap}}">
      <segment-chip for="entry in carets" selected="{{entry.Selected}}" w="{{caretChipW}}" on-click="entry.Pick">{{entry.Label}}</segment-chip>
    </row>
  </form-row>
  <check-row on="{{blink}}" on-click="flipBlink" align-items="center"><text>Blink</text></check-row>
</scroll>`
