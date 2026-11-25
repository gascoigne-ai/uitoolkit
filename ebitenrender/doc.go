// Package ebitenrender implements the render contract over Ebitengine, with
// its vector and text/v2 packages: Device makes the canvases, offscreens,
// images and font sources, Run runs a render.Host in a window, and ReadInput
// reports a tick's pointer, wheel, characters and keys. Of this module it
// imports render only.
//
// ReadInput only reports states and durations. What they mean (clicks, edits,
// shortcuts) is decided in core, so that every renderer behaves the same.
//
// DrawText is one plain pass; core.Ctx.Text expands tabs and draws a second
// pass for weight.
package ebitenrender
