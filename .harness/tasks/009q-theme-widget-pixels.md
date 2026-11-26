status: todo
kind: feature
commit: feat(theming): let themes set the widgets' own sizes
scope: core/**,theming/**,view/**,widget/**,markup/**,examples/**,docs/**,CHANGELOG.md,.harness/architecture/**
not_before: 2026-11-27
depends_on: 009p-theme-metrics

# Themes set the widgets' own sizes

Goal: The private pixel constants of widgets and views (button and field padding, menu, toast, tooltip, scrollbar, slider, fold column, table grips and the rest of the visual geometry) are theme sizes too, in the shape of the `Theme.Metrics` size table in `core`. The decisions below apply. The designer lists every constant and says for each whether it moves, with the reason when it stays.

## Decisions (user, 2026-11-27)
- Scope: layout sizes and timing are theme settings. Safety limits (paste size, undo history, cache sizes, `MinLine`/`MaxLine` and the like) stay fixed: a theme file is untrusted input.
- Baselines: the line height stays measured from the font; the theme adds extra leading in pixels. Theme files cannot set font sizes.
- Theme keys: base knobs (unit, spacing steps, leading) drive the derived sizes; a theme may also override any single named size.
- Go API: exported constants for themed values are removed and replaced by values resolved from the theme. No size may silently ignore the theme.
- Leading applies to the height classes, the bars built from them and the space between wrapped `Paragraph` lines. Single-line `Text` and the checkbox keep the measured line.
- Carried from `009p`: code and text-area lines were left as they were until this task. Name their +4 as a size and ask the user whether leading applies to them. Name the switch inset (`Steps.Hair` today) and the tooltip, menu and mono-grid pixels.

## Model
The designer writes it, extending the theme's `metrics` block.

## Acceptance
- Test first.
- With the built-in themes every dump and golden is byte-identical.
- Each moved value changes what is drawn and hit-tested when a theme sets it.
- Every new key is bounds-checked with an error.
- `docs/ui-guidelines.md`, CHANGELOG and the architecture doc are true.
- The repo's check command passes (`make check`, temp HOME/XDG).
