// Package ebitenrender implements the render contract over Ebitengine, with
// its vector and text/v2 packages: Device makes the canvases, offscreens,
// images and font sources, and Run runs a render.Host in a window. Of this
// module it imports render only.
//
// Input is not read here. Of the keys and cursors the package ships only the
// maps (KeyFromEbiten, KeyToEbiten and the cursor shapes Run applies).
// Reading a frame's input stays in core for now: it fills core's own input
// types, which this package must not import, so it can move only when core
// takes its input through the render contract (task 003).
//
// From the same device pixels a canvas draws the shapes and lines core.Ctx
// draws today; the parity tests require the same bytes. One shape differs: a
// square gradient on a box between pixels, which core fills aliased and a
// canvas anti-aliased. Text is not compared with core: DrawText is one plain
// pass, while core.Ctx.Text expands tabs and draws a second pass.
package ebitenrender
