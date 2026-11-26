status: todo
kind: chore
commit: refactor(examples): the showcase is the only example
scope: examples/doc.go,examples/example_test.go,examples/hello/**,README.md,CHANGELOG.md,doc.go,.harness/architecture.md,.harness/overview.md,.harness/architecture/toolkit.md
not_before: 2026-11-26
depends_on:

# Drop the old examples

Goal (user, 2026-11-26): drop the current examples; the showcase replaces them. Delete `examples/doc.go`, `examples/example_test.go` and `examples/hello/`, and make every doc that names them point at `go run ./examples/showcase`. The changelog line about `examples/hello` becomes a removal note. The showcase stays at `examples/showcase`. Do not run `go mod tidy` in `examples`: it fails until the first tag.

## Acceptance
- No file outside `testdata` names `examples/hello` or the removed examples.
- The repo's check command passes (`make check`, temp HOME/XDG).
