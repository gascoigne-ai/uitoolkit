status: todo
kind: feature
commit: refactor(view): draw the element tree through render.Canvas
scope: view/*
not_before: 2026-11-25
depends_on: 007-headless-text

# Move the view runtime onto render.Canvas

Goal: `view.Draw`, `DrawBelow` and `DrawFrom` take `core.Surface` and turn it into a canvas once with `ctx.CanvasOf`. drawSelf, drawAfter, kidsDst (through `ClipCanvas`), drawKids and culling by `kidsDst.Bounds()` work on `render.Canvas`. `Behavior`, `OverDrawer` and textB keep the old `*ebiten.Image` signature until 005 and are called with `ctx.EbitenImage(canvas)`; on a non-ebiten canvas they are skipped.

## Model
Model: unchanged except internal signatures. Invariant: the only ebiten in view is the Behavior, OverDrawer and textB signatures.

## Acceptance
- Test first: on a `rendertest.Canvas`, Bg, borders and the style outline record the expected ops, and a child outside a clipping parent's `Sub` bounds records none.
- Existing view tests still draw on `ebiten.NewImage` with the ebitenrender device.
- The repo's check command passes (`make check`, temp HOME/XDG).
