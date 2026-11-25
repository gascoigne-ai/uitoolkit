package engine

import (
	"fmt"
	"strings"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
)

// keyNames spells keys whose render name is not what users read. It has no
// reader outside this package, so it is not exported.
var keyNames = map[render.Key]string{
	render.KeyComma:     ",",
	render.KeyArrowUp:   "Up",
	render.KeyArrowDown: "Down",
}

// Binding is a key with modifiers; an unbound action has none.
type Binding struct {
	Key              render.Key
	Ctrl, Shift, Alt bool
	Bound            bool
}

// IsModifierKey reports whether k is a modifier key itself.
func IsModifierKey(key render.Key) bool {
	switch key {
	case render.KeyAlt, render.KeyAltLeft, render.KeyAltRight,
		render.KeyControl, render.KeyControlLeft, render.KeyControlRight,
		render.KeyShift, render.KeyShiftLeft, render.KeyShiftRight,
		render.KeyMeta, render.KeyMetaLeft, render.KeyMetaRight:
		return true
	}
	return false
}

// keyLabel is how a key is written in key strings: the friendly name
// from keyNames, else the render name.
func keyLabel(key render.Key) string {
	if name, ok := keyNames[key]; ok {
		return name
	}
	return key.String()
}

// keyByName finds a bindable key by its lower-case name, friendly or
// the render name.
var keyByName = func() map[string]render.Key {
	m := map[string]render.Key{}
	for k := render.KeyUnknown + 1; k <= render.KeyMax; k++ {
		if IsModifierKey(k) {
			continue
		}
		m[strings.ToLower(k.String())] = k
		m[strings.ToLower(keyLabel(k))] = k
	}
	return m
}()

// ParseBinding reads "Ctrl+Shift+Tab" (modifiers in any order, names
// case-insensitive). "" is a valid, unbound Binding.
func ParseBinding(text string) (Binding, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Binding{}, nil
	}
	parts := strings.Split(text, "+")
	var b Binding
	for _, mod := range parts[:len(parts)-1] {
		flag := map[string]*bool{"ctrl": &b.Ctrl, "control": &b.Ctrl, "alt": &b.Alt, "shift": &b.Shift}[strings.ToLower(strings.TrimSpace(mod))]
		if flag == nil || *flag {
			return Binding{}, fmt.Errorf("%q is not a modifier, or is repeated", mod)
		}
		*flag = true
	}
	key, ok := keyByName[strings.ToLower(strings.TrimSpace(parts[len(parts)-1]))]
	if !ok {
		return Binding{}, fmt.Errorf("%q is not a key", parts[len(parts)-1])
	}
	b.Key, b.Bound = key, true
	return b, nil
}

// String writes the Binding as users read it: modifiers in the order
// Ctrl, Alt, Shift, then the key; "" when unbound.
func (b Binding) String() string {
	if !b.Bound {
		return ""
	}
	var parts []string
	if b.Ctrl {
		parts = append(parts, "Ctrl")
	}
	if b.Alt {
		parts = append(parts, "Alt")
	}
	if b.Shift {
		parts = append(parts, "Shift")
	}
	return strings.Join(append(parts, keyLabel(b.Key)), "+")
}

// Matches reports whether key, pressed this frame under the modifiers in
// input, triggers the Binding. Modifiers must match exactly, so Ctrl+Tab
// does not fire on Ctrl+Shift+Tab.
func (b Binding) Matches(input core.Input, key render.Key) bool {
	return b.Bound && b.Key == key && b.Ctrl == input.Ctrl && b.Shift == input.Shift && b.Alt == input.Alt
}

// BindingFrom reads the combination a user pressed this frame, for a
// "press the new shortcut" prompt. Escape counts unless Ctrl or Alt is
// held; modifier keys alone are waited out (ok is false).
func BindingFrom(input core.Input) (binding Binding, ok bool) {
	if input.Escape && !input.Ctrl && !input.Alt {
		return Binding{Key: render.KeyEscape, Bound: true}, true
	}
	for _, pressed := range input.Keys {
		if pressed == render.KeyUnknown || IsModifierKey(pressed) {
			continue
		}
		return Binding{Key: pressed, Ctrl: input.Ctrl, Shift: input.Shift, Alt: input.Alt, Bound: true}, true
	}
	return Binding{}, false
}

// IsFunctionKey reports whether k is one of F1 to F12.
func IsFunctionKey(key render.Key) bool { return key >= render.KeyF1 && key <= render.KeyF12 }

// TypesText reports whether the Binding is a key the user also types
// with: no Ctrl or Alt, not a function key, not Delete. Such a binding
// should fire only while no text field has focus.
func (b Binding) TypesText() bool {
	return !b.Ctrl && !b.Alt && !IsFunctionKey(b.Key) && b.Key != render.KeyDelete
}
