status: todo
kind: bug
commit: fix(view): do not hover clipped content from outside the clip
scope: view/**,widget/**,examples/showcase/**,docs/**,CHANGELOG.md,.harness/architecture/toolkit.md
not_before: 2026-11-27
depends_on:

# Hover reaches clipped content from outside the clip

Goal: `Draw` does not hide the pointer the way `update` does. `drawSelf` calls `ctx.Hovered` with the element's full rect (`view/runtime.go`, `core/ctx.go`), so a child partly scrolled out of a clip is hovered while the pointer is over its hidden part outside the clip: its tooltip is picked up (reproduced), and from the code `HoverBg` or `Lighten` would light its visible part. Widgets that read hover in their own draw code are checked too. `docs/ui-guidelines.md` says clipped children are not hovered; that must be true.

Rating: frequency 4, severity 2.

## Model
Model: unchanged

## Acceptance
- Test first: with the pointer outside a clip over the hidden part of a child, nothing of that child is hovered (background, tooltip, cursor); over its visible part it still is. It fails for the reported reason before the fix.
- A drag still follows the pointer outside a clip.
- The repo's check command passes (`make check`, temp HOME/XDG).
