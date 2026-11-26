status: todo
kind: feature
commit: feat(theming): let themes set timing and input feel
scope: core/**,theming/**,view/**,widget/**,engine/**,desktop/**,renderer/ebiten/**,examples/**,docs/**,CHANGELOG.md,.harness/architecture/**
not_before: 2026-11-27
depends_on: 009p-theme-metrics

# Themes set timing and input feel

Goal: Caret blink period, tooltip delay, toast duration, double-click time and distance, wheel step and the similar feel values are theme settings. The decisions below apply. Exported names that carry such a value follow the "remove them" decision.

## Decisions (user, 2026-11-27)
- Scope: layout sizes and timing are theme settings. Safety limits (paste size, undo history, cache sizes, `MinLine`/`MaxLine` and the like) stay fixed: a theme file is untrusted input.
- Baselines: the line height stays measured from the font; the theme adds extra leading in pixels. Theme files cannot set font sizes.
- Theme keys: base knobs (unit, spacing steps, leading) drive the derived sizes; a theme may also override any single named size.
- Go API: exported constants for themed values are removed and replaced by values resolved from the theme. No size may silently ignore the theme.
- Leading applies to the height classes, the bars built from them and the space between wrapped `Paragraph` lines. Single-line `Text` and the checkbox keep the measured line.

## Model
The designer writes it. It must name every value that also comes from the OS or from an app setting today (key repeat is one candidate) and what wins; if the code does not settle the precedence, ask the user before building.

## Acceptance
- Test first.
- With the built-in themes behaviour is unchanged.
- Each value changes behaviour when a theme sets it; zero, negative and huge values are rejected with an error, so no theme can freeze or spin the UI.
- `docs/ui-guidelines.md`, CHANGELOG and the architecture doc are true.
- The repo's check command passes (`make check`, temp HOME/XDG).
