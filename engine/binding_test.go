package engine

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
)

func TestParseAndFormatBindings(t *testing.T) {
	tests := []struct {
		in   string
		want Binding
		out  string // the canonical text; "" for the same as in
	}{
		{"Ctrl+Enter", Binding{Key: render.KeyEnter, Ctrl: true, Bound: true}, ""},
		{"Ctrl+Shift+Tab", Binding{Key: render.KeyTab, Ctrl: true, Shift: true, Bound: true}, ""},
		{"shift+ctrl+tab", Binding{Key: render.KeyTab, Ctrl: true, Shift: true, Bound: true}, "Ctrl+Shift+Tab"},
		{"Alt+Up", Binding{Key: render.KeyArrowUp, Alt: true, Bound: true}, ""},
		{"Alt+ArrowDown", Binding{Key: render.KeyArrowDown, Alt: true, Bound: true}, "Alt+Down"},
		{"Ctrl+,", Binding{Key: render.KeyComma, Ctrl: true, Bound: true}, ""},
		{"F5", Binding{Key: render.KeyF5, Bound: true}, ""},
		{"Delete", Binding{Key: render.KeyDelete, Bound: true}, ""},
		{"Ctrl+Alt+Shift+K", Binding{Key: render.KeyK, Ctrl: true, Alt: true, Shift: true, Bound: true}, ""},
		{"ctrl+q", Binding{Key: render.KeyQ, Ctrl: true, Bound: true}, "Ctrl+Q"},
		{"", Binding{}, ""},
	}
	for _, tt := range tests {
		got, err := ParseBinding(tt.in)
		if err != nil || got != tt.want {
			t.Errorf("ParseBinding(%q) = %+v, %v; want %+v", tt.in, got, err, tt.want)
			continue
		}
		want := tt.out
		if want == "" {
			want = tt.in
		}
		if got.String() != want {
			t.Errorf("%q formats as %q, want %q", tt.in, got.String(), want)
		}
	}
}

// shortcutCases is every shortcut string the tests pin, with the key name,
// modifiers and text it must keep. Users store these strings.
var shortcutCases = []struct {
	in               string
	key              string
	ctrl, shift, alt bool
	out              string
}{
	{"Ctrl+Enter", "Enter", true, false, false, "Ctrl+Enter"},
	{"Ctrl+Shift+Tab", "Tab", true, true, false, "Ctrl+Shift+Tab"},
	{"shift+ctrl+tab", "Tab", true, true, false, "Ctrl+Shift+Tab"},
	{"Alt+Up", "ArrowUp", false, false, true, "Alt+Up"},
	{"Alt+ArrowDown", "ArrowDown", false, false, true, "Alt+Down"},
	{"Ctrl+,", "Comma", true, false, false, "Ctrl+,"},
	{"F5", "F5", false, false, false, "F5"},
	{"Delete", "Delete", false, false, false, "Delete"},
	{"Ctrl+Alt+Shift+K", "K", true, true, true, "Ctrl+Alt+Shift+K"},
	{"ctrl+q", "Q", true, false, false, "Ctrl+Q"},
	{"Escape", "Escape", false, false, false, "Escape"},
	{"Ctrl+Space", "Space", true, false, false, "Ctrl+Space"},
}

func TestEveryShortcutStringKeepsItsBindingAndText(t *testing.T) {
	for _, tt := range shortcutCases {
		got, err := ParseBinding(tt.in)
		if err != nil {
			t.Errorf("ParseBinding(%q): %v", tt.in, err)
			continue
		}
		if got.Key.String() != tt.key || got.Ctrl != tt.ctrl || got.Shift != tt.shift || got.Alt != tt.alt || !got.Bound {
			t.Errorf("ParseBinding(%q) = key %q %+v, want key %q ctrl %v shift %v alt %v", tt.in, got.Key.String(), got, tt.key, tt.ctrl, tt.shift, tt.alt)
		}
		if got.String() != tt.out {
			t.Errorf("%q prints as %q, want %q", tt.in, got.String(), tt.out)
		}
	}
}

func TestEveryKeyNameParsesAndPrintsAsItsLabel(t *testing.T) {
	n := 0
	for name, key := range keyByName {
		if IsModifierKey(key) {
			t.Errorf("modifier %q is bindable", name)
		}
		label := keyLabel(key)
		got, err := ParseBinding(label)
		if err != nil || got.Key != key || got.String() != label {
			t.Errorf("ParseBinding(%q) = %+v, %v; want key %v printing %q", label, got, err, key, label)
		}
		n++
	}
	if n < 80 {
		t.Errorf("only %d key names are bindable", n)
	}
}

func TestParseBindingRejectsNonsense(t *testing.T) {
	for _, in := range []string{"Ctrl+", "Ctrl", "Shift+Alt", "Banana", "Ctrl+Ctrl+T", "Ctrl+T+Y", "+", "Ctrl++"} {
		if b, err := ParseBinding(in); err == nil {
			t.Errorf("ParseBinding(%q) = %+v, want an error", in, b)
		}
	}
}

func TestBindingMatchesOnlyTheExactModifiers(t *testing.T) {
	ctrlTab := Binding{Key: render.KeyTab, Ctrl: true, Bound: true}
	tests := []struct {
		name    string
		binding Binding
		input   core.Input
		key     render.Key
		want    bool
	}{
		{"same key and modifiers", ctrlTab, core.Input{Ctrl: true}, render.KeyTab, true},
		{"extra shift", ctrlTab, core.Input{Ctrl: true, Shift: true}, render.KeyTab, false},
		{"missing ctrl", ctrlTab, core.Input{}, render.KeyTab, false},
		{"other key", ctrlTab, core.Input{Ctrl: true}, render.KeyEnter, false},
		{"unbound never matches", Binding{Key: render.KeyTab}, core.Input{}, render.KeyTab, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.binding.Matches(tt.input, tt.key); got != tt.want {
				t.Fatalf("Matches = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindingFromReadsTheCombinationPressedThisFrame(t *testing.T) {
	tests := []struct {
		name   string
		input  core.Input
		want   Binding
		wantOK bool
	}{
		{"key with modifiers", core.Input{Keys: []render.Key{render.KeyTab}, Ctrl: true, Shift: true}, Binding{Key: render.KeyTab, Ctrl: true, Shift: true, Bound: true}, true},
		{"modifier keys alone are waited out", core.Input{Keys: []render.Key{render.KeyControlLeft}, Ctrl: true}, Binding{}, false},
		{"modifier then key takes the key", core.Input{Keys: []render.Key{render.KeyShiftLeft, render.KeyA}, Shift: true}, Binding{Key: render.KeyA, Shift: true, Bound: true}, true},
		{"escape alone", core.Input{Escape: true}, Binding{Key: render.KeyEscape, Bound: true}, true},
		{"nothing pressed", core.Input{}, Binding{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := BindingFrom(tt.input)
			if got != tt.want || ok != tt.wantOK {
				t.Fatalf("BindingFrom = %+v, %v; want %+v, %v", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestBindingTypesText(t *testing.T) {
	tests := []struct {
		name    string
		binding Binding
		want    bool
	}{
		{"plain letter", Binding{Key: render.KeyA, Bound: true}, true},
		{"shift letter", Binding{Key: render.KeyA, Shift: true, Bound: true}, true},
		{"ctrl letter", Binding{Key: render.KeyA, Ctrl: true, Bound: true}, false},
		{"alt arrow", Binding{Key: render.KeyArrowUp, Alt: true, Bound: true}, false},
		{"function key", Binding{Key: render.KeyF5, Bound: true}, false},
		{"delete", Binding{Key: render.KeyDelete, Bound: true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.binding.TypesText(); got != tt.want {
				t.Fatalf("TypesText = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFunctionKey(t *testing.T) {
	for key, want := range map[render.Key]bool{render.KeyF1: true, render.KeyF12: true, render.KeyA: false, render.KeyEnter: false} {
		if got := IsFunctionKey(key); got != want {
			t.Errorf("IsFunctionKey(%v) = %v, want %v", key, got, want)
		}
	}
}
