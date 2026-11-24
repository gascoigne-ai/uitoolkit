# Rules

Adds to the shared `common/rules.md`, `common/commit-style.md` and `common/code-style.md`.

- Approved deps: `github.com/hajimehoshi/ebiten/v2` (GPU window, input, 2D drawing, `text/v2`) and its transitive requirements, `golang.org/x/image` (embedded Go fonts, SVG raster, `font/sfnt`). Anything else, including `github.com/ebitengine/purego` as a direct dependency, `github.com/go-text/typesetting` as a direct dependency, and a path stroker for the wgpu renderer, needs the user's approval first (tasks 007 and 011 ask).
- cgo is allowed only in the Ebitengine renderer's dependencies (Ebitengine needs it for windowing on Linux). The toolkit module itself stays free of cgo, and the wgpu renderer is built with `CGO_ENABLED=0`.
- Layers are in `CLAUDE.md` and `architecture/toolkit.md`. Drawing and layout stay separate: layout is pure (rects in, rects out) and unit-tested without a GPU. Draw code is thin.
- Persisted key names: shortcut strings users store in config come from `ebiten.Key.String()`. Any key type change keeps them readable (task 002 freezes the table).

## Design
- Every design or plan (a task design note, a layout change, a new package) is made by an Opus agent (`model: opus`), read-only, before any code is written. Do not design on a smaller model.

## Tasks
- Finish a task completely before starting another. An item that cannot be done is written into the task file with the reason and what would unblock it; the user decides whether the task closes.

## Ambiguity
- When a request, task item or design can be read in more than one way, or a choice changes behavior the user would notice, ask before building: put the options, with a recommendation first, to the user and wait. A choice with a conventional default, or a fact in the code, is decided and mentioned in the report.
- Record the answer in the task file; when it matters after the task, also in the doc comment of the code it shapes.

## Untrusted input
- Anything from a file, clipboard or import is untrusted: parsers return an error, never panic, hang or grow without bound. A new parser gets a fuzz target (`FUZZ_TARGETS` in the Makefile).
