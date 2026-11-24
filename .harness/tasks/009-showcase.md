status: todo
kind: feature
commit: feat(showcase): add the widget showcase on the Ebitengine renderer
scope: examples/showcase/**
not_before: 2026-11-25
depends_on: 008-split-ebiten-module

# Showcase app: one UI, every widget, renderer chosen at run time

Goal: Nested module `examples` (own `go.mod`; `replace` only through `go.work`). `main.go` takes `-renderer=ebiten|wgpu`; wgpu is selectable only with build tag `wgpu` and says "not built" until task 016. Layout: left nav of sections, content pane, top bar with theme switcher (dark, light, high-contrast from `theming/themes`) and scale control. Sections, each built both in Go and from `markup`: buttons and chips; text fields and text area; checkbox, switch, slider; lists, tabs, split panes, scroll; menus, tooltips, modal dialog; code view over a long file (scroll and atlas stress); images and the 28 icons; typography (UI, small, mono, bold) and DPI; theming with a live JSON edit. An overlay (toggle key) shows frame time and `core.DrawStats` counts.

## Model
Model: one `Showcase` struct holds the selected section, theme and scale; sections are a table (name, build func), no switch on section name.

## Acceptance
- Test first: `Example_*` builds each section on the fake canvas and checks its `view.Dump` text.
- Golden screenshot per section through the ebiten `Offscreen.ReadPixels`, under xvfb.
- `go run ./examples/showcase -renderer=ebiten` opens the window.
- The repo's check command passes (`make check`, temp HOME/XDG).
