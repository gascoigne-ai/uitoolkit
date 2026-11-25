// Package ebitenrender implements the render contract over Ebitengine, with
// its vector and text/v2 packages: Device makes the canvases, offscreens,
// images and font sources, Run runs a render.Host in a window, and ReadInput
// reports a tick's pointer, wheel, characters and keys. Of the toolkit it
// imports render only.
//
// The package lives in its own module, github.com/gascoigne-ai/uitoolkit/renderer/ebiten,
// so that the toolkit module needs no Ebitengine. Its name is not the last
// element of its path on purpose: a package named ebiten would clash with
// github.com/hajimehoshi/ebiten/v2 in every file that imports both, which
// every program that uses Ebitengine itself does.
//
// ReadInput only reports states and durations. What they mean (clicks, edits,
// shortcuts) is decided in core, so that every renderer behaves the same.
//
// DrawText is one plain pass; core.Ctx.Text expands tabs and draws a second
// pass for weight.
package ebitenrender
