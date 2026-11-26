status: todo
kind: feature
commit: feat(theming)!: let themes set the size rules and extra leading
scope: core/**,theming/**,view/**,widget/**,markup/**,engine/**,examples/**,docs/**,README.md,CHANGELOG.md,CLAUDE.md,.harness/architecture/**,.harness/style.md
not_before: 2026-11-27
depends_on:

# Themes set the shared size rules and extra leading

Goal: A theme file controls the shared size rules of `core`: the unit, the spacing steps, extra leading, height classes, bars, fixed-shape controls and dialog sizes. First of three tasks; `009q` does the private widget pixels, `009r` timing and input feel.

## Decisions (user, 2026-11-27)
- Scope of the three tasks: layout sizes and timing. Safety limits (paste size, undo history, cache sizes, `MinLine`/`MaxLine` and the like) stay fixed: a theme file is untrusted input.
- Baselines: the line height stays measured from the font; the theme adds extra leading to it (the designer picks add or scale and says why). Theme files still cannot set font sizes.
- Theme keys: a few base knobs (unit, spacing steps, leading) drive the derived sizes; a theme may also override any single named size.
- Go API: the exported size constants are removed and replaced by options resolved at layout, like `HOf`/`WOf`. No size may silently ignore the theme. Nothing is tagged, so the break is accepted.

## Model
The designer writes it, with the `data-model` skill: the theme file block is persisted data. It must say how base knobs, derived values and overrides combine, the key names, what `extends` inherits, and the bounds of every key.

## Acceptance
- Test first.
- With the built-in themes every dump and golden is byte-identical (no `-update`).
- A theme that changes the unit, a step, the leading or one named size changes layout accordingly, in Go and in markup (`Sizes.Scope` names come from the themed values).
- Values that other constants derive from today (`dragEdge`, `edgeSlop`, `formLabelGap`, the table cell field insets) follow the theme.
- New options keep the toolkit API rules (full side and axis families).
- Every new key is bounds-checked with an error, and the theme fuzz target covers it.
- `docs/ui-guidelines.md` (its sample compiles), CHANGELOG and `.harness/architecture/toolkit.md` are true.
- The repo's check command passes (`make check`, temp HOME/XDG).
