status: todo
kind: bug
commit: fix(widget): draw menu separator lines
scope: widget/**,examples/showcase/**,.harness/architecture/toolkit.md
not_before: 2026-11-26
depends_on:

# Menu separators are never drawn

Goal: The `HRule` inside the separator row of `widget/menu.go` is laid out 0 wide (`plain h=1 bg=border rect=10,32,0,1`), so a menu shows no separator lines. `overlays-menu.png` records the bug and is regenerated.

## Model
Model: unchanged

## Acceptance
- Test first: a menu with a separator lays its rule out across the row. It fails for the reported reason before the fix.
- Any showcase workaround for this bug is removed, and every changed golden PNG is looked at.
- The repo's check command passes (`make check`, temp HOME/XDG).
