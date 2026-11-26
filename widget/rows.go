package widget

import (
	"github.com/gascoigne-ai/uitoolkit/core"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// ListRow is a row of a list: the selected one has the selection
// background, the others a tint under the pointer. Its content and options
// (H, PadX, OnClick) are the caller's.
func ListRow(selected bool, items ...Item) *El {
	look := HoverBg(core.Hover)
	if selected {
		look = Bg(core.Selected)
	}
	return Row(append([]Item{Control("row"), look}, items...)...)
}

// Note is a one-line message under a control, in small text of the given
// color (core.Error for a problem, core.Muted for a hint). It is noteH high
// and cuts a long text with an ellipsis.
func Note(text string, tone core.Token, opts ...Option) *El {
	return Row(With([]Item{
		HOf(core.NoteOf),
		Text(text, Small(), Color(tone), Ellipsis(), Grow(1)),
	}, opts)...)
}

// FormRow is a settings-style row: a label in a column of labelWidth, then
// the controls. Rows with the same labelWidth line their controls up.
func FormRow(label string, labelWidth int, items ...Item) *El {
	return Row(append([]Item{HOf(core.RowOf), GapOf(core.GutterOf), Text(label, W(labelWidth))}, items...)...)
}

// CheckRow is a checkbox followed by its items (usually a Text label and a
// muted hint). Give it OnClick: the whole row is the click target.
func CheckRow(on bool, items ...Item) *El {
	return Row(append([]Item{HOf(core.RowOf), GapOf(core.LooseGapOf), Checkbox(on)}, items...)...)
}

// Toolbar is a bar of controls over a pane: toolbarH high, inset by the
// gutter, centered vertically, with a line under it.
func Toolbar(items ...Item) *El {
	return Row(append([]Item{
		HOf(core.ToolbarOf), PadXOf(core.GutterOf), GapOf(core.TightGapOf), AlignItems(AlignCenter), BorderB(core.Border),
	}, items...)...)
}

// TabBar is the strip that holds Tabs or ScrollTabs (and anything beside
// them): tabBarH high, TabGap of room above the tabs, a line under the
// inactive ones.
func TabBar(items ...Item) *El {
	return Row(append([]Item{HOf(core.TabBarOf), PadTOf(core.TabGapOf), BorderB(core.Border)}, items...)...)
}
