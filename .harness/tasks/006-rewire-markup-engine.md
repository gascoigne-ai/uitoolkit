status: todo
kind: feature
commit: refactor(markup): use render types in markup and engine
scope: markup/**,engine/**,desktop/systeminfo/**
not_before: 2026-11-25
depends_on: 005-rewire-widget

# Finish removing ebiten from the toolkit packages

Goal: `markup` scoped `img` takes `render.Image` (`markup/elements.go`); `engine.Binding` uses `render.Key` (`keyByName` walks the frozen name table); `systeminfo` takes the graphics row from the renderer instead of hardcoding "Ebitengine". `examples/hello` runs through `ebitenrender.Run`. Update `CLAUDE.md` rule 11 and the layers text.

## Model
Model: unchanged except signatures; `Binding` string form unchanged.

## Acceptance
- Test first: shortcut strings that parsed before parse to the same bindings.
- `go list -deps` of every package except `ebitenrender` and `examples` shows no ebiten.
- The repo's check command passes (`make check`, temp HOME/XDG).
