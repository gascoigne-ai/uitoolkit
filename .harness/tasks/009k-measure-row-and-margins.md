status: todo
kind: bug
commit: fix(view): measure row children and margined children at their width
scope: view/**,widget/**,examples/showcase/**,.harness/architecture/toolkit.md
not_before: 2026-11-26
depends_on:

# Wrapped text in a row or with side margins draws over what follows

Goal: Two cases `009b` left. A child of a `Row` is measured at the row's whole inner width, not the width it ends up with after grow or shrink, and a `Col` child's cross-axis margins are not taken off the width it is measured at. A `Paragraph` with `Grow(1)` in a row, or with `MarginX`, gets too small a height and draws over the next element.

## Model
Model: unchanged

## Acceptance
- Test first: both cases fail for the wrong height before the fix.
- The repo's check command passes (`make check`, temp HOME/XDG).
