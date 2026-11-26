status: todo
kind: bug
commit: fix(widget): keep table fills inside its rounded outline
scope: widget/**,examples/showcase/**,docs/**,.harness/architecture/toolkit.md
not_before: 2026-11-26
depends_on:

# Table header and last row fill square corners over the rounded outline

Goal: The header (`Bg(Panel)`) and a selected, hovered or zebra last row are square fills inside the table's radius 6 outline.

## Model
Model: unchanged

## Acceptance
- Test first: the header and last row fills follow the table's corner radius. It fails for the reported reason before the fix.
- `docs/ui-guidelines.md` stays true.
- The repo's check command passes (`make check`, temp HOME/XDG).
