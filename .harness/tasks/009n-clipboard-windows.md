status: todo
kind: bug
commit: fix(core): copy and paste on Windows
scope: core/**,docs/**,.harness/architecture/toolkit.md,CHANGELOG.md
not_before: 2026-11-26
depends_on:

# The system clipboard does nothing on Windows

Goal: `core.SystemClipboard` knows only wl-copy, xclip, xsel and pbcopy, so copy and paste fail on Windows. A new dependency needs the user's approval first.

## Model
Model: unchanged

## Acceptance
- Test first: the clipboard picks a working tool or API on Windows, and clipboard text is treated as untrusted. It fails for the reported reason before the fix.
- `docs/ui-guidelines.md` stays true.
- The repo's check command passes (`make check`, temp HOME/XDG).
