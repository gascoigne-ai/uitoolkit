status: todo
kind: feature
commit: feat(wgpu): embedding API for another wgpu program
scope: renderer/wgpu/**,examples/wgpu-embed/**
not_before: 2026-11-25
depends_on: 014-wgpu-text

# New/Resize/Begin/End on a caller-owned device

Goal: `New(device, queue, format, opts)`, `Resize(w, h, scale)`, `Begin(encoder, view)`, `End()`. The renderer never creates a window or presents. Handles are only valid inside the same loaded wgpu-native library and version: the embedder passes the library path or handle and a version check is made; Dawn is unsupported (written in the doc). `Host.Draw` returning "did not draw" tells the host to keep the previous frame. `examples/wgpu-embed`: a spinning triangle with the toolkit UI drawn over it in the host render pass.

## Model
Model: renderer holds no surface; frame state is valid only between `Begin` and `End`.

## Acceptance
- Test first: offscreen embed test renders UI over a cleared host texture and matches the golden.
- The repo's check command passes (`make check`, temp HOME/XDG).
