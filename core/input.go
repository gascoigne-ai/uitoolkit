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

	// TPS is the number of ticks a second the window reported, 0 when it
	// reported none. Ticks converts a duration by it.
	TPS int
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

// Repeat is how something held repeats, a key or a button: it acts when
// pressed, again after Delay, then every Interval. It is known only when both
// are above zero. The zero value, like any not known, stands for "no timing":
// Sampler.Repeat then leaves keys to the theme (Feel.KeyRepeat), and Fires and
// Due use the built-in key repeat.
type Repeat struct {
	Delay, Interval time.Duration
}

func (repeat Repeat) known() bool { return repeat.Delay > 0 && repeat.Interval > 0 }

// Fires reports whether a key held for the given number of ticks acts on
// this tick, at tps ticks a second (0 or less means 60): on the press, and
// then whenever it is Due.
func (repeat Repeat) Fires(held, tps int) bool { return held == 1 || repeat.Due(held, tps) }

// Due is Fires without the press: whether a hold of held ticks repeats on
// this tick. The timing is rounded to whole ticks, and the delay and the
// interval are at least one tick, so nothing repeats twice in a tick.
func (repeat Repeat) Due(held, tps int) bool {
	if !repeat.known() {
		repeat = Feel{}.Resolved().KeyRepeat
	}
	delay := max(ticks(repeat.Delay, tps), 1)
	interval := max(ticks(repeat.Interval, tps), 1)
	return held >= delay && (held-delay)%interval == 0
}

const (
	// nominalTPS is the tick rate assumed when a window reports none.
	nominalTPS = 60
	// maxTicks is the longest count ticks gives, over a year at nominalTPS.
	maxTicks = math.MaxInt32
)

// ticks is duration rounded to whole ticks at tps ticks a second (0 or less
// means nominalTPS): 0 for a duration of 0 or less, and at most maxTicks.
func ticks(duration time.Duration, tps int) int {
	if tps <= 0 {
		tps = nominalTPS
	}
	if duration <= 0 {
		return 0
	}
	if duration > (math.MaxInt64-time.Second/2)/time.Duration(tps) {
		return maxTicks
	}
	return int(min((duration*time.Duration(tps)+time.Second/2)/time.Second, maxTicks))
}

// Ticks is duration in the ticks of this input, rounded to whole ones.
// Widgets count a delay in ticks (see Ctx.Tick) so that a run of ticks gives
// the same result whatever the clock does.
func (in Input) Ticks(duration time.Duration) int { return ticks(duration, in.TPS) }

// inputFrom is what a window's report of one tick means on the platform goos
// with the key repeat timing repeat. A key went down this tick when it has been
// down for one; an aggregate (KeyShift...) never has, so Keys names physical
// keys only. Clicks are counted by the Sampler.
func inputFrom(state render.InputState, goos string, repeat Repeat) Input {
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
		TPS:         state.TPS,
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

// Sampler reads input each frame and counts rapid clicks.
type Sampler struct {
	// Now is the clock clicks are timed by; nil is time.Now.
	Now func() time.Time
	// Repeat is the system's own key repeat, as desktop/keyrepeat reads it.
	// When both of its fields are above zero it wins over the theme;
	// otherwise held keys repeat by the Feel passed to Sample.
	Repeat Repeat
	// HostOS is the operating system the shortcut modifier follows; empty is
	// runtime.GOOS.
	HostOS string

	last  time.Time
	x, y  int
	count int
}

// Click registers a click at x, y at time now and returns its count
// (1, 2 or 3; a fourth quick click starts over at 1). A click counts on when
// it is within feel's MultiClickTime and, on each axis, MultiClickDistance of
// the one before; feel's zero value is the built-in.
func (s *Sampler) Click(now time.Time, posX, posY int, feel Feel) int {
	feel = feel.Resolved()
	near := abs(posX-s.x) <= feel.MultiClickDistance && abs(posY-s.y) <= feel.MultiClickDistance
	if s.count > 0 && now.Sub(s.last) <= feel.MultiClickTime && near && s.count < 3 {
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
// click count. feel is the theme's, normally ctx.Feel(): it times clicks, and
// held keys when Repeat is not known. Call it once a tick, from Update.
func (s *Sampler) Sample(state render.InputState, feel Feel) Input {
	feel = feel.Resolved()
	repeat := s.Repeat
	if !repeat.known() {
		repeat = feel.KeyRepeat
	}
	in := inputFrom(state, s.hostOS(), repeat)
	if in.Click {
		in.Clicks = s.Click(s.now(), in.X, in.Y, feel)
	}
	return in
}
