status: todo
kind: feature
commit: feat(core): steady caret by default, with a style option
scope: core/**,view/**,widget/**,engine/**,markup/**,theming/**,examples/showcase/**,.harness/architecture/toolkit.md,.harness/style.md,CHANGELOG.md
not_before: 2026-11-26
depends_on:

# Caret: no blinking by default, selectable style

Goal (user, 2026-11-26): the text caret does not blink by default, and there is an option to control its style. Decisions (user): it is the text caret (fields, text area, code editor), not the mouse pointer; styles are block, bar and underline; block (today's look) stays the default; blinking stays available as an option. The showcase offers both choices.

## Model
Model: one caret setting, owned by `core.Ctx` unless the design finds a better owner: a style (enum: block, bar, underline; zero value block) and blinking (zero value off). Illegal styles are unrepresentable or fall back at the boundary.

## Acceptance
- Test first: a zero setting draws a steady block; each style draws its own shape; blinking, when on, keeps today's half period.
- Idle frame pacing does not keep redrawing for a caret that does not blink.
- The repo's check command passes (`make check`, temp HOME/XDG).
