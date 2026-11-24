status: todo
kind: feature
commit: refactor(widget): draw through render.Canvas
scope: widget/**
not_before: 2026-11-25
depends_on: 004-rewire-view

# Move widget onto the render contract

Goal: All 21 drawing files. Route the bypasses through `Canvas`: `icon.go` (`DrawImage`, `vector` paths, `ebiten.NewImage`; the icon mask cache moves behind the device), `svg.go` (caps and joins become `render` types), `image.go` (`FilterLinear` becomes a filter option), `widget.Image` takes `render.Image`. Keys in `menu.go`, `codeview.go` etc. become `render.Key`.

## Model
Model: unchanged except signatures. Invariant: no package-level GPU caches in `widget`.

## Acceptance
- Test first: widget tests (incl. `clipskip_test.go`) run on the fake canvas.
- `go list -deps ./widget | grep -c hajimehoshi` is 0.
- The repo's check command passes (`make check`, temp HOME/XDG).
