package theming

import (
	"fmt"
	"maps"
	"slices"

	"github.com/gascoigne-ai/uitoolkit/core"
)

// feelFile is the feel block of a theme file: the values of core.Feel by
// name, whole milliseconds for the durations. A pointer, so null is an error
// and not a 0.
type feelFile map[string]*int

// apply lays the block over feel, that of the base theme: a value it states
// replaces the base's, one it leaves out stays. core.Feel.Set holds the names
// and bounds.
func (file feelFile) apply(feel core.Feel) (core.Feel, error) {
	for _, name := range slices.Sorted(maps.Keys(file)) { // sorted, so the first error is always the same
		value := file[name]
		if value == nil {
			return feel, fmt.Errorf("%q needs a number", name)
		}
		set, err := feel.Set(name, *value)
		if err != nil {
			return feel, err
		}
		feel = set
	}
	return feel, nil
}
