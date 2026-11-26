# Changelog

## Unreleased

- Changed (behavior): `theming.ParseTheme` rejects a theme file that is not one JSON object (`null`, data after the object) or whose `iconStroke` is outside 0 to 4, and keeps a base's `RadiusScale`, so a theme laid over square corners stays square.
- Breaking (behavior): the default text sizes are 15/13/14 px (interface/small/code), up from 14/12/13 (`theming.DefaultFontSize`, `DefaultSmallSize`, `DefaultCodeSize`). `core.BaseLine` is 17 and `core.BaseSmallLine` 15, so `RowH` (29), `CompactH`/`widget.MenuRowH` (25), `SlimH`, `LabelH`, `NoteH`, the bars, `IconBtnW`, `WordBtnW` and `CheckboxSize` (15) grow by 1 px, as does `DefaultSizes()`. Migrate: pass `theming.FontSizes{UI: 14, Small: 12, Code: 13}` to keep the old size.
- Breaking (behavior): the text caret no longer blinks by default. Migrate: `ctx.Caret.Blink = true`. `Ctx.CaretReads` no longer moves while the caret is steady.
- Added: `core.Caret` and `core.CaretShape` (`CaretBlock`, `CaretBar`, `CaretUnderline`), set through `Ctx.Caret`, choose the shape of the text caret and whether it blinks; `core.CaretW` is the thickness of the bar and the underline.
- Breaking: the Ebitengine renderer moved to its own module `github.com/gascoigne-ai/uitoolkit/renderer/ebiten` (package `ebitenrender`); the root module no longer requires Ebitengine. Migrate: change the import path to `github.com/gascoigne-ai/uitoolkit/renderer/ebiten` and `go get` the renderer module. Tags are `vX.Y.Z` for the toolkit and `renderer/ebiten/vX.Y.Z` for the renderer.
- Removed: `core.Surface`, `Ctx.CanvasOf`, `Ctx.EbitenImage`, the old `Ctx.Clip(*ebiten.Image, Rect)`, `Sampler.Read`, `systeminfo.Ebiten` and `ebitenrender.ImageOf`. The toolkit packages no longer import Ebitengine.
- Renamed: `Ctx.ClipCanvas` is now `Ctx.Clip`. The `Ctx` draw calls and `view.Draw`, `DrawBelow` and `DrawFrom` take a `render.Canvas`. Use `sampler.Sample(ebitenrender.ReadInput())` for input.
- Changed: `engine.Binding.Key`, `IsModifierKey`, `IsFunctionKey` and `Matches` use `render.Key`; shortcut strings are unchanged. `systeminfo.Graphics` is an alias of the new `render.Graphics`, which names the renderer; `ebitenrender.ReadGraphics` fills it, and the System report shows a Renderer row first in Rendering.
- Removed: `examples/hello` and the tested examples of package `examples`; `go run ./examples/showcase` replaces them. `examples` is now a module of its own.
- Silent break: `*ebiten.Image` still satisfies `render.Image` (it has `Bounds`), so `widget.Image(img)` and `<image img>` compile but panic at draw with `render: image from another renderer`. Migrate with `ctx.Device.NewImage(img)`.
- Extracted from gascoigne-workshop as its own module, `github.com/gascoigne-ai/uitoolkit`. History before the move to `uitoolkit/` in the workshop is not carried over.
