package sections

import (
	"fmt"
	"image"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// Sizes and counts of the lists section.
const (
	listW         = 260
	listMinW      = 160
	detailMinW    = 240
	listEntries   = 60
	posedEntry    = 7
	posedListY    = 112
	posedStripTab = 2
)

var stripTabs = []TabItem{
	{Label: "Overview"}, {Label: "Parameters", Badge: 3}, {Label: "Headers", Badge: 12}, {Label: "Authorization"},
	{Label: "Body"}, {Label: "Settings"}, {Label: "Examples"}, {Label: "History"}, {Label: "Documentation"},
	{Label: "Variables"}, {Label: "Scripts"}, {Label: "Cookies"}, {Label: "Certificates"},
}

var detailTabs = []TabItem{{Label: "Summary"}, {Label: "Activity", Badge: 2}, {Label: "Notes"}}

var listsPoses = []Pose{{Apply: func(s *Showcase, _ *core.Ctx) image.Point {
	s.widgets.picked, s.widgets.list.Y, s.widgets.tab = posedEntry, posedListY, posedStripTab
	return OffCanvas
}}}

// entry is one row of the list as a view reads it.
type entry struct {
	Name, Meta string
	Selected   bool
	Pick       func()
}

func (s *Showcase) entries() []entry {
	entries := make([]entry, listEntries)
	for index := range entries {
		entries[index] = entry{
			Name:     fmt.Sprintf("Item %d", index+1),
			Meta:     fmt.Sprintf("%d KB", 3+index*7%90),
			Selected: index == s.widgets.picked,
			Pick:     func() { s.widgets.picked = index },
		}
	}
	return entries
}

func (s *Showcase) detailTitle() string { return fmt.Sprintf("Item %d", s.widgets.picked+1) }

func (s *Showcase) detailTable() TableSpec {
	return TableSpec{
		Columns: []string{"Property", "Value", "Changed"},
		Weights: []float64{2, 3, 2},
		Rows: []TableRow{
			{Cells: []string{"Owner", "Ada", "today"}},
			{Cells: []string{"Kind", "Document", "last week"}},
			{Cells: []string{"Size", s.entries()[s.widgets.picked].Meta, "today"}},
			{Cells: []string{"Shared", "No", "never"}, Muted: true},
		},
		State: &s.widgets.table,
		Zebra: true,
	}
}

func listsGo(s *Showcase, _ *core.Ctx) Node {
	state := &s.widgets
	return Col(Grow(1),
		TabBar(ScrollTabs(&state.tabStrip, stripTabs, state.tab, OnSelect(func(index int) { state.tab = index }))),
		Split(&state.split, listW, Grow(1), MinFirst(listMinW), MinSecond(detailMinW),
			Scroll(&state.list, Bg(core.Sunken),
				For(s.entries(), func(_ int, row entry) Node {
					return ListRow(row.Selected, HOf(core.RowOf), PadX(core.Gutter), Gap(core.LooseGap), AlignItems(AlignCenter),
						OnClick(row.Pick), Text(row.Name, Grow(1)), Text(row.Meta, Small(), Muted()))
				}),
			),
			Scroll(&state.detail, Pad(core.Gutter), Gap(core.Gutter),
				Text(s.detailTitle(), Bold()),
				Text("Drag the divider to resize the panes, or a column edge in the table.", Small(), Muted(), Ellipsis()),
				Table(s.detailTable()),
				TabBar(Tabs(detailTabs, state.detailTab, OnSelect(func(index int) { state.detailTab = index }))),
				Text(detailTabs[state.detailTab].Label+" of "+s.detailTitle(), Muted()),
			),
		),
	)
}

func listsMarkup(s *Showcase, ctx *core.Ctx) Node {
	state := &s.widgets
	return fill(ctx, listsTemplate, markup.Scope{
		"strip": &state.tabStrip, "stripTabs": stripTabs, "tab": state.tab,
		"pickTab": func(index int) { state.tab = index },
		"split":   &state.split, "listW": listW, "listMinW": listMinW, "detailMinW": detailMinW,
		"list": &state.list, "entries": s.entries(),
		"detail": &state.detail, "title": s.detailTitle(), "table": s.detailTable(),
		"detailTabs": detailTabs, "detailTab": state.detailTab,
		"pickDetailTab": func(index int) { state.detailTab = index },
		"shown":         detailTabs[state.detailTab].Label,
	})
}

var listsTemplate = markup.Must("lists.vml", listsSource)

const listsSource = `
<col grow="1">
  <tab-bar>
    <scroll-tabs state="strip" items="stripTabs" active="{{tab}}" on-select="pickTab"/>
  </tab-bar>
  <split state="split" default="{{listW}}" grow="1" min-first="{{listMinW}}" min-second="{{detailMinW}}">
    <scroll state="list" bg="sunken">
      <list-row for="row in entries" selected="{{row.Selected}}" h="{{rowH}}" pad-x="{{gutter}}" gap="{{looseGap}}" align-items="center" on-click="row.Pick">
        <text grow="1">{{row.Name}}</text>
        <text small muted>{{row.Meta}}</text>
      </list-row>
    </scroll>
    <scroll state="detail" pad="{{gutter}}" gap="{{gutter}}">
      <text bold>{{title}}</text>
      <text small muted ellipsis>Drag the divider to resize the panes, or a column edge in the table.</text>
      <table spec="table"/>
      <tab-bar>
        <tabs items="detailTabs" active="{{detailTab}}" on-select="pickDetailTab"/>
      </tab-bar>
      <text muted>{{shown}} of {{title}}</text>
    </scroll>
  </split>
</col>`
