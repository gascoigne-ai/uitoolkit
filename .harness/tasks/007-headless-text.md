status: todo
kind: feature
commit: feat(render): add a pure-Go shaping and metrics face
scope: render/text/**
not_before: 2026-11-25
depends_on: 003-rewire-core

# A Face that measures and shapes without a GPU

Goal: Package `render/text` over `go-text/typesetting` gives `Measure` and `Metrics` with no renderer. Toolkit tests and the wgpu renderer use it so layout (wrap, ellipsis, caret) is the same across renderers. **Asks the user first**: `go-text/typesetting` becomes a direct dependency (`.harness/rules.md`).

## Model
Model: new package; a `Face` value is comparable and cheap to copy.

## Acceptance
- Test first: conformance test over a text corpus: this face and the ebiten `text.Measure`/`Metrics` agree within 0.01 px per run.
- The repo's check command passes (`make check`, temp HOME/XDG).
