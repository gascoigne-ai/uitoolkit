package render

import "testing"

func TestKeyNameTableShape(t *testing.T) {
	if got, want := len(keyNames), int(KeyMax)+1; got != want {
		t.Fatalf("len(keyNames) = %d, want %d", got, want)
	}
	// 26 letters, 92 other keys and 4 aggregates.
	if KeyMax != 122 {
		t.Errorf("KeyMax = %d, want 122", KeyMax)
	}
	seen := map[string]Key{}
	for key := KeyA; key <= KeyMax; key++ {
		name := key.String()
		if name == "" {
			t.Errorf("key %d has no name", key)
			continue
		}
		if other, found := seen[name]; found {
			t.Errorf("keys %d and %d are both named %q", other, key, name)
		}
		seen[name] = key
	}
	for _, key := range []Key{KeyUnknown, KeyMax + 1, 255} {
		if got := key.String(); got != "" {
			t.Errorf("Key(%d).String() = %q, want empty", key, got)
		}
	}
}

// The names are the ones users already have in their shortcut settings.
func TestKeyNamesKeepAggregateAndSideModifiers(t *testing.T) {
	cases := []struct {
		key  Key
		want string
	}{
		{KeyAlt, "Alt"}, {KeyAltLeft, "AltLeft"}, {KeyAltRight, "AltRight"},
		{KeyControl, "Control"}, {KeyControlLeft, "ControlLeft"}, {KeyControlRight, "ControlRight"},
		{KeyShift, "Shift"}, {KeyShiftLeft, "ShiftLeft"}, {KeyShiftRight, "ShiftRight"},
		{KeyMeta, "Meta"}, {KeyMetaLeft, "MetaLeft"}, {KeyMetaRight, "MetaRight"},
		{KeyA, "A"}, {KeyArrowUp, "ArrowUp"}, {KeyComma, "Comma"}, {KeyDigit0, "Digit0"},
		{KeyF12, "F12"}, {KeyNumpadEnter, "NumpadEnter"}, {KeyBackquote, "Backquote"},
	}
	for _, testCase := range cases {
		t.Run(testCase.want, func(t *testing.T) {
			if got := testCase.key.String(); got != testCase.want {
				t.Errorf("Key(%d).String() = %q, want %q", testCase.key, got, testCase.want)
			}
		})
	}
}
