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
	dropPanelH   = 120
	overlayBtnW  = 140
	settingsTip  = "Settings (Ctrl+,)"
	hoverTip     = "A tooltip shows after the pointer rests"
	dropHint     = "Right-click here for a menu"
	dialogRemark = "The dialog is modal. Drag it by its title bar."
)

var overlaysPoses = []Pose{
	{Name: "menu", Apply: func(s *Showcase, ctx *core.Ctx) image.Point {
		if s.widgets.menuButton == nil {
			return OffCanvas
		}
		anchor := s.widgets.menuButton.Rect()
		s.openMenu(ctx, anchor.X, anchor.Y+anchor.H+core.NearGap)
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

// openDialog asks before the pretend delete.
func (s *Showcase) openDialog() {
	s.widgets.confirm = NewTextField("")
	s.widgets.confirm.Focused = true
	s.widgets.dialogMove = MoveState{}
	answer := func(said string) func() {
		return func() {
			s.widgets.lastAnswer = said
			s.dialog = nil
		}
	}
	s.dialog = &DialogState{
		Title: "Delete project?",
		Body:  "This cannot be undone.\nType the project name to confirm.",
		Input: &s.widgets.confirm,
		Buttons: []DialogButton{
			{Label: "Cancel", Run: answer("cancelled")},
			{Label: "Delete", Primary: true, Run: answer("deleted")},
		},
	}
}

func (s *Showcase) dismissDialog() { s.dialog = nil }

func dialogGo(s *Showcase, _ *core.Ctx) Node {
	return MessageDialog(s.dialog, Text(dialogRemark, Small(), Muted()), OnDismiss(s.dismissDialog), Moved(&s.widgets.dialogMove))
}

func dialogMarkup(s *Showcase, ctx *core.Ctx) Node {
	return fill(ctx, dialogTemplate, markup.Scope{
		"dialog": s.dialog, "remark": dialogRemark, "dismiss": s.dismissDialog, "move": &s.widgets.dialogMove,
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
	return Scroll(&state.overlaysPage, Pad(core.Gutter), Gap(core.Gutter),
		Caption("Popup menus"),
		Row(HOf(core.RowOf), Gap(core.LooseGap), state.menuButton, Text("Last pick: "+state.lastPick, Muted())),
		Row(H(dropPanelH), Bg(core.Sunken), Control("table"), OnRightClick(func(posX, posY int) { s.openMenu(ctx, posX, posY) }),
			Text(dropHint, Muted(), TextCenter(), Grow(1)),
		),
		Caption("Tooltips"),
		Row(HOf(core.RowOf), Gap(core.LooseGap),
			state.tooltipAnchor,
			IconButton(Icon("gear", Grow(1)), WOf(core.IconBtnOf), Tooltip(settingsTip)),
		),
		Caption("Modal dialog"),
		Row(HOf(core.RowOf), Gap(core.LooseGap),
			Button("Open dialog", W(overlayBtnW), OnClick(s.openDialog)),
			Text("Last answer: "+state.lastAnswer, Muted()),
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
	})
}

var (
	overlaysTemplate = markup.Must("overlays.vml", overlaysSource)
	dialogTemplate   = markup.Must("dialog.vml", dialogSource)
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
</scroll>`

const dialogSource = `
<message-dialog state="dialog" on-dismiss="dismiss" moved="move">
  <text small muted>{{remark}}</text>
</message-dialog>`

const menuSource = `<menu ctx="ctx" state="menu" on-select="pick" on-dismiss="dismiss"/>`
