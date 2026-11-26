status: todo
kind: bug
commit: fix(ebitenrender): make the image filter test deterministic
scope: renderer/ebiten/**,Makefile,.harness/architecture/toolkit.md
not_before: 2026-11-26
depends_on:

# TestDrawImageScalesWithTheFilter fails at random

Goal: On a clean export of HEAD it failed in 5 of 20 runs of `go test ./renderer/ebiten/... ./examples/...` under xvfb without `-race`, with both `nearest` and `linear` subtests seen failing, for example `44 pixels differ, by at most 200; the first at (3,12): got {200 30 0 255}, want {0 40 100 128}` and `1 pixels differ, by at most 1`. The cause is unknown: find it first, whether in the test or in the renderer, and reproduce it deterministically. Loosening the assertion or skipping the test needs the user's approval.

## Model
Model: unchanged

## Acceptance
- The cause is shown by a reproduction that fails every time before the fix.
- The test passes 40 runs in a row of the two-module command above.
- The repo's check command passes (`make check`, temp HOME/XDG).
