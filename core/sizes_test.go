package core

import "testing"

// Every size keeps the value it had before themes could set it: the golden
// scenes and the layout dumps depend on each one.
func TestDefaultSizesKeepTodaysValues(t *testing.T) {
	want := Sizes{
		Line: 17, SmallLine: 15,
		Row: 29, Compact: 25, Slim: 21, Label: 23, Note: 19,
		Toolbar: 37, TabBar: 41, DialogToolbar: 41, DialogTitle: 45,
		IconBtnW: 33, WordBtnW: 45, Checkbox: 15,
		Leading: 0,
		Steps:   Steps{Hair: 2, Near: 4, Tight: 6, Loose: 8, Label: 10, Gutter: 12, Dialog: 16},
		TabLead: 2, EdgeW: 2, CaretW: 2, NearGap: 4, ShadowOffset: 4,
		TightGap: 6, TabGap: 6, RuleInset: 6, EdgeGrab: 6,
		LooseGap: 8, CellPad: 8, LabelPad: 10, TabSepInset: 10, Gutter: 12, DialogPad: 16,
		DialogW: 380, DialogButtonMinW: 80, DialogMargin: 40, DialogMaxH: 560,
		ScrollbarW: 8, MinThumbLen: 24, ThumbInset: 1, HScrollArrowW: 18,
		MenuMinW: 120, MenuMaxW: 420, MenuSepH: 9, MenuPad: 4, MenuHintH: 16,
		MenuRowPadX: 10, MenuIconW: 14, MenuShortcutGap: 16,
		TipPadX: 8, TipPadY: 5, TipOffset: 16,
		ToastW: 320, ToastH: 36, ToastGap: 6, ToastMargin: 8,
		ButtonPadX: 10, FieldPadX: 10, SelectionInset: 6, RevealMargin: 12,
		SliderW: 120, SliderH: 20, SliderTrack: 4, SliderThumbW: 10, SliderThumbH: 16,
		SliderTick: 10, SliderPad: 7, DividerGrab: 2, EdgeReach: 4,
		TableActionW: 36, TableGripW: 20, MinColumnW: 56,
		FoldColW: 18, MonoTextLead: 4, MonoLinePad: 4, SwitchInset: 2,
	}
	if got := DefaultSizes(); got != want {
		t.Fatalf("DefaultSizes() =\n%+v, want\n%+v", got, want)
	}
}
