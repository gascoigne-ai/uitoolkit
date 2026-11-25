# uitoolkit

A small immediate-mode-feeling GUI toolkit for Go: a view tree with flex layout, a widget set, HTML-like templates (`markup`), JSON themes and desktop helpers. It started inside gascoigne-workshop and is now its own module.

It draws through `render`; the Ebitengine renderer is `ebitenrender` ([Ebitengine](https://ebitengine.org)). The plan (see `.harness/tasks/`) is to split it into a renderer-agnostic toolkit plus separate renderer modules: Ebitengine and a no-CGO wgpu renderer that can also be embedded in another wgpu program.

```
go get github.com/gascoigne-ai/uitoolkit
```

Packages: `core` (tokens, colors, geometry, input, drawing calls), `view` (element tree, layout, runtime), `widget` (components), `markup` (templates), `theming`, `engine`, `desktop/*`.

Run the example: `go run ./examples/hello`. Tests: `make check` (needs a display or `xvfb-run -a`).

Licensed under Apache 2.0.
