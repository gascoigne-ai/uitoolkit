status: todo
kind: feature
commit: refactor(view): draw through render.Canvas
scope: view/**
not_before: 2026-11-25
depends_on: 003-rewire-core

# Move view onto the render contract

Goal: `Behavior.Draw`, `OverDrawer.DrawOver`, `Draw`, `DrawBelow`, `DrawFrom`, `FontSource` use `render` types. Culling via `Bounds()` of the `Sub` canvas keeps working (`view/runtime.go`).

## Model
Model: unchanged except signatures. Invariant: `view` imports `core` and `render` only.

## Acceptance
- Test first: `view` tests run on the fake canvas with no ebiten import.
- The repo's check command passes (`make check`, temp HOME/XDG).
