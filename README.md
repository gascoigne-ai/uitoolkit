# uitoolkit

A small immediate-mode-feeling GUI toolkit for Go: a view tree with flex layout, a widget set, HTML-like templates (`markup`), JSON themes and desktop helpers. It started inside gascoigne-workshop and is now its own module.

It draws through `render`; the Ebitengine renderer is the separate module `renderer/ebiten` (package `ebitenrender`, on [Ebitengine](https://ebitengine.org)), so the toolkit module itself needs no GPU library. The plan (see `.harness/tasks/`) is to split it into a renderer-agnostic toolkit plus separate renderer modules: Ebitengine and a no-CGO wgpu renderer that can also be embedded in another wgpu program.

No version is tagged yet, so depend on a local clone through `replace` directives; [Depending on it today](docs/ui-guidelines.md#depending-on-it-today) has the exact `go.mod` lines.

Packages: `core` (tokens, colors, geometry, input, drawing calls), `view` (element tree, layout, runtime), `widget` (components), `markup` (templates), `theming`, `engine`, `desktop/*`, `render` (the renderer contract), `renderer/ebiten` (Ebitengine renderer module).

**Building an interface with it:** read the [interface guidelines](docs/ui-guidelines.md): a minimal complete app, sizes, colors, type, every widget and when to use it, keyboard, pointer, layers, scaling, theming and accessibility.

Run the showcase, every widget in one window: `go run ./examples/showcase` (`examples` is a module of its own; `go.work` joins the three modules for local work). Tests: `make check` (needs a display or `xvfb-run -a`).

Licensed under Apache 2.0.
