# Visual style

The interface guidelines (sizes, colors, type, shapes, icons, states, each widget and when to use it, keyboard, pointer, layers, scaling, theming, accessibility, and how to build an app) are public, in [`docs/ui-guidelines.md`](../docs/ui-guidelines.md). This file holds only what a contributor who changes the toolkit's look needs besides them.

## Sections moved

Code comments and `architecture/toolkit.md` cite sections of this file by name. They are now in `docs/ui-guidelines.md`:

| Cited as | Now |
|---|---|
| `style.md`, Rules | [Rules](../docs/ui-guidelines.md#rules) |
| `style.md`, Sizes | [Sizes](../docs/ui-guidelines.md#sizes) |
| `style.md`, Typography (or Type) | [Typography](../docs/ui-guidelines.md#typography) |
| `style.md`, Shapes and lines | [Shapes and lines](../docs/ui-guidelines.md#shapes-and-lines) |
| `style.md`, Icons | [Icons](../docs/ui-guidelines.md#icons) |
| `style.md`, States | [States](../docs/ui-guidelines.md#states) |
| `style.md`, Theming pipeline | [Theming pipeline](../docs/ui-guidelines.md#theming-pipeline) |
| the look of dialogs, tooltips, toasts | [Widgets](../docs/ui-guidelines.md#widgets-when-to-use-which) |

## Rules for changing the look

- **The guidelines change with the look.** A change to a token, size, control style or widget look updates `docs/ui-guidelines.md` in the same commit. A behavior the code does not have is not written there; a known bug goes under its Known limitations until it is fixed.
- **Tokens in widget code.** A widget takes colors from `core` tokens and sizes from `ctx.Sizes` through the roles (`HOf(core.RowOf)`, `PadOf(core.GutterOf)`). A size no class or role fits is a new row in the core size table (`core/sizes.go`) with its reason, and a row in the guidelines' Sizes tables.
- **Corners through the theme.** A widget takes its radius from a control style (`view.Control`, `Ctx.Radius`), so the corner setting reaches it. A square fill in a rounded panel is a bug; round it with `RoundT`, `RoundB` or `RoundBL`.
- **Both themes list the same controls and tokens** (`TestBuiltinThemesListTheSameControlsAndTokens`). A new token or control goes into `theming/themes/dark.json` and `light.json` together.
- **Contrast is tested.** `theming/palette_test.go` holds the ratios: `TestThemesAreReadable` for text, `TestNonTextElementsAreVisible` for borders, rules, icons and row states, and `theming/highcontrast_test.go` for high contrast. A color change that breaks them is a change of these rules, agreed with the user, not of a test.
- **Go and markup agree.** A widget with a template element looks the same built either way: `TestGoAndMarkupFormsDumpAlike` and `TestGoAndMarkupLayersDumpAlike` in `examples/showcase/sections` compare dumps and painted ops.
- **New icons** are the toolkit's own SVGs in `widget/icons`; an icon from an open set goes in with its license file.

## Verifying a change to the look

1. Run the showcase (`go run ./examples/showcase`) and look at the changed section in both themes, with high contrast, at a large text size and at scale 2. Layout and widget logic still get unit tests.
2. Golden screenshots (`examples/showcase/testdata/golden/*.png`: every section in each of its poses in the dark theme, theming in light and high contrast, typography on a scale-2 display) and pinned dumps (`examples/showcase/sections/testdata/dump`) run in `make check`. The screenshots need a display: run `xvfb-run -a make check`, as CI does. Ebitengine's tolerance is zero.
3. A failing shot writes `<name>.got.png` and `<name>.diff.png` to `examples/showcase/testdata/failed/` (ignored by git). Look at both.
4. Regenerate with `go test ./examples/showcase/... -update` under `xvfb-run -a`, so the PNGs come from the same software rasterizer as CI. Look at every changed PNG; only the sections the task changes may change.

## Decisions

Still-true choices of the look. Where a reason is code-specific, the doc comment named holds it in full.

| Choice | Why | Rejected | Revisit when |
|---|---|---|---|
| Dialogs are panels inside the window, each with its own title bar (`widget.Dialog`) | Ebitengine has one OS window per process | | Ebitengine offers more than one window |
| The dialog title bar is the `hero` banner | it differs from the body in every theme | | |
| The grab cursor shows only while a bar is pressed | an idle bar should read as a title, not a control | a move cursor on hover | |
| A click outside a dialog leaves it open (`KeepOnOutsideClick`) | a stray click must not throw away the user's work | closing on any outside click, as menus do | |
| Palette, surfaces, outlines and radii follow the gascoigne CLI dashboard (`TestPalettesFollowTheCLIDashboard`) | | | the dashboard changes its look |
| Borders stronger than the dashboard's web borders | a native window's edges must show (`TestNonTextElementsAreVisible`) | the web values as they are | |
| Themes are data per control, in our own small JSON (`theming` package doc) | it carries radii and gradients, and its parser stays in the standard library | | |
| High contrast ratios sit above WCAG's (`highBars`) | both palettes already meet WCAG, so the switch would change nothing visible | WCAG ratios | |
| Xwayland scale from the compositor, no supersampling (`displayscale` package doc) | Ebitengine reports 1 there | | Ebitengine reports the Xwayland scale |
| Icons are our own SVGs; an open-set icon only with its license file | the license travels with the file | | |
| Text drawn unhinted, fractionally placed, with a second pass at `StemBoost` (`Ctx.Text` doc) | macOS-like weight with no new dependency and no new font; coverage only, so measurement stays exact | changing the typeface; a new rasterizer dependency; a coverage shader or a wider outline (more code, or advances drift) | the boost looks too heavy or too light at scale 2 |
| Flat fills and one gradient, `hero` | follows the dashboard | | |
| Square tab outline, open at the bottom | the open tab runs into its pane | | |

**Not done.** Toasts take no control and ignore the corner setting. An icon before a dialog title needs a place in `DialogTitle` first. From the dashboard, still open: blurred shadows, 2 px focus rings, medium font weight and letter-spacing, a selected-row accent bar; each needs a renderer or layout change, or the user's approval.
