package view

import "github.com/gascoigne-ai/uitoolkit/core"

// sizeSlot is an attribute of an element that can follow the context's
// Sizes. El.sized is an array indexed by it: options write it, Layout reads
// it once, and it needs no allocation. Rejected: a named func field per
// attribute (the same size, more code) and a slice of pairs (allocates).
// Revisit if the size of El shows in the allocation benchmarks: move the
// array to a block allocated on first use.
type sizeSlot uint8

const (
	slotW sizeSlot = iota
	slotH
	slotPadL
	slotPadT
	slotPadR
	slotPadB
	slotGap
	slotMinW
	slotBorderInset
	slotFloatX
	slotFloatY
	slotCount
)

// slotFields is where each slot's size goes.
var slotFields = [slotCount]func(element *El) *int{
	slotW:           func(element *El) *int { return &element.w },
	slotH:           func(element *El) *int { return &element.h },
	slotPadL:        func(element *El) *int { return &element.padL },
	slotPadT:        func(element *El) *int { return &element.padT },
	slotPadR:        func(element *El) *int { return &element.padR },
	slotPadB:        func(element *El) *int { return &element.padB },
	slotGap:         func(element *El) *int { return &element.gap },
	slotMinW:        func(element *El) *int { return &element.minW },
	slotBorderInset: func(element *El) *int { return &element.borderInset },
	slotFloatX:      func(element *El) *int { return &element.fl.x },
	slotFloatY:      func(element *El) *int { return &element.fl.y },
}

// follow makes the slots take size from the context's Sizes at every Layout.
// Until the first one they hold the default size.
func (e *El) follow(size func(sizes core.Sizes) int, slots ...sizeSlot) {
	pixels := size(core.DefaultSizes())
	for _, slot := range slots {
		e.sized[slot], *slotFields[slot](e) = size, pixels
	}
}

// literal drops the roles of the slots: a fixed value given after a role
// replaces it.
func (e *El) literal(slots ...sizeSlot) {
	for _, slot := range slots {
		e.sized[slot] = nil
	}
}

func followOption(size func(sizes core.Sizes) int, slots ...sizeSlot) Option {
	return func(element *El) { element.follow(size, slots...) }
}

// PadOf is Pad for a padding that follows the context's Sizes, as HOf does
// for a height: PadOf(core.GutterOf). Like HOf it has no template
// attribute; a template writes pad="{{gutter}}".
func PadOf(size func(sizes core.Sizes) int) Option {
	return followOption(size, slotPadL, slotPadT, slotPadR, slotPadB)
}

// PadXOf is PadX that follows the context's Sizes. It takes a Go func, so
// it has no template attribute (see PadOf).
func PadXOf(size func(sizes core.Sizes) int) Option { return followOption(size, slotPadL, slotPadR) }

// PadYOf is PadY that follows the context's Sizes. It takes a Go func, so
// it has no template attribute (see PadOf).
func PadYOf(size func(sizes core.Sizes) int) Option { return followOption(size, slotPadT, slotPadB) }

// PadLOf is PadL that follows the context's Sizes. It takes a Go func, so
// it has no template attribute (see PadOf).
func PadLOf(size func(sizes core.Sizes) int) Option { return followOption(size, slotPadL) }

// PadTOf is PadT that follows the context's Sizes. It takes a Go func, so
// it has no template attribute (see PadOf).
func PadTOf(size func(sizes core.Sizes) int) Option { return followOption(size, slotPadT) }

// PadROf is PadR that follows the context's Sizes. It takes a Go func, so
// it has no template attribute (see PadOf).
func PadROf(size func(sizes core.Sizes) int) Option { return followOption(size, slotPadR) }

// PadBOf is PadB that follows the context's Sizes. It takes a Go func, so
// it has no template attribute (see PadOf).
func PadBOf(size func(sizes core.Sizes) int) Option { return followOption(size, slotPadB) }

// GapOf is Gap that follows the context's Sizes, as PadOf does for a
// padding. Like it, it has no template attribute: gap="{{looseGap}}".
func GapOf(size func(sizes core.Sizes) int) Option { return followOption(size, slotGap) }

// MinWOf is MinW that follows the context's Sizes, as PadOf does for a
// padding. Like it, it has no template attribute: min-w="{{rowH}}".
func MinWOf(size func(sizes core.Sizes) int) Option { return followOption(size, slotMinW) }

// FloatAtOf is FloatAt for a point that follows the context's Sizes, such as
// a corner kept a gutter off the window's edge. Like HOf it has no template
// attribute.
func FloatAtOf(posX, posY func(sizes core.Sizes) int) Option {
	return func(element *El) {
		element.fl.on, element.fl.mode = true, floatPoint
		element.follow(posX, slotFloatX)
		element.follow(posY, slotFloatY)
	}
}
