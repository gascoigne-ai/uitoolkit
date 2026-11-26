status: todo
kind: bug
commit: fix(widget): fold tab-indented blocks whole
scope: widget/**,examples/showcase/**,.harness/architecture/toolkit.md
not_before: 2026-11-26
depends_on:

# CodeView folding ignores tab indents

Goal: `indentOf` in `widget/codeview.go` counts only leading spaces of the first span and not `Line.Depth`, so folding a tab-indented block hides part of it and leaves the closing brace showing.

## Model
Model: unchanged

## Acceptance
- Test first: folding a tab-indented block hides the same lines as the space-indented one. It fails for the reported reason before the fix.
- Any showcase workaround for this bug is removed, and every changed golden PNG is looked at.
- The repo's check command passes (`make check`, temp HOME/XDG).
