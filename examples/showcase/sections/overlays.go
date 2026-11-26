package sections

import (
	"image"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// Sizes and texts of the overlays section.
const (
	dropPanelH  = 120
	overlayBtnW = 140
	settingsTip = "Settings (Ctrl+,)"
	hoverTip    = "A tooltip shows after the pointer rests"
	dropHint    = "Right-click here for a menu"
)

var overlaysPoses = []Pose{
	{Name: "menu", Apply: func(s *Showcase, ctx *core.Ctx) image.Point {
		if s.widgets.menuButton == nil {
			return OffCanvas
		}
		anchor := s.widgets.menuButton.Rect()
		s.openMenu(ctx, anchor.X, anchor.Y+anchor.H+ctx.Sizes.NearGap)
		return OffCanvas
	}},
	{Name: "tooltip", Apply: func(s *Showcase, _ *core.Ctx) image.Point {
		if s.widgets.tooltipAnchor == nil {
			return OffCanvas
		}
		anchor := s.widgets.tooltipAnchor.Rect()
		return image.Pt(anchor.X+anchor.W/2, anchor.Y+anchor.H/2)
	}},
	{Name: "dialog", Apply: func(s *Showcase, _ *core.Ctx) image.Point {
		s.openDialog()
		return OffCanvas
	}},
	{Name: "panel", Apply: func(s *Showcase, _ *core.Ctx) image.Point {
		s.openPanel()
		return OffCanvas
	}},
	{Name: "resizable", Apply: func(s *Showcase, _ *core.Ctx) image.Point {
		s.openSized()
		return OffCanvas
	}},
	{Name: "toasts", Apply: func(s *Showcase, _ *core.Ctx) image.Point {
		s.postToast(infoToast, ToastInfo)
		s.postToast(errorToast, ToastError)
		return OffCanvas
	}},
}

// openMenu opens the demo popup menu at posX, posY. Markup can show a menu
// but has no way to open one: widget.OpenMenu is Go only.
func (s *Showcase) openMenu(ctx *core.Ctx, posX, posY int) {
	pick := func(label string) func() { return func() { s.widgets.lastPick = label } }
	OpenMenu(ctx, posX, posY, []MenuItem{
		{Label: "New file", Icon: "plus", Shortcut: "Ctrl+N", Run: pick("New file")},
		{Label: "Copy path", Icon: "copy", Shortcut: "Ctrl+C", Run: pick("Copy path")},
		{Label: "Rename", Icon: "pencil", Run: pick("Rename")},
		{Separator: true},
		{Label: "Export", Icon: "save", Hint: "Writes the project\nto a single file", Run: pick("Export")},
		{Label: "Unavailable", Disabled: true},
		{Separator: true},
		{Label: "Delete...", Icon: "trash", Run: s.openDialog},
	})
}

func menuGo(_ *Showcase, ctx *core.Ctx) Node { return PopupMenuLayer(ctx) }

// menuMarkup is the open popup menu as a <menu>. A pick does what
// widget.PopupMenuLayer does: close, then run the row.
func menuMarkup(_ *Showcase, ctx *core.Ctx) Node {
	state := PopupMenu(ctx)
	return fill(ctx, menuTemplate, markup.Scope{
		"ctx": ctx, "menu": state, "dismiss": func() { CloseMenu(ctx) },
		"pick": func(index int) {
			run := state.Items[index].Run
			CloseMenu(ctx)
			if run != nil {
				run()
			}
		},
	})
}

func overlaysGo(s *Showcase, ctx *core.Ctx) Node {
	state := &s.widgets
	state.menuButton = PrimaryButton("Open menu", W(overlayBtnW), OnClickAt(func(posX, posY int) { s.openMenu(ctx, posX, posY) }))
	state.tooltipAnchor = Button("Hover me", W(overlayBtnW), Tooltip(hoverTip))
	return Scroll(&state.overlaysPage, PadOf(core.GutterOf), GapOf(core.GutterOf),
		Caption("Popup menus"),
		Row(HOf(core.RowOf), GapOf(core.LooseGapOf), state.menuButton, Text("Last pick: "+state.lastPick, Muted())),
		Row(H(dropPanelH), Bg(core.Sunken), Control("table"), OnRightClick(func(posX, posY int) { s.openMenu(ctx, posX, posY) }),
			Text(dropHint, Muted(), TextCenter(), Grow(1)),
		),
		Caption("Tooltips"),
		Row(HOf(core.RowOf), GapOf(core.LooseGapOf),
			state.tooltipAnchor,
			IconButton(Icon("gear", Grow(1)), WOf(core.IconBtnOf), Tooltip(settingsTip)),
		),
		Caption("Modal dialog"),
		Row(HOf(core.RowOf), GapOf(core.LooseGapOf),
			Button("Open dialog", W(overlayBtnW), OnClick(s.openDialog)),
			Text("Last answer: "+state.lastAnswer, Muted()),
		),
		Caption("Dialog parts"),
		Row(HOf(core.RowOf), Button("Open preferences", W(overlayBtnW), OnClick(s.openPanel))),
		Caption("Resizable dialog"),
		Row(HOf(core.RowOf), GapOf(core.LooseGapOf),
			Button("Open resizable", W(overlayBtnW), OnClick(s.openSized)),
			Text(s.sizedText(), Muted()),
		),
		Caption("Toasts"),
		Row(HOf(core.RowOf), GapOf(core.LooseGapOf),
			Button("Info toast", W(overlayBtnW), OnClick(func() { s.postToast(infoToast, ToastInfo) })),
			Button("Error toast", W(overlayBtnW), OnClick(func() { s.postToast(errorToast, ToastError) })),
			Text(s.toastCount(), Muted()),
		),
	)
}

func overlaysMarkup(s *Showcase, ctx *core.Ctx) Node {
	state := &s.widgets
	open := func(posX, posY int) { s.openMenu(ctx, posX, posY) }
	return fill(ctx, overlaysTemplate, markup.Scope{
		"page": &state.overlaysPage, "buttonW": overlayBtnW, "panelH": dropPanelH, "openMenu": open,
		"lastPick": state.lastPick, "lastAnswer": state.lastAnswer, "openDialog": s.openDialog,
		"dropHint": dropHint, "hoverTip": hoverTip, "settingsTip": settingsTip,
		"openPanel": s.openPanel, "openSized": s.openSized, "sizedText": s.sizedText(), "toastCount": s.toastCount(),
		"postInfo":  func() { s.postToast(infoToast, ToastInfo) },
		"postError": func() { s.postToast(errorToast, ToastError) },
	})
}

var (
	overlaysTemplate = markup.Must("overlays.vml", overlaysSource)
	menuTemplate     = markup.Must("menu.vml", menuSource)
)

const overlaysSource = `
<scroll state="page" pad="{{gutter}}" gap="{{gutter}}">
  <caption>Popup menus</caption>
  <row h="{{rowH}}" gap="{{looseGap}}">
    <primary-button w="{{buttonW}}" on-click-at="openMenu">Open menu</primary-button>
    <text muted>Last pick: {{lastPick}}</text>
  </row>
  <row h="{{panelH}}" bg="sunken" control="table" on-right-click="openMenu">
    <text muted text-center grow="1">{{dropHint}}</text>
  </row>
  <caption>Tooltips</caption>
  <row h="{{rowH}}" gap="{{looseGap}}">
    <button w="{{buttonW}}" tooltip="{{hoverTip}}">Hover me</button>
    <icon-button icon="gear" w="{{iconBtnW}}" tooltip="{{settingsTip}}"/>
  </row>
  <caption>Modal dialog</caption>
  <row h="{{rowH}}" gap="{{looseGap}}">
    <button w="{{buttonW}}" on-click="openDialog">Open dialog</button>
    <text muted>Last answer: {{lastAnswer}}</text>
  </row>
  <caption>Dialog parts</caption>
  <row h="{{rowH}}">
    <button w="{{buttonW}}" on-click="openPanel">Open preferences</button>
  </row>
  <caption>Resizable dialog</caption>
  <row h="{{rowH}}" gap="{{looseGap}}">
    <button w="{{buttonW}}" on-click="openSized">Open resizable</button>
    <text muted>{{sizedText}}</text>
  </row>
  <caption>Toasts</caption>
  <row h="{{rowH}}" gap="{{looseGap}}">
    <button w="{{buttonW}}" on-click="postInfo">Info toast</button>
    <button w="{{buttonW}}" on-click="postError">Error toast</button>
    <text muted>{{toastCount}}</text>
  </row>
</scroll>`

const menuSource = `<menu ctx="ctx" state="menu" on-select="pick" on-dismiss="dismiss"/>`
