status: todo
kind: feature
commit: feat(wgpu): fills, gradients, scissor and rounded SDF
scope: renderer/wgpu/**
not_before: 2026-11-25
depends_on: 011-wgpu-ffi-spike

# wgpu renderer stage b and c: geometry

Goal: Implements `render.Canvas` for solid and gradient rects, `Sub` as a clamped scissor (a scissor outside the target is a validation error; consider a per-vertex clip rect so two clips per widget do not flush the batch), rounded fill and ring outline as SDF in the fragment shader with per-corner radii, AA lines. Non-sRGB target view required (ebiten blends in gamma space; an sRGB view would change antialiasing); pipelines are keyed by format, sample count and depth format. One dynamic vertex buffer per frame. Offscreen is encoded as separate passes first, the main pass last.

## Model
Model: a frame records a CPU command list; `End` encodes offscreen passes, then the main pass.

## Acceptance
- Test first: `rendertest` scenes for these ops pass within the declared tolerance on a software adapter.
- The repo's check command passes (`make check`, temp HOME/XDG).
