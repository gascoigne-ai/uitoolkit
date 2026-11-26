status: todo
kind: bug
commit: fix(theming): reject trailing data and keep square corners
scope: theming/**,.harness/architecture/toolkit.md
not_before: 2026-11-26
depends_on:

# ParseTheme accepts bad input and loses square corners

Goal: `ParseTheme` returns no error for `{"colors":{}}x`, a second JSON value or `null`, because only the first value is decoded. Over a base whose `RadiusScale` is 0 it returns 1. A theme file is untrusted input; also check what an absurd `iconStroke` such as `1e30` does and reject it if drawing cannot take it.

## Model
Model: unchanged

## Acceptance
- Test first: a table of trailing-data and `null` inputs expects errors, and a square-cornered base stays square. It fails for the reported reason before the fix.
- Any showcase workaround for this bug is removed, and every changed golden PNG is looked at.
- The repo's check command passes (`make check`, temp HOME/XDG).
