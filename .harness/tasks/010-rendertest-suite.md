status: todo
kind: feature
commit: test(render): add the cross-renderer golden suite
scope: render/rendertest/**
not_before: 2026-11-25
depends_on: 009-showcase

# One golden-image suite every renderer must pass

Goal: Scenes for every `Canvas` op (fills, gradients, rounded corners, rings, AA lines, paths with caps and joins, clip via `Sub`, text, images, offscreen) plus the showcase sections. The ebiten renderer is the reference; a renderer passes with a per-pixel tolerance it declares.

## Model
Model: a scene is a name and a draw func; goldens are PNG files under `render/rendertest/testdata`.

## Acceptance
- Test first: the suite runs against `renderer/ebiten` with tolerance 0.
- The repo's check command passes (`make check`, temp HOME/XDG).
