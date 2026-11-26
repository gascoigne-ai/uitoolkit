status: todo
kind: feature
commit: feat(core): blink the text caret by default again
scope: core/**,widget/**,engine/**,examples/showcase/**,.harness/architecture/toolkit.md,.harness/style.md,CHANGELOG.md
not_before: 2026-11-26
depends_on:

# The caret blinks by default

Goal (user, 2026-11-26): the toolkit blinks the text caret by default; an app can still choose a steady one. This reverses the default of task 009g and keeps its shapes. Decision (user): the showcase follows the toolkit, so its Blink checkbox starts checked. The changelog no longer announces a change of blink behavior, since the unreleased default is the released one again.

## Model
Model: `core.Caret`'s zero value is a blinking block.

## Acceptance
- Test first: a zero `Caret` blinks at `CaretHalfPeriod`; the steady choice stays shown and counts no reads.
- The showcase starts with Blink checked, and unchecking it gives a steady caret.
- The repo's check command passes (`make check`, temp HOME/XDG).
