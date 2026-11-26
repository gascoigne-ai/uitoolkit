status: todo
kind: docs
commit: docs(style): make style.md the toolkit's interface guidelines
scope: .harness/style.md,.harness/overview.md,.harness/architecture/toolkit.md,.harness/architecture.md,README.md
not_before: 2026-11-26
depends_on:

# User interface guidelines of the toolkit

Goal (user, 2026-11-26): document the user interface guidelines in the toolkit. Decision (user): rewrite `.harness/style.md`, copied from gascoigne-workshop, into the toolkit's own guidelines (spacing, sizes, color and contrast, type, each widget and when to use it, keyboard and focus, pointer, layers, scaling, accessibility); remove the app-only content (its window layout, settings and `internal/` paths) and link it from the README.

## Acceptance
- Every rule, name and number matches the code as it is now; a behavior the code does not have is not written as one.
- No reference to files or packages that are not in this repository.
- The README links to the guidelines.
- The repo's check command passes (`make check`, temp HOME/XDG).
