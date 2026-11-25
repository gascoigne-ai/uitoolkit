# Changelog

## Unreleased

- Removed: `core.Surface`, `Ctx.CanvasOf`, `Ctx.EbitenImage`, the old `Ctx.Clip(*ebiten.Image, Rect)`, `Sampler.Read`, `systeminfo.Ebiten` and `ebitenrender.ImageOf`. The toolkit packages no longer import Ebitengine.
- Renamed: `Ctx.ClipCanvas` is now `Ctx.Clip`. The `Ctx` draw calls and `view.Draw`, `DrawBelow` and `DrawFrom` take a `render.Canvas`. Use `sampler.Sample(ebitenrender.ReadInput())` for input.
- Changed: `engine.Binding.Key`, `IsModifierKey`, `IsFunctionKey` and `Matches` use `render.Key`; shortcut strings are unchanged. `systeminfo.Graphics` is an alias of the new `render.Graphics`, which names the renderer; `ebitenrender.ReadGraphics` fills it, and the System report shows a Renderer row first in Rendering.
- Changed: `examples/hello` runs through `ebitenrender.Run` and is HiDPI-sharp.
- Silent break: `*ebiten.Image` still satisfies `render.Image` (it has `Bounds`), so `widget.Image(img)` and `<image img>` compile but panic at draw with `render: image from another renderer`. Migrate with `ctx.Device.NewImage(img)`.
- Extracted from gascoigne-workshop as its own module, `github.com/gascoigne-ai/uitoolkit`. History before the move to `uitoolkit/` in the workshop is not carried over.
