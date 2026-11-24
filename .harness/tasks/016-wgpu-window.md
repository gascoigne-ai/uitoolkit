status: todo
kind: feature
commit: feat(wgpu): standalone window loop on X11
scope: renderer/wgpu/window/**,examples/showcase/**
not_before: 2026-11-25
depends_on: 015-wgpu-embed-api

# wgpu renderer stage h: window, input, showcase

Goal: Window and surface without CGO: start with X11 (works under XWayland); state the Wayland scope (needs `libwayland-client`, `libxkbcommon`, xdg-shell, decorations, IME, cursor shapes, fractional scale: its own later task). Events become `core.Input` and `render.Key`; the host delivers key repeat (no ticks). Skipped frames keep the previous image. Wire `-renderer=wgpu` in the showcase behind the build tag; the wgpu-only binary builds with `CGO_ENABLED=0` (the showcase itself links ebiten, so the check applies to a wgpu-only main).

## Model
Model: window state machine is a table of events to transitions.

## Acceptance
- Test first: event translation tests; showcase runs on both renderers and the sections match within tolerance.
- The repo's check command passes (`make check`, temp HOME/XDG).
