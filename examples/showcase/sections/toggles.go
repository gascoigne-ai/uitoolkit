package sections

import (
	"fmt"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/markup"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// Fixed sizes and slider ranges of the toggles section.
const (
	sliderW     = 320
	valueW      = 60
	volumeMax   = 100
	zoomMin     = 50
	zoomMax     = 200
	zoomStep    = 5
	zoomDefault = 100
)

// flip is the handler that turns a setting over.
func flip(value *bool) func() { return func() { *value = !*value } }

// switchRow is a switch and its label. Markup has no switch element, so both
// forms take this node: the template puts it in with slot.
func switchRow(label string, toggle *El) Node {
	return Row(HOf(core.RowOf), GapOf(core.LooseGapOf), AlignItems(AlignCenter), toggle, Text(label))
}

func (s *Showcase) switches() (autosave, offline, disabled Node) {
	state := &s.widgets
	return switchRow("Save automatically", Switch(state.autosave, OnChange(flip(&state.autosave)))),
		switchRow("Work offline", Switch(state.offline, OnChange(flip(&state.offline)))),
		switchRow("Disabled switch", DisabledSwitch(true))
}

func (s *Showcase) volumeText() string {
	return fmt.Sprintf("%d%%", s.widgets.volumeDrag.Shown(s.widgets.volume))
}

func (s *Showcase) zoomText() string {
	return fmt.Sprintf("%d%%", s.widgets.zoomDrag.Shown(s.widgets.zoom))
}

func togglesGo(s *Showcase, _ *core.Ctx) Node {
	state := &s.widgets
	autosave, offline, disabled := s.switches()
	return Scroll(&state.togglesPage, PadOf(core.GutterOf), GapOf(core.GutterOf),
		Caption("Checkboxes"),
		CheckRow(state.wrapLines, OnClick(flip(&state.wrapLines)), AlignItems(AlignCenter), Text("Wrap long lines")),
		CheckRow(state.showHidden, OnClick(flip(&state.showHidden)), AlignItems(AlignCenter), Text("Show hidden files")),
		Row(HOf(core.RowOf), GapOf(core.LooseGapOf), AlignItems(AlignCenter),
			Checkbox(state.termsAgreed, OnChange(flip(&state.termsAgreed))),
			Text("A checkbox on its own"),
		),
		Caption("Switches"),
		autosave, offline, disabled,
		Caption("Sliders"),
		FormRow("Volume", formLabelW,
			Slider(&state.volumeDrag, 0, volumeMax, 1, state.volume, W(sliderW), Grow(0), OnSlide(func(value int) { state.volume = value })),
			Text(s.volumeText(), W(valueW)),
		),
		FormRow("Zoom", formLabelW,
			Slider(&state.zoomDrag, zoomMin, zoomMax, zoomStep, state.zoom, W(sliderW), Grow(0), SliderMid(zoomDefault),
				OnSlideEnd(func(value int) { state.zoom = value })),
			Text(s.zoomText(), W(valueW)),
		),
	)
}

func togglesMarkup(s *Showcase, ctx *core.Ctx) Node {
	state := &s.widgets
	autosave, offline, disabled := s.switches()
	return fill(ctx, togglesTemplate, markup.Scope{
		"page":      &state.togglesPage,
		"wrapLines": state.wrapLines, "flipWrap": flip(&state.wrapLines),
		"showHidden": state.showHidden, "flipHidden": flip(&state.showHidden),
		"termsAgreed": state.termsAgreed, "flipTerms": flip(&state.termsAgreed),
		"autosave": autosave, "offline": offline, "disabled": disabled,
		"volumeDrag": &state.volumeDrag, "volume": state.volume, "volumeMax": volumeMax,
		"setVolume": func(value int) { state.volume = value }, "volumeText": s.volumeText(),
		"zoomDrag": &state.zoomDrag, "zoom": state.zoom, "zoomMin": zoomMin, "zoomMax": zoomMax, "zoomStep": zoomStep,
		"zoomDefault": zoomDefault, "setZoom": func(value int) { state.zoom = value }, "zoomText": s.zoomText(),
		"labelW": formLabelW, "sliderW": sliderW, "valueW": valueW,
	})
}

var togglesTemplate = markup.Must("toggles.vml", togglesSource)

const togglesSource = `
<scroll state="page" pad="{{gutter}}" gap="{{gutter}}">
  <caption>Checkboxes</caption>
  <check-row on="{{wrapLines}}" on-click="flipWrap" align-items="center"><text>Wrap long lines</text></check-row>
  <check-row on="{{showHidden}}" on-click="flipHidden" align-items="center"><text>Show hidden files</text></check-row>
  <row h="{{rowH}}" gap="{{looseGap}}" align-items="center">
    <checkbox on="{{termsAgreed}}" on-change="flipTerms"/>
    <text>A checkbox on its own</text>
  </row>
  <caption>Switches</caption>
  <slot name="autosave"/>
  <slot name="offline"/>
  <slot name="disabled"/>
  <caption>Sliders</caption>
  <form-row label="Volume" label-width="{{labelW}}">
    <slider state="volumeDrag" min="0" max="{{volumeMax}}" step="1" value="{{volume}}" w="{{sliderW}}" grow="0" on-slide="setVolume"/>
    <text w="{{valueW}}">{{volumeText}}</text>
  </form-row>
  <form-row label="Zoom" label-width="{{labelW}}">
    <slider state="zoomDrag" min="{{zoomMin}}" max="{{zoomMax}}" step="{{zoomStep}}" value="{{zoom}}" w="{{sliderW}}" grow="0" slider-mid="{{zoomDefault}}" on-slide-end="setZoom"/>
    <text w="{{valueW}}">{{zoomText}}</text>
  </form-row>
</scroll>`
