status: todo
kind: feature
commit: feat(widget): restart the caret blink while text is edited
scope: core/**,view/**,widget/**,markup/**,examples/**,docs/**,CHANGELOG.md,.harness/architecture/**
not_before: 2026-11-27
depends_on:

# The toolkit restarts the caret blink itself

Goal: A blinking caret is solid and starts its cycle again on typing, caret movement and focus in the toolkit's text widgets, with no app code. Today an app has to call `RestartCaret`.

## Decisions (user, 2026-11-27)
- The toolkit calls it. Apps that already call it are unaffected.

## Model
The designer writes it.

## Acceptance
- Test first, with a fake clock.
- Typing, moving the caret and focusing each restart the cycle in every widget that draws a caret; nothing else does, and a steady caret stays steady.
- `docs/ui-guidelines.md`, CHANGELOG and the architecture doc are true.
- The repo's check command passes (`make check`, temp HOME/XDG).
