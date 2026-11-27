package sections

import (
	"cmp"
	"fmt"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// Sizes of the resizable dialog; sizedW by sizedH is its default.
const (
	sizedW, sizedH       = 480, 320
	sizedMinW, sizedMinH = 360, 240
	sizedMaxW, sizedMaxH = 800, 560
	sizedNoteCount       = 16
)

// Texts of the dialogs and the toasts.
const (
	dialogRemark    = "The dialog is modal. Drag it by its title bar."
	panelTitle      = "Preferences"
	panelSubtitle   = "Editor"
	panelSearchHint = "Search preferences"
	sizedTitle      = "Resizable dialog"
	infoToast       = "Settings saved"
	errorToast      = "Export failed"
)

// dialogKind is which dialog is open.
type dialogKind int

// The dialogs of the showcase.
const (
	noDialog      dialogKind = iota
	messageDialog            // a widget.MessageDialog
	panelDialog              // a widget.Dialog put together from its parts
	sizedDialog              // a widget.ResizableDialog
)

// dialogView is one dialog built with the Go DSL and from markup.
type dialogView struct{ Go, Markup builder }

// dialogViews are the dialogs by kind; noDialog has none. A new dialog is a
// new kind and a new row.
var dialogViews = [...]dialogView{
	messageDialog: {Go: messageDialogGo, Markup: messageDialogMarkup},
	panelDialog:   {Go: panelDialogGo, Markup: panelDialogMarkup},
	sizedDialog:   {Go: sizedDialogGo, Markup: sizedDialogMarkup},
}

// showDialog opens the dialog of one kind, centered.
func (s *Showcase) showDialog(kind dialogKind) {
	s.widgets.dialogMove = MoveState{}
	s.dialog = kind
}

func (s *Showcase) dismissDialog() { s.dialog = noDialog }

// openDialog asks before the pretend delete.
func (s *Showcase) openDialog() {
	s.widgets.confirm = NewTextField("")
	s.widgets.confirm.Focused = true
	answer := func(said string) func() {
		return func() {
			s.widgets.lastAnswer = said
			s.dismissDialog()
		}
	}
	s.widgets.message = DialogState{
		Title: "Delete project?",
		Body:  "This cannot be undone.\nType the project name to confirm.",
		Input: &s.widgets.confirm,
		Buttons: []DialogButton{
			{Label: "Cancel", Run: answer("cancelled")},
			{Label: "Delete", Primary: true, Run: answer("deleted")},
		},
	}
	s.showDialog(messageDialog)
}

// openPanel shows the dialog made of parts.
func (s *Showcase) openPanel() { s.showDialog(panelDialog) }

// openSized shows the resizable dialog.
func (s *Showcase) openSized() { s.showDialog(sizedDialog) }

func messageDialogGo(s *Showcase, _ *core.Ctx) Node {
	return MessageDialog(&s.widgets.message, Text(dialogRemark, Small(), Muted()), OnDismiss(s.dismissDialog), Moved(&s.widgets.dialogMove))
}

func messageDialogMarkup(s *Showcase, ctx *core.Ctx) Node {
	return fill(ctx, messageDialogTemplate, markup.Scope{
		"dialog": &s.widgets.message, "remark": dialogRemark, "dismiss": s.dismissDialog, "move": &s.widgets.dialogMove,
	})
}

func (s *Showcase) panelScopeChoices() []choice {
	var choices []choice
	for index, label := range []string{"All", "Modified"} {
		choices = append(choices, choice{Label: label, Selected: s.widgets.panelScope == index, Pick: func() { s.widgets.panelScope = index }})
	}
	return choices
}

func panelDialogGo(s *Showcase, _ *core.Ctx) Node {
	state := &s.widgets
	return Dialog(WOf(core.DialogWOf), OnDismiss(s.dismissDialog), Moved(&state.dialogMove),
		DialogTitle(panelTitle, OnDismiss(s.dismissDialog), DialogSubtitle(panelSubtitle)),
		DialogToolbar(
			Field(&state.panelSearch, HOf(core.RowOf), Grow(1), Bg(core.Sunken), Placeholder(panelSearchHint)),
			For(s.panelScopeChoices(), func(_ int, entry choice) Node {
				return SegmentChip(entry.Label, entry.Selected, HOf(core.RowOf), OnClick(entry.Pick))
			}),
		),
		Col(PadOf(core.DialogPadOf), GapOf(core.GutterOf),
			CheckRow(state.wrapLines, OnClick(flip(&state.wrapLines)), AlignItems(AlignCenter), Text("Wrap long lines")),
			CheckRow(state.showHidden, OnClick(flip(&state.showHidden)), AlignItems(AlignCenter), Text("Show hidden files")),
			CheckRow(state.autosave, OnClick(flip(&state.autosave)), AlignItems(AlignCenter), Text("Save automatically")),
		),
	)
}

func panelDialogMarkup(s *Showcase, ctx *core.Ctx) Node {
	state := &s.widgets
	return fill(ctx, panelDialogTemplate, markup.Scope{
		"width": ctx.Sizes.DialogW, "dismiss": s.dismissDialog, "move": &state.dialogMove,
		"title": panelTitle, "subtitle": panelSubtitle, "search": &state.panelSearch, "searchHint": panelSearchHint,
		"scopes":    s.panelScopeChoices(),
		"wrapLines": state.wrapLines, "flipWrap": flip(&state.wrapLines),
		"showHidden": state.showHidden, "flipHidden": flip(&state.showHidden),
		"autosave": state.autosave, "flipAutosave": flip(&state.autosave),
	})
}

// sizedSize is the size of the resizable dialog: the one the user gave it,
// else the default.
func (s *Showcase) sizedSize() (width, height int) {
	return cmp.Or(s.widgets.sized.W, sizedW), cmp.Or(s.widgets.sized.H, sizedH)
}

func (s *Showcase) sizedText() string {
	width, height := s.sizedSize()
	text := fmt.Sprintf("Size: %d × %d", width, height)
	if s.widgets.sized.W == 0 {
		text += " (default)"
	}
	return text
}

func sizedNotes() []string {
	notes := []string{"Drag an edge or a corner.", "A double click on an edge restores the default size."}
	for row := range sizedNoteCount {
		notes = append(notes, fmt.Sprintf("Row %d of a body that scrolls", row+1))
	}
	return notes
}

func sizedDialogGo(s *Showcase, _ *core.Ctx) Node {
	state := &s.widgets
	width, height := s.sizedSize()
	return ResizableDialog(&state.sized, sizedMinW, sizedMinH, sizedMaxW, sizedMaxH,
		W(width), H(height), OnDismiss(s.dismissDialog), Moved(&state.dialogMove),
		DialogTitle(sizedTitle, OnDismiss(s.dismissDialog)),
		Scroll(&state.sizedBody, Grow(1), PadOf(core.DialogPadOf), GapOf(core.LooseGapOf),
			For(sizedNotes(), func(_ int, note string) Node { return Text(note, Muted()) }),
		),
	)
}

func sizedDialogMarkup(s *Showcase, ctx *core.Ctx) Node {
	width, height := s.sizedSize()
	return fill(ctx, sizedDialogTemplate, markup.Scope{
		"sized": &s.widgets.sized, "minW": sizedMinW, "minH": sizedMinH, "maxW": sizedMaxW, "maxH": sizedMaxH,
		"width": width, "height": height, "dismiss": s.dismissDialog, "move": &s.widgets.dialogMove,
		"title": sizedTitle, "body": &s.widgets.sizedBody, "notes": sizedNotes(),
	})
}

// postToast shows a message in the corner of the window.
func (s *Showcase) postToast(ctx *core.Ctx, text string, kind ToastKind) {
	s.widgets.toasts.Show(text, kind, s.now(), ctx.Feel())
}

func (s *Showcase) toastCount() string { return fmt.Sprintf("%d shown", len(s.widgets.toasts)) }

func toastsGo(s *Showcase, _ *core.Ctx, width, height int) Node {
	return Toasts(&s.widgets.toasts, width, height, 0)
}

func toastsMarkup(s *Showcase, ctx *core.Ctx, width, height int) Node {
	return fill(ctx, toastsTemplate, markup.Scope{"stack": &s.widgets.toasts, "width": width, "height": height})
}

var (
	messageDialogTemplate = markup.Must("message-dialog.vml", messageDialogSource)
	panelDialogTemplate   = markup.Must("panel-dialog.vml", panelDialogSource)
	sizedDialogTemplate   = markup.Must("sized-dialog.vml", sizedDialogSource)
	toastsTemplate        = markup.Must("toasts.vml", toastsSource)
)

const messageDialogSource = `
<message-dialog state="dialog" on-dismiss="dismiss" moved="move">
  <text small muted>{{remark}}</text>
</message-dialog>`

const panelDialogSource = `
<dialog w="{{width}}" on-dismiss="dismiss" moved="move">
  <dialog-title title="{{title}}" on-dismiss="dismiss">
    <dialog-subtitle>{{subtitle}}</dialog-subtitle>
  </dialog-title>
  <dialog-toolbar>
    <field state="search" h="{{rowH}}" grow="1" bg="sunken" placeholder="{{searchHint}}"/>
    <segment-chip for="entry in scopes" selected="{{entry.Selected}}" h="{{rowH}}" on-click="entry.Pick">{{entry.Label}}</segment-chip>
  </dialog-toolbar>
  <col pad="{{dialogPad}}" gap="{{gutter}}">
    <check-row on="{{wrapLines}}" on-click="flipWrap" align-items="center"><text>Wrap long lines</text></check-row>
    <check-row on="{{showHidden}}" on-click="flipHidden" align-items="center"><text>Show hidden files</text></check-row>
    <check-row on="{{autosave}}" on-click="flipAutosave" align-items="center"><text>Save automatically</text></check-row>
  </col>
</dialog>`

const sizedDialogSource = `
<resizable-dialog state="sized" min-width="{{minW}}" min-height="{{minH}}" max-width="{{maxW}}" max-height="{{maxH}}"
    w="{{width}}" h="{{height}}" on-dismiss="dismiss" moved="move">
  <dialog-title title="{{title}}" on-dismiss="dismiss"/>
  <scroll state="body" grow="1" pad="{{dialogPad}}" gap="{{looseGap}}">
    <text for="note in notes" muted>{{note}}</text>
  </scroll>
</resizable-dialog>`

const toastsSource = `<toasts stack="stack" width="{{width}}" height="{{height}}" bottom-inset="0"/>`
