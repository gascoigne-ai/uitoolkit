status: todo
kind: feature
commit: refactor: drop the ebiten bridge from the toolkit packages
scope: core/*,view/*,widget/*,markup/*,engine/*,desktop/systeminfo/*,ebitenrender/*,render/*,examples/*,CLAUDE.md,.harness/architecture/toolkit.md,README.md,CHANGELOG.md
not_before: 2026-11-25
depends_on: 005-rewire-widget

# Remove the bridge; the toolkit packages stop importing ebiten

Goal: Delete `core/ebitenbridge.go` (user approved with the bridge plan): `core.Surface` becomes `render.Canvas` in every draw signature, the old `Clip` and `Sampler.Read` go, `ClipCanvas` is renamed `Clip` (view/runtime.go, widget/codeview.go twice, field.go, textarea.go). `view.Draw`, `DrawBelow`, `DrawFrom` take `render.Canvas`. ebitenrender drops `ImageOf` and the raw `*ebiten.Image` case in `pictureOf`. markup uses `scoped[render.Image]("img")`. engine `Binding.Key`, `IsModifierKey`, `IsFunctionKey`, `keyByName` and `Matches` use `render.Key`. The graphics info moves to `render.Graphics` with a renderer name and version field, `desktop/systeminfo` may import `render` (user's answer) and keeps `systeminfo.Graphics` as an alias; `ebitenrender.ReadGraphics()` fills it and the hardcoded "Ebitengine" row comes from it. `examples/hello` runs through `ebitenrender.Run`. Update `CLAUDE.md` rule 11 and the layers text.

## Model
Model: unchanged except signatures; the `Binding` string form is unchanged.

## Acceptance
- Test first: every shortcut string in the existing tests parses to the same binding and prints the same string.
- `go list -deps` of every package except ebitenrender and examples shows no `hajimehoshi` (tests included, except ebitenrender's own and `render/text`'s conformance test).
- The repo's check command passes (`make check`, temp HOME/XDG).
