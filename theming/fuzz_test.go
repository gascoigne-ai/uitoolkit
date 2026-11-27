package theming

import (
	"reflect"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
)

func FuzzParseTheme(fuzz *testing.F) {
	fuzz.Add([]byte(`{"name":"x","extends":"dark","colors":{"bg":"#101010"}}`))
	fuzz.Add([]byte(`{"colors":{}}x`))
	fuzz.Add([]byte(`null`))
	fuzz.Add([]byte(`{"iconStroke":1e30}`))
	fuzz.Add([]byte(`{"metrics":{"unit":4,"leading":8,"steps":{"hair":16,"near":16,"tight":16,"loose":16,"label":16,"gutter":16,"dialog":16}}}`))
	fuzz.Add([]byte(`{"extends":"dark","metrics":{"unit":1,"sizes":{"gutter":0,"rowH":160,"dialogW":1600,"dialogMaxH":4000}}}`))
	fuzz.Add([]byte(`{"metrics":{"unit":0,"sizes":{"gutter":null,"nope":1e30}}}`))
	fuzz.Add([]byte(`{"extends":"dark","feel":{"caretHalfPeriod":200,"keyRepeatDelay":150,"keyRepeatInterval":10,"holdDelay":1000,"toastTime":60000,"wheelColumns":40}}`))
	fuzz.Add([]byte(`{"feel":{"tooltipDelay":0,"holdInterval":null,"wheelStep":1e30,"multiClickTime":-1,"nope":9223372036854775807}}`))
	fuzz.Fuzz(func(t *testing.T, data []byte) {
		base := Dark()
		base.RadiusScale = 0
		th, err := ParseTheme(data, &base)
		if err != nil {
			return
		}
		if th.IconStroke < 0 || th.IconStroke > 4 {
			t.Errorf("iconStroke %v is outside 0 and 4", th.IconStroke)
		}
		if th.Controls == nil {
			t.Error("no controls map")
		}
		if th.RadiusScale != base.RadiusScale {
			t.Errorf("radius scale %v, want the base's %v", th.RadiusScale, base.RadiusScale)
		}
		checkSizesBounded(t, th.Metrics)
		// An accepted theme states only values within their bounds, which
		// Resolved therefore leaves as they are.
		resolved := th.Feel.Resolved()
		checkFeelStated(t, reflect.ValueOf(th.Feel), reflect.ValueOf(resolved))
	})
}

// checkFeelStated fails for a value of stated that is neither unset nor what
// it resolves to.
func checkFeelStated(t *testing.T, stated, resolved reflect.Value) {
	t.Helper()
	if stated.Kind() == reflect.Struct {
		for index := range stated.NumField() {
			checkFeelStated(t, stated.Field(index), resolved.Field(index))
		}
		return
	}
	if value := stated.Int(); resolved.Int() <= 0 || value != 0 && value != resolved.Int() {
		t.Errorf("feel value %d resolves to %d", value, resolved.Int())
	}
}

// checkSizesBounded resolves metrics at both ends of the line heights: no
// theme file may give a size that is negative or large enough to overflow
// the layout's sums.
func checkSizesBounded(t *testing.T, metrics core.Metrics) {
	t.Helper()
	const limit = 4096
	for _, line := range []int{core.MinLine, core.MaxLine} {
		sizes := core.SizesFor(line, line, metrics)
		steps := sizes.Steps
		values := []any{sizes.Leading, steps.Hair, steps.Near, steps.Tight, steps.Loose, steps.Label, steps.Gutter, steps.Dialog}
		for _, value := range sizes.Scope() {
			values = append(values, value)
		}
		for _, value := range values {
			if pixels := value.(int); pixels < 0 || pixels > limit {
				t.Errorf("line %d: size %d is outside 0 and %d in %+v", line, pixels, limit, sizes)
			}
		}
	}
}
