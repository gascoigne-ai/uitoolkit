package widget

import (
	"github.com/gascoigne-ai/uitoolkit/core"

	. "github.com/gascoigne-ai/uitoolkit/view"
)

// dialogLook is what makes a panel a dialog: a floating, modal menu-style
// panel in the middle of the window that owns the keyboard focus. It keeps
// Sizes.DialogMargin off the window's edges, is at most Sizes.DialogMaxH
// high and clips its content to the panel; a drag that leaves the panel goes
// on, since its owner reads Ctx.PointerX and PointerY, which no clip hides.
var dialogLook = []Item{
	Bg(core.Bg), Control("menu"), Shadow(), FloatCenter(), FocusScope(), Modal(), KeepOnOutsideClick(),
	MarginOf(core.DialogMarginOf), MaxHOf(core.DialogMaxHOf), Clip(),
}

// Dialog is a modal panel centered in the window over a dimmed scrim. Give
// it its size (W, H), layout options and content, and OnDismiss for what
// an outside click does. Mount it in the window's Stack while it is open.
//
// The panel stays dialogMargin off the window's edges, dragged or not, and
// is at most dialogMaxH high, a height you give it with H included; your own
// MaxH (MaxH(0) for no limit) or Margin replaces either. Content that does
// not fit is clipped to the panel and nothing scrolls it: put what may be
// long in a Scroll with Grow(1), inside containers that have Shrink().
//
// Dialogs are floating panels inside the one window, each with a title bar
// of its own (DialogTitle), because Ebitengine has one OS window per
// process. Their look is in .harness/style.md.
func Dialog(items ...Item) *El {
	return Frag(Row(Bg(core.Shadow)), Col(append(append([]Item{}, dialogLook...), items...)...))
}

// ResizableDialog is a Dialog the user can resize between the given
// bounds; state keeps the size across frames. It has the limits of a Dialog:
// state.H also stops at dialogMaxH, or at your own MaxH.
func ResizableDialog(state *ResizeState, minW, minH, maxW, maxH int, items ...Item) *El {
	return Frag(Row(Bg(core.Shadow)), Resizable(state, minW, minH, maxW, maxH, append(append([]Item{}, dialogLook...), items...)...))
}

// DialogTitle is the title bar of a dialog, drawn like the window's header:
// the hero banner (so it differs from the body in every theme) with a Border
// line under it. It holds the title, any items you give (a DialogSubtitle),
// free room and a close button at the end. The user drags the dialog by the
// bar (the room between the title and the buttons); the grab cursor shows
// only while the button is held. Put controls in a DialogToolbar below it,
// not in the bar. OnDismiss is what the close button does.
func DialogTitle(title string, items ...Item) *El {
	var bar *El
	closeButton := Row(WOf(core.IconBtnOf), HOf(core.CompactOf), Control("row"), HoverBg(core.Shadow), Tooltip("Close (Esc)"),
		OnClick(func() {
			bar.Dismiss()
		}),
		CrossIcon(Grow(1), Color(core.HeroMuted)),
	)
	bar = Row(append(append([]Item{
		HOf(core.DialogTitleOf), DragHandle(), PadXOf(core.DialogPadOf), GapOf(core.LooseGapOf), AlignItems(AlignCenter),
		Bg(core.Hero), Control("hero"), RoundT("menu"), BorderB(core.Border), Text(title, Bold(), Color(core.HeroText)),
	}, append(items, Spacer())...), closeButton)...)
	return bar
}

// DialogSubtitle is what a title bar says besides its name: the page shown,
// a count. It is muted (the banner's muted color) and small, gives way
// before the close button and is cut with an ellipsis; pass it to
// DialogTitle as an item.
func DialogSubtitle(text string) *El {
	return Text(text, Color(core.HeroMuted), Small(), Ellipsis(), Shrink())
}

// DialogToolbar is the row of controls under a dialog's title bar: a search
// field, Save, a switch. Give it the controls; fields that should take the
// free room get Grow(1).
func DialogToolbar(items ...Item) *El {
	return Row(append([]Item{HOf(core.DialogToolbarOf), PadXOf(core.DialogPadOf), GapOf(core.LooseGapOf), AlignItems(AlignCenter), BorderB(core.BorderSoft)}, items...)...)
}

// DialogButton is a button of a message dialog; the primary one is lit and
// is what Enter runs.
type DialogButton struct {
	Label   string
	Primary bool
	Run     func()
}

// DialogState is a modal message: a title, a text, an optional input line
// (a prompt) and buttons. Buttons close the dialog themselves in their
// Run, so the state needs no knowledge of its owner. Scroll is how far a
// MessageDialog too tall for the window has scrolled its body; assign a new
// DialogState to open the next message, which starts at the top, or set
// Scroll to its zero value when you change the fields of a kept one.
type DialogState struct {
	Title, Body string
	Input       *TextField
	Buttons     []DialogButton
	Scroll      ScrollState
}

// Default runs the primary button, if the dialog has one: what Enter does.
func (state *DialogState) Default() {
	for _, button := range state.Buttons {
		if button.Primary {
			button.Run()
			return
		}
	}
}

// MessageDialog shows state as a Dialog: a title bar with a close button,
// then the body, any content items you add (a form, between the body and
// the input), the input line and the buttons on the right. Options you give
// apply to the dialog panel: OnDismiss is what an outside click and the
// close button do, Moved lets the user drag it. Enter in the input runs the
// primary button. When the dialog reaches its height limit (see Dialog), the
// body text and the content scroll (state.Scroll) while the input and the
// buttons stay in view; a window too low even for those clips them.
func MessageDialog(state *DialogState, items ...Item) *El {
	var options, content []Item
	for _, item := range items {
		if _, ok := item.(Option); ok {
			options = append(options, item)
		} else {
			content = append(content, item)
		}
	}
	var frag *El
	dismiss := func() {
		frag.Children()[1].Dismiss()
	}
	scrolled := Scroll(&state.Scroll, append([]Item{
		GapOf(core.GutterOf),
		If(state.Body != "", Lines(state.Body, core.Muted)),
	}, content...)...)
	body := Col(
		PadOf(core.DialogPadOf), GapOf(core.GutterOf), Shrink(),
		// An empty scroll would still take a gap, and content may build to
		// nothing (a false If).
		If(len(scrolled.Children()) > 0, scrolled),
		If(state.Input != nil, Field(state.Input, HOf(core.RowOf), Bg(core.Sunken), OnEnter(state.Default), TextMenu(state.Input))),
		Row(HOf(core.RowOf), GapOf(core.LooseGapOf),
			Spacer(),
			For(state.Buttons, func(_ int, button DialogButton) Node {
				return SegmentChip(button.Label, button.Primary, MinWOf(core.DialogButtonMinWOf), OnClick(button.Run))
			}),
		),
	)
	frag = Dialog(append([]Item{WOf(core.DialogWOf), DialogTitle(state.Title, OnDismiss(dismiss)), body}, options...)...)
	return frag
}
