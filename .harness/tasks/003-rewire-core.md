status: todo
kind: feature
commit: refactor(core): draw through render.Canvas and a Device
scope: core/**
not_before: 2026-11-25
depends_on: 002-ebiten-adapter

# Move core.Ctx onto the render contract

Goal: `Ctx` draw methods take `render.Canvas`; fonts come from a `render.Device` / font loader that `NewCtx`, `SetScale` and `SetTheme` use (`NewCtx` has 94 call sites: list them in the design note and change them in this task). `Input.Keys`, `Pressed`, `Consume`, `editKeys` use `render.Key`. Corner and ring mask code and `gradientSrc` are gone from `core` (the adapter owns them); `Ctx` asks `FillStyledCorners` and the like. Width, Advance, LineH, Ascent use `render.Face`.

## Model
Model: `core.Fonts` holds `render.Face` values; caches `widthKey`/`faceKey` keyed by the comparable `Face`. Invariant: `core` imports no ebiten.

## Acceptance
- Test first: existing `core` tests ported to the fake canvas and a headless face (task 007 gives real metrics; until then a fixed-advance test face); `go list -deps ./core | grep -c hajimehoshi` is 0.
- Workshop-visible goldens unchanged for the ebiten adapter.
- The repo's check command passes (`make check`, temp HOME/XDG).
