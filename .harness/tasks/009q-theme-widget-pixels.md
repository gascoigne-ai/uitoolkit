status: todo
kind: feature
commit: feat(theming): let themes set the widgets' own sizes
scope: core/**,theming/**,view/**,widget/**,markup/**,examples/**,docs/**,CHANGELOG.md,.harness/architecture/**
not_before: 2026-11-27
depends_on: 009p-theme-metrics

# Themes set the widgets' own sizes

Goal: The private pixel constants of widgets and views (button and field padding, menu, toast, tooltip, scrollbar, slider, fold column, table grips and the rest of the visual geometry) are theme sizes too, in the shape `009p` built. The decisions in `009p` apply. The designer lists every constant and says for each whether it moves, with the reason when it stays.

## Model
The designer writes it, extending the theme block of `009p`.

## Acceptance
- Test first.
- With the built-in themes every dump and golden is byte-identical.
- Each moved value changes what is drawn and hit-tested when a theme sets it.
- Every new key is bounds-checked with an error.
- `docs/ui-guidelines.md`, CHANGELOG and the architecture doc are true.
- The repo's check command passes (`make check`, temp HOME/XDG).
