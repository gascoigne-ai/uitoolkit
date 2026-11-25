# uitoolkit

A small immediate-mode-feeling GUI toolkit for Go: a view tree with flex layout, a widget set, HTML-like templates (`markup`), JSON themes and desktop helpers. It started inside gascoigne-workshop and is now its own module.

It draws through `render`; the Ebitengine renderer is the separate module `renderer/ebiten` (package `ebitenrender`, on [Ebitengine](https://ebitengine.org)), so the toolkit module itself needs no GPU library. The plan (see `.harness/tasks/`) is to split it into a renderer-agnostic toolkit plus separate renderer modules: Ebitengine and a no-CGO wgpu renderer that can also be embedded in another wgpu program.

```
go get github.com/gascoigne-ai/uitoolkit
go get github.com/gascoigne-ai/uitoolkit/renderer/ebiten   # the Ebitengine renderer
```

Packages: `core` (tokens, colors, geometry, input, drawing calls), `view` (element tree, layout, runtime), `widget` (components), `markup` (templates), `theming`, `engine`, `desktop/*`, `render` (the renderer contract), `renderer/ebiten` (Ebitengine renderer module).

Run the example: `go run ./examples/hello` (`examples` is a module of its own; `go.work` joins the three modules for local work). Tests: `make check` (needs a display or `xvfb-run -a`).

Licensed under Apache 2.0.
