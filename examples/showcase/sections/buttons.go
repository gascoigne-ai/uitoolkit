package sections

import (
	"fmt"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// Fixed widths of the buttons section.
const (
	buttonW     = 96
	chipW       = 80
	modeChipW   = 40
	chipHeading = "Segmented chips"
)

// choice is one chip of a segmented control as a view reads it.
type choice struct {
	Label, Icon, Tip string
	Selected         bool
	Pick             func()
}

func (s *Showcase) sizeChoices() []choice {
	var choices []choice
	for index, label := range []string{"Small", "Medium", "Large"} {
		choices = append(choices, choice{Label: label, Selected: s.widgets.chip == index, Pick: func() { s.widgets.chip = index }})
	}
	return choices
}

func (s *Showcase) viewChoices() []choice {
	choices := []choice{{Icon: "folder", Tip: "Files"}, {Icon: "history", Tip: "History"}, {Icon: "search", Tip: "Search"}}
	for index := range choices {
		choices[index].Selected = s.widgets.viewChosen == index
		choices[index].Pick = func() { s.widgets.viewChosen = index }
	}
	return choices
}

// pill is a status label as a view reads it.
type pill struct {
	Label       string
	Color, Fill core.Token
}

var pills = []pill{
	{"200 OK", core.OK, core.OKSoft},
	{"Slow", core.Warn, core.WarnSoft},
	{"Failed", core.Error, core.ErrorSoft},
	{"Info", core.Info, core.InfoSoft},
}

func (s *Showcase) clickedText() string { return fmt.Sprintf("Clicked %d times", s.widgets.clicks) }

func buttonsGo(s *Showcase, _ *core.Ctx) Node {
	count := func() { s.widgets.clicks++ }
	return Scroll(&s.widgets.buttonsPage, Pad(core.Gutter), Gap(core.Gutter),
		Caption("Buttons"),
		Row(HOf(core.RowOf), Gap(core.LooseGap),
			Button("Button", W(buttonW), OnClick(count)),
			PrimaryButton("Primary", W(buttonW), OnClick(count)),
			LinkButton("Link button", OnClick(count)),
			IconButton(Icon("plus", Grow(1)), WOf(core.IconBtnOf), Tooltip("Add"), OnClick(count)),
			IconButton(Icon("trash", Grow(1)), WOf(core.IconBtnOf), Tooltip("Delete"), OnClick(count)),
		),
		Text(s.clickedText(), Muted()),
		Caption(chipHeading),
		Row(HOf(core.RowOf), Gap(core.TightGap),
			For(s.sizeChoices(), func(_ int, entry choice) Node {
				return SegmentChip(entry.Label, entry.Selected, W(chipW), OnClick(entry.Pick))
			}),
		),
		Row(HOf(core.CompactOf), Gap(core.TightGap),
			For(s.viewChoices(), func(_ int, entry choice) Node {
				return SegmentIconChip(entry.Icon, entry.Selected, W(modeChipW), Tooltip(entry.Tip), OnClick(entry.Pick))
			}),
		),
		Caption("Status pills"),
		Row(HOf(core.SlimOf), Gap(core.LooseGap),
			For(pills, func(_ int, entry pill) Node {
				return Row(PadX(core.LabelPad), Bg(entry.Fill), Control("pill"),
					Text(entry.Label, Small(), Bold(), Color(entry.Color)))
			}),
		),
	)
}

func buttonsMarkup(s *Showcase, ctx *core.Ctx) Node {
	return fill(ctx, buttonsTemplate, markup.Scope{
		"page": &s.widgets.buttonsPage, "count": func() { s.widgets.clicks++ }, "clicked": s.clickedText(),
		"sizes": s.sizeChoices(), "views": s.viewChoices(), "pills": pills,
		"buttonW": buttonW, "chipW": chipW, "modeChipW": modeChipW, "chipHeading": chipHeading,
	})
}

var buttonsTemplate = markup.Must("buttons.vml", buttonsSource)

const buttonsSource = `
<scroll state="page" pad="{{gutter}}" gap="{{gutter}}">
  <caption>Buttons</caption>
  <row h="{{rowH}}" gap="{{looseGap}}">
    <button w="{{buttonW}}" on-click="count">Button</button>
    <primary-button w="{{buttonW}}" on-click="count">Primary</primary-button>
    <link-button on-click="count">Link button</link-button>
    <icon-button icon="plus" w="{{iconBtnW}}" tooltip="Add" on-click="count"/>
    <icon-button icon="trash" w="{{iconBtnW}}" tooltip="Delete" on-click="count"/>
  </row>
  <text muted>{{clicked}}</text>
  <caption>{{chipHeading}}</caption>
  <row h="{{rowH}}" gap="{{tightGap}}">
    <segment-chip for="entry in sizes" selected="{{entry.Selected}}" w="{{chipW}}" on-click="entry.Pick">{{entry.Label}}</segment-chip>
  </row>
  <row h="{{compactH}}" gap="{{tightGap}}">
    <segment-icon-chip for="entry in views" icon="{{entry.Icon}}" selected="{{entry.Selected}}" w="{{modeChipW}}" tooltip="{{entry.Tip}}" on-click="entry.Pick"/>
  </row>
  <caption>Status pills</caption>
  <row h="{{slimH}}" gap="{{looseGap}}">
    <row for="entry in pills" pad-x="{{labelPad}}" bg="{{entry.Fill}}" control="pill">
      <text small bold color="{{entry.Color}}">{{entry.Label}}</text>
    </row>
  </row>
</scroll>`
