status: todo
kind: feature
commit: feat(wgpu): images, offscreen and path rendering
scope: renderer/wgpu/**
not_before: 2026-11-25
depends_on: 012-wgpu-pipelines

# wgpu renderer stage d and e: textures and paths

Goal: Textured quads, `Image` upload, `Offscreen` with `ReadPixels`. Paths: all 28 icons use strokes with caps and joins and `x/image/vector` only fills, so this task adds a stroker (**asks the user first** if it needs a new dependency) and rasterizes to a cached texture. Icon cache lives in the renderer.

Decision (user, 2026-11-26): the stroker is decided when this task is reached: compare open-source options and propose three, then wait for the pick.

## Model
Model: texture cache keyed by device; dropped on device loss.

## Acceptance
- Test first: `rendertest` image, offscreen and path scenes pass.
- The repo's check command passes (`make check`, temp HOME/XDG).
