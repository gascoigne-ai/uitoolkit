package render

// Key is a physical key of a US keyboard, or one of the aggregates KeyAlt,
// KeyControl, KeyShift and KeyMeta, which stand for either side of a modifier.
// The numbers may change between versions; the names String returns do not,
// because users store them in shortcut settings.
type Key uint8

// The keys. KeyUnknown is the zero value and has no name.
const (
	KeyUnknown Key = iota
	KeyA
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	KeyAltLeft
	KeyAltRight
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	KeyArrowUp
	KeyBackquote
	KeyBackslash
	KeyBackspace
	KeyBracketLeft
	KeyBracketRight
	KeyCapsLock
	KeyComma
	KeyContextMenu
	KeyControlLeft
	KeyControlRight
	KeyDelete
	KeyDigit0
	KeyDigit1
	KeyDigit2
	KeyDigit3
	KeyDigit4
	KeyDigit5
	KeyDigit6
	KeyDigit7
	KeyDigit8
	KeyDigit9
	KeyEnd
	KeyEnter
	KeyEqual
	KeyEscape
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
	KeyF21
	KeyF22
	KeyF23
	KeyF24
	KeyHome
	KeyInsert
	KeyIntlBackslash
	KeyMetaLeft
	KeyMetaRight
	KeyMinus
	KeyNumLock
	KeyNumpad0
	KeyNumpad1
	KeyNumpad2
	KeyNumpad3
	KeyNumpad4
	KeyNumpad5
	KeyNumpad6
	KeyNumpad7
	KeyNumpad8
	KeyNumpad9
	KeyNumpadAdd
	KeyNumpadDecimal
	KeyNumpadDivide
	KeyNumpadEnter
	KeyNumpadEqual
	KeyNumpadMultiply
	KeyNumpadSubtract
	KeyPageDown
	KeyPageUp
	KeyPause
	KeyPeriod
	KeyPrintScreen
	KeyQuote
	KeyScrollLock
	KeySemicolon
	KeyShiftLeft
	KeyShiftRight
	KeySlash
	KeySpace
	KeyTab
	KeyAlt
	KeyControl
	KeyShift
	KeyMeta

	// KeyMax is the highest key.
	KeyMax = KeyMeta
)

// keyNames is indexed by key. KeyUnknown has no entry, so its name is "".
var keyNames = [KeyMax + 1]string{
	KeyA:              "A",
	KeyB:              "B",
	KeyC:              "C",
	KeyD:              "D",
	KeyE:              "E",
	KeyF:              "F",
	KeyG:              "G",
	KeyH:              "H",
	KeyI:              "I",
	KeyJ:              "J",
	KeyK:              "K",
	KeyL:              "L",
	KeyM:              "M",
	KeyN:              "N",
	KeyO:              "O",
	KeyP:              "P",
	KeyQ:              "Q",
	KeyR:              "R",
	KeyS:              "S",
	KeyT:              "T",
	KeyU:              "U",
	KeyV:              "V",
	KeyW:              "W",
	KeyX:              "X",
	KeyY:              "Y",
	KeyZ:              "Z",
	KeyAltLeft:        "AltLeft",
	KeyAltRight:       "AltRight",
	KeyArrowDown:      "ArrowDown",
	KeyArrowLeft:      "ArrowLeft",
	KeyArrowRight:     "ArrowRight",
	KeyArrowUp:        "ArrowUp",
	KeyBackquote:      "Backquote",
	KeyBackslash:      "Backslash",
	KeyBackspace:      "Backspace",
	KeyBracketLeft:    "BracketLeft",
	KeyBracketRight:   "BracketRight",
	KeyCapsLock:       "CapsLock",
	KeyComma:          "Comma",
	KeyContextMenu:    "ContextMenu",
	KeyControlLeft:    "ControlLeft",
	KeyControlRight:   "ControlRight",
	KeyDelete:         "Delete",
	KeyDigit0:         "Digit0",
	KeyDigit1:         "Digit1",
	KeyDigit2:         "Digit2",
	KeyDigit3:         "Digit3",
	KeyDigit4:         "Digit4",
	KeyDigit5:         "Digit5",
	KeyDigit6:         "Digit6",
	KeyDigit7:         "Digit7",
	KeyDigit8:         "Digit8",
	KeyDigit9:         "Digit9",
	KeyEnd:            "End",
	KeyEnter:          "Enter",
	KeyEqual:          "Equal",
	KeyEscape:         "Escape",
	KeyF1:             "F1",
	KeyF2:             "F2",
	KeyF3:             "F3",
	KeyF4:             "F4",
	KeyF5:             "F5",
	KeyF6:             "F6",
	KeyF7:             "F7",
	KeyF8:             "F8",
	KeyF9:             "F9",
	KeyF10:            "F10",
	KeyF11:            "F11",
	KeyF12:            "F12",
	KeyF13:            "F13",
	KeyF14:            "F14",
	KeyF15:            "F15",
	KeyF16:            "F16",
	KeyF17:            "F17",
	KeyF18:            "F18",
	KeyF19:            "F19",
	KeyF20:            "F20",
	KeyF21:            "F21",
	KeyF22:            "F22",
	KeyF23:            "F23",
	KeyF24:            "F24",
	KeyHome:           "Home",
	KeyInsert:         "Insert",
	KeyIntlBackslash:  "IntlBackslash",
	KeyMetaLeft:       "MetaLeft",
	KeyMetaRight:      "MetaRight",
	KeyMinus:          "Minus",
	KeyNumLock:        "NumLock",
	KeyNumpad0:        "Numpad0",
	KeyNumpad1:        "Numpad1",
	KeyNumpad2:        "Numpad2",
	KeyNumpad3:        "Numpad3",
	KeyNumpad4:        "Numpad4",
	KeyNumpad5:        "Numpad5",
	KeyNumpad6:        "Numpad6",
	KeyNumpad7:        "Numpad7",
	KeyNumpad8:        "Numpad8",
	KeyNumpad9:        "Numpad9",
	KeyNumpadAdd:      "NumpadAdd",
	KeyNumpadDecimal:  "NumpadDecimal",
	KeyNumpadDivide:   "NumpadDivide",
	KeyNumpadEnter:    "NumpadEnter",
	KeyNumpadEqual:    "NumpadEqual",
	KeyNumpadMultiply: "NumpadMultiply",
	KeyNumpadSubtract: "NumpadSubtract",
	KeyPageDown:       "PageDown",
	KeyPageUp:         "PageUp",
	KeyPause:          "Pause",
	KeyPeriod:         "Period",
	KeyPrintScreen:    "PrintScreen",
	KeyQuote:          "Quote",
	KeyScrollLock:     "ScrollLock",
	KeySemicolon:      "Semicolon",
	KeyShiftLeft:      "ShiftLeft",
	KeyShiftRight:     "ShiftRight",
	KeySlash:          "Slash",
	KeySpace:          "Space",
	KeyTab:            "Tab",
	KeyAlt:            "Alt",
	KeyControl:        "Control",
	KeyShift:          "Shift",
	KeyMeta:           "Meta",
}

// String returns the name of the key: its constant without the Key prefix. It
// returns "" for KeyUnknown and for a value above KeyMax.
func (key Key) String() string {
	if key > KeyMax {
		return ""
	}
	return keyNames[key]
}
