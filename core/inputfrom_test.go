package core

import (
	"reflect"
	"testing"
	"time"

	"github.com/gascoigne-ai/uitoolkit/render"
)

func TestInputFromReadsAWindowsReport(t *testing.T) {
	down := func(key render.Key, ticks int) render.KeyDown { return render.KeyDown{Key: key, Ticks: ticks} }
	// ebiten reports a held modifier twice: the physical key, timed, and the
	// aggregate, which has no duration.
	control := []render.KeyDown{down(render.KeyControlLeft, 5), down(render.KeyControl, 0)}
	shift := []render.KeyDown{down(render.KeyShiftLeft, 5), down(render.KeyShift, 0)}
	alt := []render.KeyDown{down(render.KeyAltLeft, 5), down(render.KeyAlt, 0)}
	meta := []render.KeyDown{down(render.KeyMetaLeft, 5), down(render.KeyMeta, 0)}
	keys := func(groups ...[]render.KeyDown) render.InputState {
		state := render.InputState{TPS: 60}
		for _, group := range groups {
			state.Down = append(state.Down, group...)
		}
		return state
	}
	one := func(key render.Key, ticks int) []render.KeyDown { return []render.KeyDown{down(key, ticks)} }
	custom := Repeat{Delay: 500 * time.Millisecond, Interval: 100 * time.Millisecond}
	tests := []struct {
		name   string
		goos   string
		repeat Repeat
		state  render.InputState
		want   Input
	}{
		{name: "nothing happens", goos: "linux", state: render.InputState{TPS: 60}},
		{name: "pointer, wheel and typed characters", goos: "linux",
			state: render.InputState{X: 3, Y: 4, WheelX: 1.5, WheelY: -2, Typed: []rune("aż")},
			want:  Input{X: 3, Y: 4, WheelX: 1.5, Wheel: -2, Typed: []rune("aż")}},
		{name: "the left button goes down", goos: "linux",
			state: render.InputState{Buttons: render.MouseLeft, ButtonsPressed: render.MouseLeft},
			want:  Input{Click: true, Held: true}},
		{name: "the left button stays down", goos: "linux",
			state: render.InputState{Buttons: render.MouseLeft},
			want:  Input{Held: true}},
		{name: "the middle button goes down", goos: "linux",
			state: render.InputState{Buttons: render.MouseMiddle, ButtonsPressed: render.MouseMiddle},
			want:  Input{MiddleClick: true, MiddleHeld: true}},
		{name: "the right button goes down", goos: "linux",
			state: render.InputState{Buttons: render.MouseRight, ButtonsPressed: render.MouseRight},
			want:  Input{RightClick: true}},
		{name: "enter, escape and tab go down", goos: "linux",
			state: keys(one(render.KeyEnter, 1), one(render.KeyEscape, 1), one(render.KeyTab, 1)),
			want:  Input{Enter: true, Escape: true, Tab: true, Keys: []render.Key{render.KeyEnter, render.KeyEscape, render.KeyTab}}},
		{name: "a held key does not go down again", goos: "linux",
			state: keys(one(render.KeyEnter, 2), one(render.KeyA, 30))},
		{name: "shift and alt are held", goos: "linux",
			state: keys(shift, alt),
			want:  Input{Shift: true, Alt: true}},
		{name: "a modifier that goes down is a key, its aggregate is not", goos: "linux",
			state: keys(one(render.KeyControlLeft, 1), one(render.KeyControl, 0)),
			want:  Input{Ctrl: true, Keys: []render.Key{render.KeyControlLeft}}},
		{name: "ctrl+a selects all", goos: "linux",
			state: keys(one(render.KeyA, 1), control),
			want:  Input{Ctrl: true, SelectAll: true, Keys: []render.Key{render.KeyA}}},
		{name: "ctrl with a held a selects nothing", goos: "linux",
			state: keys(one(render.KeyA, 2), control),
			want:  Input{Ctrl: true}},
		{name: "ctrl+c copies", goos: "linux",
			state: keys(one(render.KeyC, 1), control),
			want:  Input{Ctrl: true, Copy: true, Keys: []render.Key{render.KeyC}}},
		{name: "ctrl+x cuts and ctrl+v pastes", goos: "windows",
			state: keys(one(render.KeyV, 1), one(render.KeyX, 1), control),
			want:  Input{Ctrl: true, Cut: true, Paste: true, Keys: []render.Key{render.KeyV, render.KeyX}}},
		{name: "ctrl+z undoes", goos: "linux",
			state: keys(one(render.KeyZ, 1), control),
			want:  Input{Ctrl: true, Undo: true, Keys: []render.Key{render.KeyZ}}},
		{name: "ctrl+shift+z redoes", goos: "linux",
			state: keys(one(render.KeyZ, 1), control, shift),
			want:  Input{Ctrl: true, Shift: true, Redo: true, Keys: []render.Key{render.KeyZ}}},
		{name: "ctrl+y redoes", goos: "linux",
			state: keys(one(render.KeyY, 1), control),
			want:  Input{Ctrl: true, Redo: true, Keys: []render.Key{render.KeyY}}},
		{name: "a plain c is no shortcut", goos: "linux",
			state: keys(one(render.KeyC, 1)),
			want:  Input{Keys: []render.Key{render.KeyC}}},
		{name: "command is the shortcut key on darwin", goos: "darwin",
			state: keys(one(render.KeyC, 1), meta),
			want:  Input{Ctrl: true, Copy: true, Keys: []render.Key{render.KeyC}}},
		{name: "the windows key is no shortcut key", goos: "linux",
			state: keys(one(render.KeyC, 1), meta),
			want:  Input{Keys: []render.Key{render.KeyC}}},
		{name: "ctrl+left moves a word", goos: "linux",
			state: keys(one(render.KeyArrowLeft, 1), control),
			want:  Input{Ctrl: true, Edits: []Edit{EditWordLeft}, Keys: []render.Key{render.KeyArrowLeft}}},
		{name: "option+left moves a word on darwin", goos: "darwin",
			state: keys(one(render.KeyArrowLeft, 1), alt),
			want:  Input{Alt: true, Edits: []Edit{EditWordLeft}, Keys: []render.Key{render.KeyArrowLeft}}},
		{name: "command+backspace deletes to the line start on darwin", goos: "darwin",
			state: keys(one(render.KeyBackspace, 1), meta),
			want:  Input{Ctrl: true, Edits: []Edit{EditDeleteToLineStart}, Keys: []render.Key{render.KeyBackspace}}},
		{name: "edits come in the order of the edit keys", goos: "linux",
			state: keys(one(render.KeyArrowDown, 1), one(render.KeyBackspace, 1)),
			want:  Input{Edits: []Edit{EditBackspace, EditDown}, Keys: []render.Key{render.KeyArrowDown, render.KeyBackspace}}},
		{name: "a held editing key waits for the repeat delay", goos: "linux",
			state: keys(one(render.KeyBackspace, 24))},
		{name: "a held editing key repeats after the delay", goos: "linux",
			state: keys(one(render.KeyBackspace, 25)),
			want:  Input{Edits: []Edit{EditBackspace}}},
		{name: "between two repeats nothing happens", goos: "linux",
			state: keys(one(render.KeyBackspace, 27))},
		{name: "a held editing key repeats at the interval", goos: "linux",
			state: keys(one(render.KeyBackspace, 28)),
			want:  Input{Edits: []Edit{EditBackspace}}},
		{name: "the system's repeat delay is counted in ticks", goos: "linux", repeat: custom,
			state: keys(one(render.KeyDelete, 30)),
			want:  Input{Edits: []Edit{EditDelete}}},
		{name: "the system's repeat delay follows the tick rate", goos: "linux", repeat: custom,
			state: render.InputState{TPS: 30, Down: one(render.KeyDelete, 15)},
			want:  Input{Edits: []Edit{EditDelete}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.want.TPS = test.state.TPS // always passed on
			if got := inputFrom(test.state, test.goos, test.repeat); !reflect.DeepEqual(got, test.want) {
				t.Errorf("inputFrom = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestSampleCountsTheClicksOfWhatItIsGiven(t *testing.T) {
	now := time.Unix(100, 0)
	sampler := &Sampler{Now: func() time.Time { return now }, HostOS: "linux"}
	click := render.InputState{X: 5, Y: 5, Buttons: render.MouseLeft, ButtonsPressed: render.MouseLeft}
	for want := 1; want <= 2; want++ {
		if got := sampler.Sample(click, Feel{}); got.Clicks != want || !got.Click {
			t.Fatalf("click %d: Sample = %+v", want, got)
		}
	}
	if got := sampler.Sample(render.InputState{X: 5, Y: 5}, Feel{}); got.Clicks != 0 {
		t.Fatalf("no click: Sample = %+v, want no click count", got)
	}
}
