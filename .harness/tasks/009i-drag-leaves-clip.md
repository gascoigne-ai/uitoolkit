status: todo
kind: bug
commit: fix(view): keep a drag going outside a clipping parent
scope: view/**,widget/**,examples/showcase/**,.harness/architecture/toolkit.md
not_before: 2026-11-26
depends_on:

# A drag breaks when the pointer leaves a clipping parent

Goal: In `view/runtime.go` `update`, a clipping element hands its children the pointer `-1<<20, -1<<20` when the pointer is outside it, even while a child is mid-drag. Dragging the right edge of a `Resizable` and then moving into the top bar drops its width to the minimum; a `Split` divider collapses its pane the same way.

## Model
Model: unchanged

## Acceptance
- Test first: a drag that started inside a clipping parent follows the real pointer after it leaves; a pointer outside still does not hover or click clipped children. It fails for the reported reason before the fix.
- The repo's check command passes (`make check`, temp HOME/XDG).
