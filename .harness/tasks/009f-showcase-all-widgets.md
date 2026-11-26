status: todo
kind: feature
commit: feat(showcase): show every widget
scope: examples/showcase/**,.harness/architecture/toolkit.md,.harness/overview.md
not_before: 2026-11-26
depends_on: 009e-examples-only-showcase

# The showcase shows every widget

Goal (user, 2026-11-26): the showcase has an example of every widget. Missing today: `Dialog` with `DialogTitle`, `DialogSubtitle` and `DialogToolbar`, `ResizableDialog`, `Resizable`, `KeyChips`, `Toasts`, `view.Custom`, `view.Lines`, `view.VRuleInset`, and markup forms of `stack` and `hscroll`. Confirm that the `*Icon` constructors are covered by the icon list and `Menu` by `OpenMenu`. Both forms (Go and markup) show each one; only what has no markup element goes through `<slot>`.

## Model
Model: new state is named fields of the showcase's widget state; sections stay table rows.

## Acceptance
- Test first: the section dumps show each added widget.
- Every exported element constructor of `widget` and `view` is used by the showcase or listed in the report with how it is covered.
- Every new or changed golden is looked at.
- The repo's check command passes (`make check`, temp HOME/XDG).
