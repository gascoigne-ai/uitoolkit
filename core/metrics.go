package core

import (
	"fmt"
	"maps"
)

// Metrics are the size rules a theme sets: a few knobs every size is derived
// from, and the named sizes it fixes itself (Pin). A zero field is its
// default, so a Theme that says nothing has the default sizes. SizesFor
// turns Metrics into pixels for one pair of line heights.
type Metrics struct {
	// Unit is the grid in pixels every derived padding and gap is a
	// multiple of, 1 to 4; 0 is 2.
	Unit int
	// Leading is extra pixels of leading, 0 to 8, added to the measured UI
	// and small line. It applies to the height classes, the bars built from
	// them, and the space between the wrapped lines of widget.Paragraph. A
	// single-line text and the checkbox keep the measured line; code and
	// text area lines have their own spacing, which Leading does not move.
	// It is added rather than scaled: it stays a whole pixel, every rule
	// keeps the form "line plus a padding", and both faces get the same
	// amount whatever text size the user picks. A theme cannot set font
	// sizes, so this is how it airs out its lines.
	Leading int
	// Steps are the spacing steps in Units, 1 to 16 each; a 0 step is its
	// default (1, 2, 3, 4, 5, 6, 8).
	Steps Steps

	// pins are the named sizes the theme fixes; an absent one is derived.
	// Only Pin writes it, always into a clone, so copies of a Theme never
	// share a change. nil until the first pin. A map,
	// since it is sparse and read once per resolve; rejected: an array and a
	// presence bitmask. Revisit if Metrics must be comparable.
	pins map[sizeID]int
}

// Steps are the spacing steps, from tight to wide. Every spacing role of
// Sizes is one of them, so two roles on one step stay equal when it moves.
// They count Units in Metrics and pixels in Sizes.
type Steps struct{ Hair, Near, Tight, Loose, Label, Gutter, Dialog int }

// The bounds of the knobs of Metrics; a step is in Units. A theme file
// outside them is rejected, and SizesFor clamps a Go literal into them. A
// theme cannot move them: a theme file is untrusted input.
const (
	MinUnit, MaxUnit       = 1, 4
	MinLeading, MaxLeading = 0, 8
	MinStep, MaxStep       = 1, 16
)

const defaultUnit = 2

var defaultSteps = Steps{Hair: 1, Near: 2, Tight: 3, Loose: 4, Label: 5, Gutter: 6, Dialog: 8}

// Pin is a copy of m that fixes the named size at px pixels, whatever the
// knobs and the faces say. The name is the one templates use ("rowH",
// "gutter"), see Sizes.Scope. Sizes derived from a pinned one follow it. An
// unknown name or a px outside the size's bounds is an error.
func (m Metrics) Pin(name string, px int) (Metrics, error) {
	for id, rule := range sizeRules {
		if rule.name != name {
			continue
		}
		if px < rule.min || px > rule.max {
			return m, fmt.Errorf("size %q must be within %d and %d", name, rule.min, rule.max)
		}
		pins := maps.Clone(m.pins)
		if pins == nil {
			pins = map[sizeID]int{}
		}
		pins[sizeID(id)] = px
		m.pins = pins
		return m, nil
	}
	return m, fmt.Errorf("unknown size %q", name)
}

// unit is the grid in pixels, the default for 0, clamped to its bounds.
func (m Metrics) unit() int {
	if m.Unit == 0 {
		return defaultUnit
	}
	return min(max(m.Unit, MinUnit), MaxUnit)
}

// stepPixels are the steps in pixels.
func (m Metrics) stepPixels() Steps {
	unit := m.unit()
	pixels := func(units, fallback int) int {
		if units == 0 {
			units = fallback
		}
		return min(max(units, MinStep), MaxStep) * unit
	}
	return Steps{
		Hair:   pixels(m.Steps.Hair, defaultSteps.Hair),
		Near:   pixels(m.Steps.Near, defaultSteps.Near),
		Tight:  pixels(m.Steps.Tight, defaultSteps.Tight),
		Loose:  pixels(m.Steps.Loose, defaultSteps.Loose),
		Label:  pixels(m.Steps.Label, defaultSteps.Label),
		Gutter: pixels(m.Steps.Gutter, defaultSteps.Gutter),
		Dialog: pixels(m.Steps.Dialog, defaultSteps.Dialog),
	}
}
