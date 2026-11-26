status: todo
kind: feature
commit: feat(theming): let themes set timing and input feel
scope: core/**,theming/**,view/**,widget/**,engine/**,desktop/**,renderer/ebiten/**,examples/**,docs/**,CHANGELOG.md,.harness/architecture/**
not_before: 2026-11-27
depends_on: 009p-theme-metrics

# Themes set timing and input feel

Goal: Caret blink period, tooltip delay, toast duration, double-click time and distance, wheel step and the similar feel values are theme settings. The decisions in `009p` apply. Exported names that carry such a value follow the "remove them" decision.

## Model
The designer writes it. It must name every value that also comes from the OS or from an app setting today (key repeat is one candidate) and what wins; if the code does not settle the precedence, ask the user before building.

## Acceptance
- Test first.
- With the built-in themes behaviour is unchanged.
- Each value changes behaviour when a theme sets it; zero, negative and huge values are rejected with an error, so no theme can freeze or spin the UI.
- `docs/ui-guidelines.md`, CHANGELOG and the architecture doc are true.
- The repo's check command passes (`make check`, temp HOME/XDG).
