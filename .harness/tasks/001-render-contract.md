status: todo
kind: feature
commit: feat(render): add the renderer contract package
scope: render/**
not_before: 2026-11-25

# The render package: Canvas, Image, Face, Path, Key, Host

Goal: Add `render`, a package that imports no GPU library and says what the toolkit draws through. Nothing uses it yet. Design note first (Opus, read-only), then code. Review findings that shape it: `Sub` keeps the parent coordinate space and only clips (widgets draw in absolute coordinates: `widget/field.go`, `widget/textarea.go`, `view/runtime.go`); `Canvas` works in device pixels, snapping stays in `Ctx`; no `DrawTriangles`, use `FillRectGradient`; rounded fill, rounded ring outline and AA line are canvas ops; `Offscreen` is a `Canvas` and an `Image` with `Clear`, `Release`, `ReadPixels`; `Face` is comparable (it is a cache map key) and `FontSource` has `Face(size)`; `Key` keeps the aggregate keys (`KeyControl/Shift/Alt/Meta`) and the left/right ones; `Host.Draw` returns whether it drew; use `image.Rectangle` and `color.RGBA`, never `core.Rect`, so there is no import cycle.

## Model
Model: new types only; invariants: `render` imports only the standard library; `Face` usable as a map key; `Sub(r).Bounds()` is the intersection of the parent bounds and r, with no translation. Collections: none.

## Acceptance
- Test first: a fake in-memory `Canvas` in `render/rendertest` records ops; tests state `Sub` intersection, `Face` comparability and `Key` name table shape.
- `go list -deps ./render` shows no third-party package.
- The repo's check command passes (`make check`, temp HOME/XDG).
