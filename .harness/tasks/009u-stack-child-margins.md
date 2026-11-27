status: todo
kind: feature
commit: feat(view): keep the margins of a Stack's non-floating children
scope: view/**,widget/**,markup/**,examples/showcase/**,docs/**,CHANGELOG.md,.harness/architecture/**
not_before: 2026-11-27
depends_on: 009k-measure-row-and-margins

# A Stack honours the margins of every child

Goal: A `Stack` ignores the margins of its non-floating children, which fill the whole stack, while floating children keep theirs. Non-floating children keep their margins too, in measuring and in layout. This changes what such a child looks like today, so the CHANGELOG lists it under changed behaviour.

## Decisions (user, 2026-11-27)
- This is its own task, separate from the measuring fix in `009k`.

## Model
Model: unchanged

## Acceptance
- Test first.
- With no margins, layout is unchanged. A built-in dump or golden that would change is shown to the user before it is updated.
- `docs/ui-guidelines.md`, CHANGELOG and the architecture doc are true.
- The repo's check command passes (`make check`, temp HOME/XDG).
