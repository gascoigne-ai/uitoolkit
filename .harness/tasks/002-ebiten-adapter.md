status: todo
kind: feature
commit: feat(ebitenrender): add the Ebitengine adapter package
scope: ebitenrender/**
not_before: 2026-11-25
depends_on: 001-render-contract

# Implement the render contract over Ebitengine, inside this module

Goal: Package `ebitenrender` implements `Canvas`, `Image`, `Offscreen`, `Face` and `FontSource` over ebiten, `text/v2` and `vector`. It keeps `sliceQuads`, `ringQuads`, `buildCornerMask` and `buildRingMask` (from `core/cornermask.go`) and their tests unchanged in behavior, so the pixels stay identical, and the icon mask cache. It owns `readInput` (from `core/input.go`), the ebiten-to-`render.Key` map and a `Run(host, opts)` wrapper for `ebiten.RunGame` with the skip-clear mode and the cursor mapping the workshop has in `internal/app/game.go`. It lives in this module for now so every commit builds (task 008 splits it out).

## Model
Model: new package; the key name table is generated once from `ebiten.Key.String()` and frozen in a test. Invariant: every old name resolves to the same key, because users store shortcuts by these names (`engine/binding.go`).

## Acceptance
- Test first: key-name round trip over `ebiten.Key(0)..KeyMax`; offscreen fill then `ReadPixels` equals the current `core` output for rounded fills, rings and gradients.
- The repo's check command passes (`make check`, temp HOME/XDG).
