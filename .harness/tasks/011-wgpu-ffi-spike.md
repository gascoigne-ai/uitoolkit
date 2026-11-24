status: todo
kind: feature
commit: feat(wgpu): spike the no-CGO wgpu-native bindings
scope: renderer/wgpu/internal/webgpu/**
not_before: 2026-11-25
depends_on: 010-rendertest-suite

# Prove purego can drive wgpu-native with CGO_ENABLED=0

Goal: **Asks the user first**: `github.com/ebitengine/purego` becomes a direct dependency. New module `renderer/wgpu`. Spike (throwaway notes in the task file, then a small package): `Dlopen` of `libwgpu_native`; instance, adapter, device through the async callbacks; by-value structs (`CallbackInfo`, `WGPUStringView`) and which purego forms work on linux/amd64; float args via `RegisterFunc`; callbacks registered once and dispatched by userdata (slots are limited and never freed); callbacks on foreign threads; pinning of nested descriptors (`runtime.Pinner`) or C memory; surface status Outdated/Lost. Generate struct offsets from the pinned `webgpu.h` with a C compiler in CI and compare in a test. Pin the wgpu-native version.

## Model
Model: handle types are `uintptr` newtypes; invariant: no Go pointer is retained by C after a call returns.

## Acceptance
- Test first: offset test against the generated table; a test that clears a texture on a software adapter and reads it back (skipped when the library is absent).
- Builds with `CGO_ENABLED=0`.
- The repo's check command passes (`make check`, temp HOME/XDG).
