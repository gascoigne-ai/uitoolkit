status: todo
kind: chore
commit: build(ebitenrender): move the Ebitengine renderer to its own module
scope: renderer/ebiten/**,go.work
not_before: 2026-11-25
depends_on: 006-rewire-markup-engine

# renderer/ebiten becomes its own module

Goal: Move `ebitenrender` to `renderer/ebiten` with its own `go.mod` (`github.com/gascoigne-ai/uitoolkit/renderer/ebiten`); the root module drops ebiten. Add `go.work` for local development only. Tag rules: root `vX.Y.Z`, renderer `renderer/ebiten/vX.Y.Z`; release order root first, then bump the renderer `require`, then tag it. CI: add a job that fails when `go list -deps ./...` of the root module shows `hajimehoshi` or `wgpu`. Update `.harness/architecture/toolkit.md`, `CLAUDE.md` layers, README, CHANGELOG.

## Acceptance
- `go list -deps ./...` in the root module has no ebiten.
- Tests in `renderer/ebiten` pass under xvfb.
- The repo's check command passes (`make check`, temp HOME/XDG).
