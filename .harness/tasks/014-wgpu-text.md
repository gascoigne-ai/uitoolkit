status: todo
kind: feature
commit: feat(wgpu): glyph atlas text
scope: renderer/wgpu/**
not_before: 2026-11-25
depends_on: 013-wgpu-images-paths

# wgpu renderer stage f: text

Goal: Shape with `render/text` (task 007), rasterize glyphs to an R8 atlas (shelf packer, LRU eviction, 8 subpixel-x bins, whole-pixel y, keyed by face, size and scale); stem boost as a coverage-gamma uniform. Tabs are pre-expanded by `tabsAsGaps`.

## Model
Model: atlas entry = glyph id, face, size, scale, subpixel bin; eviction LRU; structure and the rejected alternative go in the doc comment.

## Acceptance
- Test first: `rendertest` text scenes pass; showcase code-view scroll stress does not grow the atlas without bound.
- The repo's check command passes (`make check`, temp HOME/XDG).
