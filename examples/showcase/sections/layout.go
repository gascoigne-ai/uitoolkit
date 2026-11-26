package sections

import (
	"cmp"
	"fmt"
	"strconv"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	"github.com/gascoigne-ai/uitoolkit/render"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// Fixed sizes and counts of the layout section.
const (
	stackH               = 72
	tagW                 = 80
	tagCount             = 24
	boxPanelH            = 150
	boxW, boxH           = 240, 80
	boxMinW, boxMinH     = 160, 60
	boxMaxW, boxMaxH     = 600, 130
	meterW, meterH       = 120, 32
	meterBars, meterFull = 10, 100
)

// Texts of the layout section. linesSample has line breaks, which only a
// bound value keeps in a template: literal template text joins them.
const (
	stackHint   = "A stack layers its children"
	stackPill   = "Floating"
	linesSample = "Lines draws one text per line,\nso a short note keeps its breaks\nwithout a paragraph."
)

// meterB is a level meter: an element of the showcase's own, made with
// view.Custom. It draws and takes no input.
type meterB struct {
	Base
	level int // 0 to meterFull
}

func (meterB) Measure(*El, *core.Ctx, int) (int, int) { return meterW, meterH }

// Draw paints rising bars, lit up to the level.
func (m meterB) Draw(element *El, ctx *core.Ctx, dst render.Canvas) {
	rect := element.Rect()
	lit := (m.level*meterBars + meterFull/2) / meterFull
	step := rect.W / meterBars
	for bar := range meterBars {
		height := rect.H * (bar + 1) / meterBars
		token := core.Rule
		if bar < lit {
			token = core.Accent
		}
		ctx.Fill(dst, core.Rect{X: rect.X + bar*step, Y: rect.Y + rect.H - height, W: step - core.NearGap, H: height}, ctx.Color(token))
	}
}

// Describe puts the level in view.Dump.
func (m meterB) Describe() string { return "meter=" + strconv.Itoa(m.level) }

// meter shows the volume of the toggles section.
func (s *Showcase) meter() Node { return Custom(meterB{level: s.widgets.volume}) }

// boxSize is the size of the resizable box: the one the user gave it, else
// the default.
func (s *Showcase) boxSize() (width, height int) {
	return cmp.Or(s.widgets.box.W, boxW), cmp.Or(s.widgets.box.H, boxH)
}

// resizableBox is a Go node in both forms: markup has no element for
// widget.Resizable. The parent must center it on both axes, since it grows by
// twice the movement of the pointer.
func (s *Showcase) resizableBox() Node {
	width, height := s.boxSize()
	return Resizable(&s.widgets.box, boxMinW, boxMinH, boxMaxW, boxMaxH,
		W(width), H(height), Bg(core.Panel), Border(core.Border), JustifyContent(JustifyCenter),
		Text(fmt.Sprintf("%d × %d", width, height), TextCenter()),
		Text("Drag an edge", Small(), Muted(), TextCenter()),
	)
}

func (s *Showcase) tagChoices() []choice {
	choices := make([]choice, tagCount)
	for index := range choices {
		choices[index] = choice{Label: fmt.Sprintf("Tag %d", index+1), Selected: s.widgets.tag == index, Pick: func() { s.widgets.tag = index }}
	}
	return choices
}

func layoutGo(s *Showcase, _ *core.Ctx) Node {
	rule := func() Node { return VRuleInset(core.TightGap, HOf(core.RowOf)) }
	return Scroll(&s.widgets.layoutPage, Pad(core.Gutter), Gap(core.Gutter),
		Caption("Stack"),
		Stack(H(stackH), Bg(core.Sunken), Control("table"),
			Text(stackHint, Muted(), TextCenter()),
			Row(FloatTR(core.NearGap, core.NearGap), HOf(core.SlimOf), PadX(core.LabelPad), Bg(core.InfoSoft), Control("pill"),
				Text(stackPill, Small(), Bold(), Color(core.Info))),
		),
		Caption("Horizontal scroll"),
		HScroll(&s.widgets.tagStrip, HOf(core.RowOf), Gap(core.TightGap),
			For(s.tagChoices(), func(_ int, entry choice) Node {
				return SegmentChip(entry.Label, entry.Selected, W(tagW), OnClick(entry.Pick))
			}),
		),
		Caption("Inset rules"),
		Row(HOf(core.RowOf), Gap(core.LooseGap), AlignItems(AlignCenter),
			Text("Bold", Bold()), rule(), Text("Italic"), rule(), Text("Code", Mono()),
		),
		Caption("Lines and a custom element"),
		Row(Gap(core.Gutter), AlignItems(AlignStart),
			Lines(linesSample, core.Muted),
			s.meter(),
		),
		Caption("Resizable"),
		Row(H(boxPanelH), Bg(core.Sunken), Control("table"), AlignItems(AlignCenter), JustifyContent(JustifyCenter),
			s.resizableBox(),
		),
	)
}

func layoutMarkup(s *Showcase, ctx *core.Ctx) Node {
	return fill(ctx, layoutTemplate, markup.Scope{
		"page": &s.widgets.layoutPage, "stackH": stackH, "stackHint": stackHint, "pill": stackPill,
		"strip": &s.widgets.tagStrip, "tags": s.tagChoices(), "tagW": tagW,
		"sample": linesSample, "meter": s.meter(), "panelH": boxPanelH, "resizable": s.resizableBox(),
	})
}

var layoutTemplate = markup.Must("layout.vml", layoutSource)

const layoutSource = `
<scroll state="page" pad="{{gutter}}" gap="{{gutter}}">
  <caption>Stack</caption>
  <stack h="{{stackH}}" bg="sunken" control="table">
    <text muted text-center>{{stackHint}}</text>
    <row float-tr="{{nearGap}} {{nearGap}}" h="{{slimH}}" pad-x="{{labelPad}}" bg="info-soft" control="pill">
      <text small bold color="info">{{pill}}</text>
    </row>
  </stack>
  <caption>Horizontal scroll</caption>
  <hscroll state="strip" h="{{rowH}}" gap="{{tightGap}}">
    <segment-chip for="entry in tags" selected="{{entry.Selected}}" w="{{tagW}}" on-click="entry.Pick">{{entry.Label}}</segment-chip>
  </hscroll>
  <caption>Inset rules</caption>
  <row h="{{rowH}}" gap="{{looseGap}}" align-items="center">
    <text bold>Bold</text>
    <vrule-inset inset="{{tightGap}}" h="{{rowH}}"/>
    <text>Italic</text>
    <vrule-inset inset="{{tightGap}}" h="{{rowH}}"/>
    <text mono>Code</text>
  </row>
  <caption>Lines and a custom element</caption>
  <row gap="{{gutter}}" align-items="start">
    <lines color="muted">{{sample}}</lines>
    <slot name="meter"/>
  </row>
  <caption>Resizable</caption>
  <row h="{{panelH}}" bg="sunken" control="table" align-items="center" justify-content="center">
    <slot name="resizable"/>
  </row>
</scroll>`
