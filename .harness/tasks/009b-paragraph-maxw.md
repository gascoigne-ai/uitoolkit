status: todo
kind: bug
commit: fix(widget): measure Paragraph at its MaxW
scope: widget/**,view/**,examples/showcase/**,.harness/architecture/toolkit.md
not_before: 2026-11-26
depends_on:

# Paragraph with MaxW draws over the next element

Goal: Height is measured at the offered width but the text wraps at the narrower `MaxW` width, so the extra lines draw over the element below.

## Model
Model: unchanged

## Acceptance
- Test first: a `Paragraph` with `MaxW` narrower than the offered width is as tall as its wrapped lines. It fails for the reported reason before the fix.
- Any showcase workaround for this bug is removed, and every changed golden PNG is looked at.
- The repo's check command passes (`make check`, temp HOME/XDG).
