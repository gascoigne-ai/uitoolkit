package theming

import "testing"

func FuzzParseTheme(fuzz *testing.F) {
	fuzz.Add([]byte(`{"name":"x","extends":"dark","colors":{"bg":"#101010"}}`))
	fuzz.Add([]byte(`{"colors":{}}x`))
	fuzz.Add([]byte(`null`))
	fuzz.Add([]byte(`{"iconStroke":1e30}`))
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
	})
}
