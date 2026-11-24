status: todo
kind: feature
commit: refactor(core): draw through render.Canvas behind an ebiten bridge
scope: core/*,render/*,ebitenrender/*,view/leaf.go,view/el.go,view/options_look.go,view/fontsource_test.go,view/helpers_test.go,view/custom_test.go,widget/caret.go,widget/paragraph.go,widget/menu.go,widget/codeview.go,widget/*_test.go,markup/attributes.go,markup/*_test.go,engine/binding.go,engine/binding_test.go,theming/theming_test.go,examples/hello/*,CLAUDE.md,.harness/architecture/toolkit.md
not_before: 2026-11-25
depends_on: 002-ebiten-adapter

# Move core.Ctx onto the render contract, with a temporary ebiten bridge

Goal: Old and new APIs exist side by side; task 006 removes the old one. The wider scope is the user's answer: fonts, input keys and `NewCtx` change type and cannot be bridged inside core.

- `NewCtx(th Theme, device render.Device)` is the only constructor. `render.Device` is a new interface: `NewFontSource([]byte) (FontSource, error)`, `NewOffscreen(w, h) Offscreen`, `NewImage(image.Image) Image`; `ebitenrender.Device` returns these. `Ctx.Device` is exported.
- `Fonts` fields are `render.Face`. Width, Advance, LineH, Ascent, Ellipsize and CursorAt use `face.Measure` and `face.Metrics`.
- `core.Surface` (`interface{ Bounds() image.Rectangle }`) is the type of `dst` in every `Ctx` draw method (Fill, Outline, HLine, VLine, Line, Text, TextMid, TextCenter, FillStyled, FillStyledCorners, OutlineStyled), so old and new callers compile. `core/ebitenbridge.go` turns an `*ebiten.Image` into a canvas (`Device.NewCanvas`, memoized on the last dst). It is the only core file importing ebiten or ebitenrender, and holds `CanvasOf` (panics with a clear message when `Device` is not `*ebitenrender.Device`), `EbitenImage`, the old `Clip`, and the old `Sampler.Read`.
- New `(*Ctx).ClipCanvas(render.Canvas, Rect) render.Canvas` beside the old `Clip`.
- Text is two `DrawText` passes; the stem-boost pass uses the color scaled by 0.5.
- Corner and ring masks, `gradientSrc`, `maskMaker`/`ringMaker` leave core. `Stats.MaskFills`/`RingFills` count `FillRoundRect`/`StrokeRoundRect` calls.
- `core.Corners` and `core.Cursor` become permanent aliases of the render types, constants included.
- `Input.Keys` is `[]render.Key`; `Pressed`, `Consume`, `editKeys` use it. Deferred readInput from 002: `ebitenrender.ReadInput() render.InputState` (in `ebitenrender/input.go`, thin, drops `KeyUnknown`), a pure `core.inputFrom(state, goos, repeat)`, and `Sampler.Sample(render.InputState) Input`.
- ebitenrender bridge: `ImageOf(render.Canvas) (*ebiten.Image, bool)`; `pictureOf` also accepts a raw `*ebiten.Image`.
- Forced call sites: view/leaf.go face and chosenFace return `render.Face`; `view.FontSource(render.FontSource)`, where a nil pointer in the interface counts as nil; widget/caret.go and paragraph.go face types; widget/menu.go and codeview.go `render.Key*`; engine `BindingFrom` converts with `ebitenrender.KeyToEbiten` (Binding stays ebiten until 006); markup/attributes.go `handler` names its type with `reflect.TypeFor[F]()`; every `NewCtx` call (45 in 25 files).
- Delete (user approved): `core/cornermask.go`; the mask cache tests in `core/cornermask_test.go` (covered by `ebitenrender/mask_test.go`; the RingFills/MaskFills count tests are ported to `rendertest.Canvas`, not dropped); `ebitenrender/coreparity_test.go` (an import cycle once core imports ebitenrender).
- Tests that measure or draw text use `ebitenrender.NewDevice()` until 007, so expected numbers stay.

## Model
`Fonts{UI, Small, Mono, Bold, SmallBold render.Face}`. `Ctx` holds `Device` and the loaded sources; SetScale and SetTheme only resize faces, SetFonts reloads through the device. `widthKey{face render.Face; content string}`, `faceKey = render.Face`: both comparable. `render.InputState{X, Y int; WheelX, WheelY float64; Typed []rune; Buttons, ButtonsPressed MouseButtons; Down []KeyDown; TPS int}`, `KeyDown{Key Key; Ticks int}`; Ticks is 1 on the tick a key goes down, so "just pressed" is derived; `Down` keeps the aggregate keys. Invariant: core imports ebiten and ebitenrender only in `core/ebitenbridge.go`.

## Acceptance
- Test first: table test of `inputFrom` (Ctrl+A is SelectAll, Cmd on darwin, word and line edits, repeat timing, keys present in `Keys`); bridge tests (an `*ebiten.Image` and a `rendertest.Canvas` given to `Fill` give the same ops; `CanvasOf` panics with another device); the nil-FontSource view test.
- Existing tests change only where a type forces it; expected numbers unchanged.
- In core only `ebitenbridge.go` imports `hajimehoshi` or `ebitenrender` (direct imports; the transitive check is in 006).
- Fills, strokes and lines match today; text edge pixels may differ by up to 1/255 (user accepted). Report any known difference.
- The repo's check command passes (`make check`, temp HOME/XDG).
