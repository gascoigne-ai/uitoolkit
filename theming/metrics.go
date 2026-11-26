package theming

import (
	"fmt"
	"maps"
	"slices"

	"github.com/gascoigne-ai/uitoolkit/core"
)

// metricsFile is the metrics block of a theme file. A knob is a pointer, so
// one the file leaves out (the base's stays) differs from an explicit 0 (an
// error for the unit and a step).
type metricsFile struct {
	Unit    *int       `json:"unit"`
	Leading *int       `json:"leading"`
	Steps   *stepsFile `json:"steps"`
	// Sizes pins named sizes by their template names. A pointer, so null is
	// an error and not a pin at 0.
	Sizes map[string]*int `json:"sizes"`
}

type stepsFile struct {
	Hair   *int `json:"hair"`
	Near   *int `json:"near"`
	Tight  *int `json:"tight"`
	Loose  *int `json:"loose"`
	Label  *int `json:"label"`
	Gutter *int `json:"gutter"`
	Dialog *int `json:"dialog"`
}

// knob is one number a metrics block may state, its bounds and where it goes.
type knob struct {
	name      string
	stated    *int
	low, high int
	target    *int
}

// apply lays the block over metrics, those of the base theme: a knob it
// states replaces the base's, and its pins merge with the base's by name.
func (file metricsFile) apply(metrics core.Metrics) (core.Metrics, error) {
	steps := file.Steps
	if steps == nil {
		steps = &stepsFile{}
	}
	step := func(name string, stated, target *int) knob {
		return knob{fmt.Sprintf("step %q", name), stated, core.MinStep, core.MaxStep, target}
	}
	for _, entry := range []knob{
		{"unit", file.Unit, core.MinUnit, core.MaxUnit, &metrics.Unit},
		{"leading", file.Leading, core.MinLeading, core.MaxLeading, &metrics.Leading},
		step("hair", steps.Hair, &metrics.Steps.Hair),
		step("near", steps.Near, &metrics.Steps.Near),
		step("tight", steps.Tight, &metrics.Steps.Tight),
		step("loose", steps.Loose, &metrics.Steps.Loose),
		step("label", steps.Label, &metrics.Steps.Label),
		step("gutter", steps.Gutter, &metrics.Steps.Gutter),
		step("dialog", steps.Dialog, &metrics.Steps.Dialog),
	} {
		if entry.stated == nil {
			continue
		}
		if *entry.stated < entry.low || *entry.stated > entry.high {
			return metrics, fmt.Errorf("%s must be within %d and %d", entry.name, entry.low, entry.high)
		}
		*entry.target = *entry.stated
	}
	for _, name := range slices.Sorted(maps.Keys(file.Sizes)) { // sorted, so the first error is always the same
		pixels := file.Sizes[name]
		if pixels == nil {
			return metrics, fmt.Errorf("size %q needs a number", name)
		}
		pinned, err := metrics.Pin(name, *pixels)
		if err != nil {
			return metrics, err
		}
		metrics = pinned
	}
	return metrics, nil
}
