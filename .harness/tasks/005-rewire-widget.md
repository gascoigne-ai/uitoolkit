status: todo
kind: feature
commit: refactor(widget): draw widgets and behaviors through render.Canvas
scope: widget/*,view/*
not_before: 2026-11-25
depends_on: 004-rewire-view

# Move widget and the Behavior interface onto render.Canvas

Goal: `Behavior.Draw`, `Base.Draw` and `OverDrawer.DrawOver` take `render.Canvas`; view stops calling `EbitenImage` (view changes are only this signature: custom.go, leaf.go, runtime.go, layers.go, test probes and helpers; user accepted the scope). Every `*ebiten.Image` in widget becomes `render.Canvas`; `Clip` becomes `ClipCanvas`. icon.go uses `render.Path`, `FillPath`/`StrokePath`, `DrawMask`; the package-level mask cache goes, and the design note decides its new owner (device offscreens or none) with a bench. svg.go caps and joins become render types; a zero MiterLimit keeps today's bevel. `widget.Image(render.Image)` draws with `render.FilterLinear` and treats a nil pointer in the interface as no picture. Markup may still pass `*ebiten.Image` until 006. View and widget tests move to `rendertest`, with faces from `render/text`.

## Model
Model: unchanged except signatures. Invariant: no package-level GPU caches in widget.

## Acceptance
- Test first: clipskip and icon tests on the fake canvas; `Image` with a typed-nil picture measures 0x0.
- `go list -f '{{.Imports}} {{.TestImports}} {{.XTestImports}}' ./view ./widget` shows no `hajimehoshi` (the transitive check is in 006).
- The repo's check command passes (`make check`, temp HOME/XDG).
