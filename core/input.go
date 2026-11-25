package core

import (
	"math"
	"runtime"
	"slices"
	"time"

	"github.com/gascoigne-ai/uitoolkit/render"
)

// Input is one frame of user input.
type Input struct {
	X, Y        int
	Click       bool    // left button went down this frame
	MiddleClick bool    // middle button went down this frame
	MiddleHeld  bool    // middle button is down
	RightClick  bool    // right button went down this frame
	Wheel       float64 // vertical wheel movement, positive = up
	WheelX      float64 // horizontal wheel movement (a touchpad's two-finger swipe), positive = right
	Typed       []rune
	Edits       []Edit
	Enter       bool
	Escape      bool
	Tab         bool

	Held             bool // left button is down
	Shift, Ctrl, Alt bool
	// Ctrl+key shortcuts that went down this frame.
	SelectAll, Copy, Cut, Paste bool
	Undo, Redo                  bool
	Clicks                      int // 1, 2 or 3 for single, double, triple click (with Click)

	// Keys went down this frame, for application shortcuts.
	Keys []render.Key
}

// Scaled converts the pointer position from device to logical pixels.
func (in Input) Scaled(scale float64) Input {
	if !(scale > 0) || math.IsInf(scale, 0) { // int(x/0) is not a position
		return in
	}
	in.X = int(float64(in.X) / scale)
	in.Y = int(float64(in.Y) / scale)
	return in
}

// Pressed reports whether key went down this frame.
func (in Input) Pressed(key render.Key) bool { return slices.Contains(in.Keys, key) }

// Consume removes keys from this frame's input in every form they take
// (the Enter, Escape and Tab flags, the edit of an editing key such as Up
// or Backspace, the Keys list), so a widget under an overlay that used them
// does not act on them too: Enter that accepts a menu item must not type a
// line break.
func (in *Input) Consume(keys ...render.Key) {
	for _, key := range keys {
		switch key {
		case render.KeyEnter:
			in.Enter = false
		case render.KeyEscape:
			in.Escape = false
		case render.KeyTab:
			in.Tab = false
		}
		for _, editKey := range editKeys {
			if editKey.key == key {
				in.Edits = withoutEdit(in.Edits, editKey.edit)
			}
		}
	}
	remaining := make([]render.Key, 0, len(in.Keys))
	for _, down := range in.Keys {
		if !slices.Contains(keys, down) {
			remaining = append(remaining, down)
		}
	}
	in.Keys = remaining
}

func withoutEdit(edits []Edit, drop Edit) []Edit {
	kept := make([]Edit, 0, len(edits))
	for _, edit := range edits {
		if edit != drop {
			kept = append(kept, edit)
		}
	}
	return kept
}

// RestartsCaret reports whether this input shows a text caret again:
// typing, editing keys and clicks do; a pointer moving over the window
// does not.
func (in Input) RestartsCaret() bool {
	return len(in.Typed) > 0 || len(in.Keys) > 0 || len(in.Edits) > 0 || in.Click || in.Enter ||
		in.Paste || in.Cut || in.Undo || in.Redo || in.SelectAll
}

// Hit reports whether a click landed inside r.
func (in Input) Hit(rect Rect) bool { return in.Click && rect.Contains(in.X, in.Y) }

// Over reports whether the pointer is inside r.
func (in Input) Over(rect Rect) bool { return rect.Contains(in.X, in.Y) }

var editKeys = []struct {
	key  render.Key
	edit Edit
}{
	{render.KeyBackspace, EditBackspace},
	{render.KeyDelete, EditDelete},
	{render.KeyArrowLeft, EditLeft},
	{render.KeyArrowRight, EditRight},
	{render.KeyHome, EditHome},
	{render.KeyEnd, EditEnd},
	{render.KeyArrowUp, EditUp},
	{render.KeyArrowDown, EditDown},
	{render.KeyPageUp, EditPageUp},
	{render.KeyPageDown, EditPageDown},
}

// KeyRepeat is how a held key repeats: it fires once when pressed, again
// after Delay, then every Interval. The zero value is the built-in timing
// (25 ticks, then every 3), for when the system's own is unknown.
type KeyRepeat struct {
	Delay, Interval time.Duration
}

const (
	defaultRepeatDelayTicks    = 25
	defaultRepeatIntervalTicks = 3
)

// Fires reports whether a key held for the given number of ticks acts on
// this tick, at tps ticks a second. The timing is rounded to whole ticks
// and an interval is at least one tick.
func (repeat KeyRepeat) Fires(held, tps int) bool {
	delay, interval := defaultRepeatDelayTicks, defaultRepeatIntervalTicks
	if repeat.Delay > 0 && repeat.Interval > 0 {
		delay = max(ticks(repeat.Delay, tps), 1)
		interval = max(ticks(repeat.Interval, tps), 1)
	}
	return held == 1 || (held >= delay && (held-delay)%interval == 0)
}

func ticks(duration time.Duration, tps int) int {
	return int((duration*time.Duration(tps) + time.Second/2) / time.Second)
}

// inputFrom is what a window's report of one tick means on the platform goos
// with the key repeat timing repeat. A key went down this tick when it has been
// down for one; an aggregate (KeyShift...) never has, so Keys names physical
// keys only. Clicks are counted by the Sampler.
func inputFrom(state render.InputState, goos string, repeat KeyRepeat) Input {
	held := make(map[render.Key]int, len(state.Down))
	var keys []render.Key
	for _, down := range state.Down {
		held[down.Key] = down.Ticks
		if down.Ticks == 1 {
			keys = append(keys, down.Key)
		}
	}
	isDown := func(key render.Key) bool {
		_, ok := held[key]
		return ok
	}
	wentDown := func(key render.Key) bool { return held[key] == 1 }
	in := Input{
		X: state.X, Y: state.Y,
		Click:       state.ButtonsPressed&render.MouseLeft != 0,
		Wheel:       state.WheelY,
		WheelX:      state.WheelX,
		Typed:       state.Typed,
		Enter:       wentDown(render.KeyEnter),
		Escape:      wentDown(render.KeyEscape),
		Tab:         wentDown(render.KeyTab),
		Held:        state.Buttons&render.MouseLeft != 0,
		MiddleClick: state.ButtonsPressed&render.MouseMiddle != 0,
		MiddleHeld:  state.Buttons&render.MouseMiddle != 0,
		RightClick:  state.ButtonsPressed&render.MouseRight != 0,
		Keys:        keys,
		Shift:       isDown(render.KeyShift),
		Alt:         isDown(render.KeyAlt),
		Ctrl:        PrimaryDown(goos, isDown(render.KeyControl), isDown(render.KeyMeta)),
	}
	if in.Ctrl {
		in.SelectAll = wentDown(render.KeyA)
		in.Copy = wentDown(render.KeyC)
		in.Cut = wentDown(render.KeyX)
		in.Paste = wentDown(render.KeyV)
		undoKey := wentDown(render.KeyZ)
		in.Undo = undoKey && !in.Shift
		in.Redo = wentDown(render.KeyY) || (undoKey && in.Shift)
	}
	for _, editKey := range editKeys {
		// A key that is not down has no ticks, and Fires refuses zero.
		if !repeat.Fires(held[editKey.key], state.TPS) {
			continue
		}
		in.Edits = append(in.Edits, modifiedEdit(goos, editKey.edit, in.Ctrl, in.Alt))
	}
	return in
}

// wordEdits are the word-wise variants of movement and deletion keys,
// taken while the platform's word modifier is held.
var wordEdits = map[Edit]Edit{
	EditLeft:      EditWordLeft,
	EditRight:     EditWordRight,
	EditBackspace: EditDeleteWordLeft,
	EditDelete:    EditDeleteWordRight,
}

// docEdits are the document-wide variants of Home and End, taken with the
// primary modifier.
var docEdits = map[Edit]Edit{
	EditHome: EditDocStart,
	EditEnd:  EditDocEnd,
}

// macLineEdits are what Command does to the arrows and delete keys in the
// macOS text editors: it acts on the line, since Mac keyboards often have
// no Home and End keys. Elsewhere Control with these keys is word-wise.
var macLineEdits = map[Edit]Edit{
	EditLeft:      EditHome,
	EditRight:     EditEnd,
	EditBackspace: EditDeleteToLineStart,
	EditDelete:    EditDeleteToLineEnd,
}

// modifiedEdit is edit as the held modifiers turn it. Word movement and
// deletion follow each platform's text editors: Option on macOS, where
// Command is the shortcut key and Control belongs to the system, and
// Control elsewhere; Command on macOS makes them line edits and wins over
// Option. primary is the platform's shortcut modifier (PrimaryDown), alt
// the Alt (Option) key.
func modifiedEdit(goos string, edit Edit, primary, alt bool) Edit {
	wordModifier := primary
	if goos == "darwin" {
		wordModifier = alt
		// Return here: Command+Left is the line start, and going on to
		// docEdits would turn the Home it becomes into the document start.
		if line, ok := macLineEdits[edit]; ok && primary {
			return line
		}
	}
	if word, ok := wordEdits[edit]; ok && wordModifier {
		return word
	}
	if doc, ok := docEdits[edit]; ok && primary {
		return doc
	}
	return edit
}

// Multi-click detection: a click within this time and distance of the
// previous one extends the count (double, triple).
const (
	multiClickGap  = 400 * time.Millisecond
	multiClickDist = 4
)

// Sampler reads input each frame and counts rapid clicks.
type Sampler struct {
	// Now is the clock clicks are timed by; nil is time.Now.
	Now func() time.Time
	// Repeat is how held keys repeat; the zero value is the built-in timing.
	Repeat KeyRepeat
	// HostOS is the operating system the shortcut modifier follows; empty is
	// runtime.GOOS.
	HostOS string

	last  time.Time
	x, y  int
	count int
}

// Click registers a click at x, y at time now and returns its count
// (1, 2 or 3; a fourth quick click starts over at 1).
func (s *Sampler) Click(now time.Time, posX, posY int) int {
	near := abs(posX-s.x) <= multiClickDist && abs(posY-s.y) <= multiClickDist
	if s.count > 0 && now.Sub(s.last) <= multiClickGap && near && s.count < 3 {
		s.count++
	} else {
		s.count = 1
	}
	s.last, s.x, s.y = now, posX, posY
	return s.count
}

func (s *Sampler) hostOS() string {
	if s.HostOS != "" {
		return s.HostOS
	}
	return runtime.GOOS
}

func (s *Sampler) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// Sample is the input of the tick a window reported in state, including the
// click count. Call it once a tick, from Update.
func (s *Sampler) Sample(state render.InputState) Input {
	in := inputFrom(state, s.hostOS(), s.Repeat)
	if in.Click {
		in.Clicks = s.Click(s.now(), in.X, in.Y)
	}
	return in
}
