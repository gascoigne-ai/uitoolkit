status: todo
kind: feature
commit: feat(theming): make the default text one pixel bigger
scope: core/**,view/**,widget/**,markup/**,theming/**,renderer/ebiten/**,examples/showcase/**,.harness/architecture/toolkit.md,.harness/style.md,CHANGELOG.md
not_before: 2026-11-26
depends_on:

# Bigger default fonts at 100% scale

Goal (user, 2026-11-26): text at 100% scale is a bit bigger. Decision (user): the toolkit defaults change from 14/12/13 px to 15/13/14 (interface/small/code), for every program using the toolkit.

## Model
Model: unchanged

## Acceptance
- Test first: the resolved zero `FontSizes` are 15, 13, 14, and `core` defaults agree with `theming`.
- No showcase section clips or overlaps at 1024x720; every changed golden is looked at.
- The repo's check command passes (`make check`, temp HOME/XDG).
