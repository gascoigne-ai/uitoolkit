package core

import "github.com/gascoigne-ai/uitoolkit/render"

// MaskKey names a cached mask: what it shows, at which scale and weight.
// Source identifies the shape drawn (widget uses a pointer to the parsed icon,
// which stays the same until the icon is replaced); it must be comparable.
type MaskKey struct {
	Source        any
	Scale, Weight float64
}

// maskLimit bounds the masks of a Ctx. When the cache is full it is dropped
// whole, and what is drawn next fills it again.
const maskLimit = 512

// Mask is the offscreen of width by height pixels that paint drew on, for
// key: made once, then returned until the cache is dropped. Draw it with
// Canvas.DrawMask to tint it. A mask lives in the Ctx, not in a package, so
// it belongs to the Device that made it and goes with SetScale and SetTheme,
// which change what the same key would paint.
func (c *Ctx) Mask(key MaskKey, width, height int, paint func(render.Canvas)) render.Offscreen {
	if mask, ok := c.masks[key]; ok {
		return mask
	}
	if len(c.masks) >= maskLimit {
		c.dropMasks()
	}
	mask := c.Device.NewOffscreen(max(width, 1), max(height, 1))
	paint(mask)
	if c.masks == nil {
		c.masks = map[MaskKey]render.Offscreen{}
	}
	c.masks[key] = mask
	return mask
}

// MaskCount is the number of masks the Ctx holds.
func (c *Ctx) MaskCount() int { return len(c.masks) }

// dropMasks releases every mask.
func (c *Ctx) dropMasks() {
	for _, mask := range c.masks {
		mask.Release()
	}
	c.masks = nil
}
