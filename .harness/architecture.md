# Architecture

Start with `overview.md`.

| File | Module | Covers |
|---|---|---|
| `architecture/toolkit.md` | whole repo | core, view, widgets, markup API and drawing |
| `architecture/toolkit.md` (renderer/ebiten) | `github.com/gascoigne-ai/uitoolkit/renderer/ebiten` | the Ebitengine renderer, package `ebitenrender`; own `go.mod`, tagged `renderer/ebiten/vX.Y.Z` |
| (none) | `github.com/gascoigne-ai/uitoolkit/examples` | runnable examples and `hello`; own `go.mod`, never tagged |
