package sections

import (
	"slices"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// Sizes and texts of the theming section.
const (
	swatchW        = 96
	swatchH        = 28
	swatchesPerRow = 8
	sampleFieldW   = 160
	editorH        = 150
	editorHint     = `A theme file laid over the chosen theme, for example {"colors": {"accent": "#7ab8ff"}}`
)

var swatchTokens = []core.Token{
	core.Frame, core.Bg, core.Panel, core.Sunken, core.Hover, core.Selected, core.Chip, core.Track,
	core.Border, core.BorderStrong, core.BorderSoft, core.Rule, core.Text, core.Muted, core.Faint, core.Accent,
	core.AccentLine, core.AccentSoft, core.OK, core.Warn, core.Error, core.Info, core.Selection, core.FindMatch,
}

func swatchRows() [][]core.Token {
	return slices.Collect(slices.Chunk(swatchTokens, swatchesPerRow))
}

func (s *Showcase) sampleSwitch() Node {
	return Switch(s.widgets.sampleOn, OnChange(flip(&s.widgets.sampleOn)))
}

func themingGo(s *Showcase, _ *core.Ctx) Node {
	state := &s.widgets
	return Scroll(&state.themingPage, PadOf(core.GutterOf), GapOf(core.GutterOf),
		Caption("Colors"),
		For(swatchRows(), func(_ int, tokens []core.Token) Node {
			return Row(GapOf(core.LooseGapOf),
				For(tokens, func(_ int, token core.Token) Node {
					return Col(W(swatchW), GapOf(core.NearGapOf),
						Row(H(swatchH), Bg(token), Control("swatch")),
						Text(string(token), Small(), Muted()),
					)
				}),
			)
		}),
		Caption("Widgets"),
		Row(HOf(core.RowOf), GapOf(core.LooseGapOf), AlignItems(AlignCenter),
			Button("Button", HOf(core.RowOf)),
			PrimaryButton("Primary", HOf(core.RowOf)),
			LinkButton("Link", HOf(core.RowOf)),
			Checkbox(state.sampleChecked, OnChange(flip(&state.sampleChecked))),
			s.sampleSwitch(),
			Field(&state.sample, W(sampleFieldW), HOf(core.RowOf), Bg(core.Sunken)),
			Text("Muted text", Muted()),
		),
		Caption("Live theme (JSON)"),
		Text(editorHint, Small(), Muted(), Ellipsis()),
		Area(&s.ThemeJSON, H(editorH), Bg(core.Sunken), Control("field"), OnChange(s.ApplyThemeJSON)),
		If(s.ThemeErr != "", Note(s.ThemeErr, core.Error)),
	)
}

func themingMarkup(s *Showcase, ctx *core.Ctx) Node {
	state := &s.widgets
	return fill(ctx, themingTemplate, markup.Scope{
		"page": &state.themingPage, "rows": swatchRows(), "swatchW": swatchW, "swatchH": swatchH,
		"checked": state.sampleChecked, "flipChecked": flip(&state.sampleChecked), "switch": s.sampleSwitch(),
		"sample": &state.sample, "fieldW": sampleFieldW, "hint": editorHint,
		"editor": &s.ThemeJSON, "editorH": editorH, "apply": s.ApplyThemeJSON,
		"hasError": s.ThemeErr != "", "error": s.ThemeErr,
	})
}

var themingTemplate = markup.Must("theming.vml", themingSource)

const themingSource = `
<scroll state="page" pad="{{gutter}}" gap="{{gutter}}">
  <caption>Colors</caption>
  <row for="tokens in rows" gap="{{looseGap}}">
    <col for="token in tokens" w="{{swatchW}}" gap="{{nearGap}}">
      <row h="{{swatchH}}" bg="{{token}}" control="swatch"/>
      <text small muted>{{token}}</text>
    </col>
  </row>
  <caption>Widgets</caption>
  <row h="{{rowH}}" gap="{{looseGap}}" align-items="center">
    <button h="{{rowH}}">Button</button>
    <primary-button h="{{rowH}}">Primary</primary-button>
    <link-button h="{{rowH}}">Link</link-button>
    <checkbox on="{{checked}}" on-change="flipChecked"/>
    <slot name="switch"/>
    <field state="sample" w="{{fieldW}}" h="{{rowH}}" bg="sunken"/>
    <text muted>Muted text</text>
  </row>
  <caption>Live theme (JSON)</caption>
  <text small muted ellipsis>{{hint}}</text>
  <area state="editor" h="{{editorH}}" bg="sunken" control="field" on-change="apply"/>
  <note tone="error" if="hasError">{{error}}</note>
</scroll>`
