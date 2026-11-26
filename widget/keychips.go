package widget

import (
	"strings"

	"github.com/gascoigne-ai/uitoolkit/core"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// KeyChips shows key text such as "Ctrl+Shift+Tab" as one small chip per
// key, so a combination reads as separate keys. A plus that is itself the
// key ("Ctrl++") gets its own chip.
func KeyChips(text string, opts ...Option) *El {
	keys := strings.Split(text, "+")
	if strings.HasSuffix(text, "++") {
		keys = append(strings.Split(strings.TrimSuffix(text, "++"), "+"), "+")
	}
	return Row(With([]Item{
		GapOf(core.NearGapOf), AlignItems(AlignCenter),
		For(keys, func(_ int, key string) Node {
			return Row(Bg(core.Chip), Control("button"), PadXOf(core.NearGapOf), HOf(core.CompactOf),
				Text(key, Small(), Color(core.ChipText)))
		}),
	}, opts)...)
}
