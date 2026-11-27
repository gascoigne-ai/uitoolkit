# Interface guidelines

How to build an app with uitoolkit and how its interface is meant to look and behave. For people and agents who write an app on the toolkit, and for those who change its widgets.

Every number here is a constant or theme value in the code, named so you can find it. Sizes are in logical pixels unless a section says otherwise. Where the toolkit leaves something to the app, this page says so.

Contents: [Depending on it today](#depending-on-it-today) · [A minimal app](#a-minimal-app) · [What the app owns](#what-the-app-owns) · [Rules](#rules) · [Sizes](#sizes) · [Color roles](#color-roles) · [Typography](#typography) · [Shapes and lines](#shapes-and-lines) · [Icons](#icons) · [States](#states) · [Widgets](#widgets-when-to-use-which) · [Keyboard and focus](#keyboard-and-focus) · [Pointer](#pointer) · [Layers](#layers) · [Scaling](#scaling) · [Theming pipeline](#theming-pipeline) · [Accessibility](#accessibility) · [Composing a screen](#composing-a-screen) · [Known limitations](#known-limitations)

## Depending on it today

The repository holds three Go modules (Go 1.27.4 or newer):

| Module | What it is |
|---|---|
| `github.com/gascoigne-ai/uitoolkit` | the toolkit: `core`, `view`, `widget`, `markup`, `theming`, `engine`, `desktop/*`, `render`. No GPU library, no cgo. |
| `github.com/gascoigne-ai/uitoolkit/renderer/ebiten` | the renderer on Ebitengine (package `ebitenrender`) |
| `github.com/gascoigne-ai/uitoolkit/examples` | the showcase: every widget in one window |

No version is tagged yet. The renderer module requires the toolkit at the placeholder version `v0.0.0-00010101000000-000000000000`, which no proxy serves, so a plain `go get` of the renderer does not resolve. Until the first tag, clone the repository and point your module at the clone:

```
require (
	github.com/gascoigne-ai/uitoolkit v0.0.0-00010101000000-000000000000
	github.com/gascoigne-ai/uitoolkit/renderer/ebiten v0.0.0-00010101000000-000000000000
	github.com/hajimehoshi/ebiten/v2 v2.10.5
)

replace github.com/gascoigne-ai/uitoolkit => ../uitoolkit
replace github.com/gascoigne-ai/uitoolkit/renderer/ebiten => ../uitoolkit/renderer/ebiten
```

Then run `go mod tidy`. Ebitengine needs cgo and, on Linux, the X11 and OpenGL development packages (`libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config` on Debian and Ubuntu).

To see every widget before you start: `go run ./examples/showcase` in the clone.

## A minimal app

One window with a form, a button that shows a toast, a message dialog, a popup menu and tooltips. It compiles against the modules above.

```go
package main

import (
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	ebitenrender "github.com/gascoigne-ai/uitoolkit/renderer/ebiten"
	"github.com/gascoigne-ai/uitoolkit/theming"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// app is the whole program state. Widget state (fields, scroll positions,
// the tooltip delay) lives here, because the view is rebuilt every frame.
type app struct {
	ctx     *core.Ctx
	sampler core.Sampler
	catalog *theming.Catalog
	prefs   theming.Prefs

	width, height int // the window in logical pixels

	name     TextField
	tooltip  TooltipState
	toasts   ToastStack
	dialog   *DialogState // nil while closed
	dialogAt MoveState
}

func main() {
	device := ebitenrender.NewDevice()
	catalog := theming.NewCatalog(theming.Builtins())
	prefs := theming.Prefs{Theme: "dark"}
	ctx, err := core.NewCtx(catalog.Theme(prefs), device)
	if err != nil {
		log.Fatal(err)
	}
	a := &app{ctx: ctx, catalog: catalog, prefs: prefs, name: NewTextField("")}
	ebiten.SetWindowSize(720, 480)
	ebiten.SetWindowTitle("Hello")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebitenrender.Run(a, nil); err != nil {
		log.Fatal(err)
	}
}

// Layout implements render.Host: the canvas is the window in device pixels,
// the view works in logical pixels.
func (a *app) Layout(width, height int, device float64) (int, int) {
	canvasW, canvasH := int(math.Ceil(float64(width)*device)), int(math.Ceil(float64(height)*device))
	_ = a.ctx.SetScale(device) // never fails: the fonts are loaded; multiply by your own interface scale here
	a.width, a.height = int(math.Ceil(float64(canvasW)/device)), int(math.Ceil(float64(canvasH)/device))
	return canvasW, canvasH
}

// Update implements render.Host: one tick of input.
func (a *app) Update() error {
	ctx := a.ctx
	ctx.Tick++
	ctx.In = a.sampler.Sample(ebitenrender.ReadInput()).Scaled(ctx.Scale)
	ctx.Cursor = core.CursorDefault
	if ctx.In.RestartsCaret() {
		ctx.RestartCaret() // without this the caret never blinks
	}
	a.toasts.Prune(time.Now())
	if ctx.In.Escape { // the toolkit leaves Escape to the app: top layer first
		if MenuOpen(ctx) {
			CloseMenu(ctx)
		} else {
			a.dialog = nil
		}
	}
	PopupMenuKeys(ctx, ctx.In)
	root := a.view()
	Layout(ctx, root, a.width, a.height)
	Update(ctx, root)
	return nil
}

// Draw implements render.Host: build, lay out and paint the same tree.
func (a *app) Draw(screen render.Canvas) bool {
	root := a.view()
	Layout(a.ctx, root, a.width, a.height)
	Draw(a.ctx, root, screen)
	return true
}

// Cursor implements render.Host.
func (a *app) Cursor() render.Cursor { return a.ctx.Cursor }

// view is the window: the content, then the layers over it, top last.
func (a *app) view() Node {
	var dialog Node // Go evaluates If's argument even when false: build it only when open
	if a.dialog != nil {
		dialog = MessageDialog(a.dialog, Moved(&a.dialogAt), OnDismiss(func() { a.dialog = nil }))
	}
	return Stack(
		a.content(),
		dialog,
		Toasts(&a.toasts, a.width, a.height, 0),
		PopupMenuLayer(a.ctx),
		TooltipLayer(&a.tooltip),
	)
}

func (a *app) content() Node {
	return Col(Bg(core.Bg),
		Toolbar(
			Caption("Hello"),
			Spacer(),
			IconButton(GearIcon(), WOf(core.IconBtnOf), HOf(core.CompactOf), Tooltip("Theme"),
				OnClickAt(func(posX, posY int) { a.themeMenu(posX, posY) })),
		),
		Col(Grow(1), PadOf(core.GutterOf), GapOf(core.LooseGapOf),
			FormRow("Name", 80,
				Field(&a.name, Grow(1), Bg(core.Sunken), Placeholder("Your name"), OnEnter(a.greet), TextMenu(&a.name)),
			),
			Row(HOf(core.RowOf), GapOf(core.TightGapOf),
				Spacer(),
				Button("Ask", HOf(core.RowOf), OnClick(a.ask)),
				PrimaryButton("Greet", HOf(core.RowOf), OnClick(a.greet)),
			),
		),
	)
}

func (a *app) greet() {
	a.toasts.Show("Hello, "+a.name.Text, ToastInfo, time.Now())
}

func (a *app) ask() {
	a.dialog = &DialogState{
		Title: "Clear the name?",
		Body:  "The field goes back to empty.",
		Buttons: []DialogButton{
			{Label: "Cancel", Run: func() { a.dialog = nil }},
			{Label: "Clear", Primary: true, Run: func() { a.name = NewTextField(""); a.dialog = nil }},
		},
	}
}

func (a *app) themeMenu(posX, posY int) {
	var items []MenuItem
	for _, choice := range a.catalog.All() {
		id := choice.ID
		items = append(items, MenuItem{Label: choice.Label, Run: func() {
			a.prefs.Theme = id
			_ = a.ctx.SetTheme(a.catalog.Theme(a.prefs)) // never fails: the fonts are loaded
		}})
	}
	OpenMenu(a.ctx, posX, posY, items)
}
```

What each part is for:

- **Frame loop.** `ebitenrender.Run` calls the `render.Host` methods: `Layout` on resize and scale change, `Update` once a tick (60 a second by default), `Draw` once a frame. Every `Update` and every `Draw` builds the tree again, lays it out and then updates or paints it. The tree is a pure function of the app state: there is no diffing and no retained widget object.
- **Input.** `core.Sampler.Sample` turns the renderer's report into `core.Input` and counts double and triple clicks. `Scaled` brings the pointer into logical pixels. Set `Sampler.Repeat` from `desktop/keyrepeat` to repeat held keys at the desktop's speed.
- **Caret.** The zero `ctx.Caret` is a blinking block, but it blinks only after the first `ctx.RestartCaret()`. Call it when `ctx.In.RestartsCaret()` is true, and again when you turn `Caret.Steady` off.
- **Cursor.** Widgets ask for a cursor shape in `ctx.Cursor` during `Update`. Reset it to `core.CursorDefault` at the start of each tick; `Run` shows what `Host.Cursor` returns.
- **Layers.** The root is a `Stack`: content first, then the dialog, the toasts, the popup menu and the tooltip layer last. See [Layers](#layers).
- **Theme.** `theming.Catalog.Theme(prefs)` turns preferences into a `core.Theme`; `ctx.SetTheme` applies it.
- **Pacing.** The sample draws every frame. To save power, package `engine` decides when a frame must be drawn (`MustDraw`, at least `MinFPS` 10 frames a second, so the caret's phase still shows) and how fast to update (`Pacer`: 60 ticks a second, 30 after 60 idle ticks). Applying the rate (`ebiten.SetTPS`) and returning false from `Draw` is the app's job.

## What the app owns

The toolkit draws and lays out. These stay with the app:

| Concern | What the toolkit gives | What the app does |
|---|---|---|
| State | state types (`TextField`, `ScrollState`, `SplitState`, `MoveState`, `ToastStack`, ...) | keeps them in its own structs, so they outlive the per-frame rebuild; saves what should persist |
| Escape | nothing for menus or dialogs | closes the top layer: the popup menu first, then the dialog |
| Focus | Tab order inside a `FocusScope`; one focus owner | sets `TextField.Focused` to move focus; wires any other keys |
| Caret blink | `Ctx.CaretVisible`, `RestartCaret` | calls `RestartCaret` on `Input.RestartsCaret` |
| Shortcuts | `engine.Binding` (parse, match, print), `Input.Pressed` | owns the keymap and what each key does |
| Window | `render.Host` | title, size, icon (`desktop.SetDockIcon` on macOS), quitting |
| Scale | `Ctx.SetScale`, `desktop/displayscale` | picks the device factor times its own interface scale |
| Theme choice | `theming.Catalog`, `Prefs`, user theme files | stores the preferences, offers the choices |
| Files | `desktop/filedialog` runs the system's dialog program | runs it off the frame loop; has a fallback when `Available` is false |
| Clipboard | `core.SystemClipboard` (wl-copy, xclip, xsel or pbcopy) | swaps `ctx.Clipboard` for its own on a platform without those tools |
| Accessibility | contrast, text size, high contrast, steady caret | everything else (see [Accessibility](#accessibility)) |

## Rules

- **Tokens, never literals.** Colors come from `core` tokens (`core.Bg`, `core.Muted`, ...), sizes and spacing from the `Sizes` roles in this page. A size no role fits is a new row in the core size table with its reason, never a bare number in a view.
- **Sizes that follow the text and the theme.** Write sizes as roles (`HOf(core.RowOf)`, `WOf(core.IconBtnOf)`, `PadOf(core.GutterOf)`, `GapOf(core.LooseGapOf)`), so a larger text size and a theme's `metrics` reach them. Code that has a `Ctx`, such as a `Draw` method, reads `ctx.Sizes`. `core.DefaultSizes()` holds the values at the default text size and theme; a number taken from it does not follow either.
- **Corners through the theme.** A component takes its fill, radius and outline from a control style (`Control("button")`) or `ctx.Radius(name)`, so the user's corner setting reaches it. A square fill inside a rounded panel needs `RoundT`, `RoundB` or `RoundBL` with the panel's control name.
- **Icon-only controls have a tooltip** (`Tooltip("...")`); name the shortcut in it when there is one, as the dialog's close button says "Close (Esc)".
- **One primary action per place.** `PrimaryButton` (or `DialogButton.Primary`) marks the action Enter runs; everything else is a `Button`.
- **Layout is a flexbox subset.** `Row` and `Col` lay out, `Grow` shares the free room, `Gap` spaces siblings, `Pad` insets, `Stack` overlays. Prefer `Gap` and `Pad` to spacer elements of fixed size.
- **Build optional parts with Go's `if` when they are costly or need a non-nil value.** `If(cond, node)` is a function call: Go builds `node` even when `cond` is false. A nil `Node` in a container adds nothing.

## Sizes

Every shared size is a rule in one table in `core` (`sizes.go`), never a free number. Each has one name in three places: the Go role (`core.GutterOf`), the template scope (`{{gutter}}`) and the theme file (`"gutter"`); the field of `ctx.Sizes` is `Gutter`. The rules start from the line heights of the faces and from the theme's knobs (`core.Metrics`):

| Knob | Default | Range | What it is |
|---|---|---|---|
| line | 17 | 13..28 | line height of the UI face (17 is Go regular at 15 px); measured, not a theme value |
| small line | 15 | 1..line | line height of the small face (15 is Go regular at 13 px); measured |
| `unit` | 2 | 1..4 | the grid in pixels every derived padding and gap sits on |
| `leading` | 0 | 0..8 | extra pixels of leading on the line and the small line |
| `steps` | 1, 2, 3, 4, 5, 6, 8 | 1..16 each | the spacing steps, in units |

**Leading** is added to the height classes, the bars built from them, and between the wrapped lines of `Paragraph`. A single-line `Text` and the checkbox keep the measured line, and the lines of `Code` and `Area` keep their own spacing.

**Height classes** are a line plus leading plus padding on both sides, so text sits on whole pixels (with an odd leading or unit, a pixel can be left over):

| Name | Role | Rule | Default | Pin range | Use |
|---|---|---|---|---|---|
| `rowH` | `RowOf` | line + leading + 6 units | 29 | 8..160 | list and table rows, buttons, fields, form rows |
| `compactH` | `CompactOf` | line + leading + 4 units | 25 | 8..160 | icon buttons, chips, inline fields, menu rows, sliders |
| `labelH` | `LabelOf` | line + leading + 3 units | 23 | 8..160 | captions; each line of `Lines` |
| `slimH` | `SlimOf` | line + leading + 2 units | 21 | 8..160 | pills, the smallest buttons |
| `noteH` | `NoteOf` | small line + leading + 2 units | 19 | 8..160 | one-line notes (`Note`), small `Lines` |

**Spacing roles** are each one step. The defaults are the step in pixels, `units * unit`:

| Step (default px) | Name | Role | Pin range | Where it goes |
|---|---|---|---|---|
| `hair` (2) | `tabLead` | `TabLeadOf` | 0..32 | room before the first tab |
| | `edgeW` | `EdgeWOf` | 1..8 | open tab outline |
| | `caretW` | `CaretWOf` | 1..8 | bar or underline caret |
| `near` (4) | `nearGap` | `NearGapOf` | 0..64 | items that read as one: a label and its count |
| | `shadowOffset` | `ShadowOffsetOf` | 0..16 | how far a shadow is shifted |
| `tight` (6) | `tightGap` | `TightGapOf` | 0..64 | controls side by side |
| | `tabGap` | `TabGapOf` | 0..64 | room above tabs |
| | `ruleInset` | `RuleInsetOf` | 0..64 | menu separator |
| | `edgeGrab` | `EdgeGrabOf` | 1..32 | resize grab inside a border |
| `loose` (8) | `looseGap` | `LooseGapOf` | 0..64 | blocks stacked in a pane |
| | `cellPad` | `CellPadOf` | 0..64 | table cells |
| `label` (10) | `labelPad` | `LabelPadOf` | 0..64 | side inset of a tab or pill label |
| | `tabSepInset` | `TabSepInsetOf` | 0..64 | line between tabs |
| `gutter` (12) | `gutter` | `GutterOf` | 0..64 | inset between a container and its content; gap between its sections |
| `dialog` (16) | `dialogPad` | `DialogPadOf` | 0..64 | inset of a dialog's title bar, toolbar and body |

`ctx.Sizes.Steps` holds the steps in pixels. The switch's track is inset by the `hair` step.

**Bars, fixed shapes and dialogs:**

| Name | Role | Rule | Default | Pin range |
|---|---|---|---|---|
| `toolbarH` | `ToolbarOf` | `rowH` + 2 `nearGap` | 37 | 8..320 |
| `tabBarH` | `TabBarOf` | `rowH` + 2 `tabGap` | 41 | 8..320 |
| `dialogToolbarH` | `DialogToolbarOf` | `rowH` + 2 `tightGap` | 41 | 8..320 |
| `dialogTitleH` | `DialogTitleOf` | `rowH` + 2 `looseGap` | 45 | 8..320 |
| `iconBtnW` | `IconBtnOf` | `compactH` + 2 `nearGap` | 33 | 8..320 |
| `wordBtnW` | `WordBtnOf` | `iconBtnW` + `gutter` | 45 (a button with one short word) | 8..320 |
| `checkboxSize` | `CheckboxOf` | small line | 15 | 4..64 |
| `dialogW` | `DialogWOf` | 190 units | 380 (a message dialog) | 120..1600 |
| `dialogButtonMinW` | `DialogButtonMinWOf` | 5 `dialog` steps | 80 (least width of a dialog button) | 0..400 |
| `dialogMargin` | `DialogMarginOf` | 40 | 40 | 0..200 |
| `dialogMaxH` | `DialogMaxHOf` | 560 | 560 | 120..4000 |

`dialogMargin` (room around a dialog) and `dialogMaxH` (tallest default dialog) are for app layouts; no toolkit widget applies them, so a theme that sets them changes only a layout that reads them.

**Sizes follow the faces and the theme.** `ctx.Sizes` is `core.SizesFor(line, small, ctx.Th.Metrics)` for the line heights of the faces the context really draws with. `NewCtx`, `SetTheme` and `SetFonts` recompute it; `SetScale` does not, since the scale multiplies at draw time. The UI line is clamped to 13..28 (`MinLine`, `MaxLine`) and the small line to 1..UI line; spacing does not move with the text size. `DefaultSizes()` is the result for lines 17 and 15 with the default knobs.

**In templates** the names in the tables come from `ctx.SizesScope()`. Pass it as a fallback scope: `tpl.MustBuild(scope, ctx.SizesScope())`, then write `h="{{rowH}}"` or `pad="{{gutter}}"`.

**A new component** picks a height class and a spacing role. A size that fits no class or role is a new row in the size table in `core/sizes.go` with its reason.

## Color roles

Tokens are names; a theme maps them to colors. An unknown token draws in `Text`, so a typo stays visible.

| Role | Tokens | Use |
|---|---|---|
| Surfaces, back to front | `Frame`, `Bg`, `Panel`, `Sunken` | `Frame`: window chrome, bars, tooltips. `Bg`: panes, dialogs, menus, the open tab. `Panel`: table headers, toasts, floating info. `Sunken`: lists, fields, buttons' fill, code. |
| Row states | `Hover`, `Selected`, `Alt` | the row under the pointer, the chosen row, every second row of a zebra table |
| Lines | `Border`, `BorderStrong`, `BorderSoft`, `Rule` | `Border`: bars, dividers, `HRule`/`VRule`. `BorderStrong`: control outlines. `BorderSoft`: hairlines between table rows, `Separator`, dialog toolbar line. `Rule`: an unchecked checkbox, the slider's middle tick. |
| Text | `Text`, `Muted`, `Faint`, `ChipText` | body text; secondary text and icons; the quietest text (decoration only); text on a button or chip |
| Accent | `Accent`, `AccentLine`, `AccentText`, `AccentSoft` | primary fill, checkbox, caret; focus ring, open tab, links; text on `Accent`; a pale accent background |
| Status | `OK`, `Warn`, `Error`, `Info` and `OKSoft`, `WarnSoft`, `ErrorSoft`, `InfoSoft` | status text; their pale backgrounds for pills |
| Banner | `Hero`, `HeroText`, `HeroMuted` | the flat fallback of the `hero` gradient behind a dialog title; text and icons on it |
| Text selection and search | `Selection`, `FindMatch`, `FindActive` | selected text; search matches; the current match |
| Other | `Shadow`, `Chip`, `Track` | scrim and drop shadows (with alpha), key chips and slider track, the empty part of a switch |

- Text on a banner uses `HeroText` and `HeroMuted`, never `Text` and `Muted`.
- Put information in `Text` or `Muted`. `Faint` has no contrast floor.
- An app adds its own tokens in a theme file (`"colors": {"status.busy": "#..."}`) and names them as `core.Token("status.busy")`.

## Typography

| Face | Field | Default size | Range | Use |
|---|---|---|---|---|
| UI | `ctx.F.UI` | 15 | 10..24 | labels, body text, fields |
| Small | `ctx.F.Small` | 13 | 8..24 | captions, hints, notes, shortcuts, badges, tooltips, table headers |
| Code | `ctx.F.Mono` | 14 | 8..32 | `Code`, `Area`, `Mono()` text |
| Bold, SmallBold | `ctx.F.Bold`, `ctx.F.SmallBold` | as UI, Small | | `Bold()` |

- The faces are the embedded Go fonts (regular, bold, mono) unless the app calls `ctx.SetFonts` with other font data; `desktop/fonts` finds installed families. Fonts that cannot be read change nothing.
- Sizes are whole pixels and come from `theming.FontSizes{UI, Small, Code}` in `Prefs.Fonts`; 0 is the default and a size out of range is clamped. The three are independent. Theme files do not set sizes: a `fontSize`, `smallSize` or `monoSize` key loads and is ignored.
- **Bold** marks names that structure the window: dialog titles, table headers. Labels and list items stay regular.
- **Secondary text** is small and muted: `Caption` (small, upper case, muted), `Note`, `Paragraph` (small, muted, wraps), `DialogSubtitle`, menu hints and shortcuts, tab badges.
- **Long text** in a fixed width gets `Ellipsis()`; text that should wrap is a `Paragraph`. `Lines` stacks the lines of a string.
- **Tabs in text** are display only: the text keeps the real tab. In `Code` and `Area` a tab advances to the next stop of `core.TabStop` (4) columns; in proportional text it is a gap of four spaces.
- **Rendering.** Glyphs are unhinted outlines at fractional x positions, with the baseline on a whole device pixel. `ctx.Text` draws them twice, the second pass at alpha `core.StemBoost` (0.5), so partly covered edge pixels get heavier: a pixel at coverage a ends at a + 0.5·a·(1−a). Outlines and advances do not change, so `Width`, `Advance` and `LineH` measure exactly what is drawn.

## Shapes and lines

The built-in themes give these control styles. `Control(name)` on any element applies the fill gradient, radius and outline of that style; a name the theme does not list draws plain.

| Control | Radius | Outline | Used by |
|---|---|---|---|
| `button`, `icon-button` | 4 | `border-strong` | `Button`, `IconButton`, `SegmentChip`, `KeyChips` |
| `primary` | 4 | `accent-line` | `PrimaryButton`, the chosen `SegmentChip` |
| `field` | 4 | `border-strong` | `Field` |
| `tooltip` | 4 | `border-strong` | the tooltip box |
| `row` | 4 | none | `ListRow`, menu rows, a dialog's close button |
| `slider` | 4 | none | the slider thumb |
| `table` | 6 | `border-soft` | `Table` |
| `menu` | 8 | `border-strong` | `Menu`, `Dialog` |
| `scrollbar` | 5 | none | scrollbar thumbs |
| `checkbox` | 3 | none | `Checkbox` |
| `hero` | 0 | none; gradient `hero-top` to `hero-bottom` | `DialogTitle` |
| `pill` (999), `swatch` (3, `border-strong`), `match` (2), `tabstrip`, `statusbar` | | | defined for apps; no toolkit widget uses them |

- The corner setting multiplies every radius: square 0, soft 1 (the theme's own), round 1.5 (`theming.CornerChoices`). Flat (`Prefs.Flat`) turns gradients into flat fills.
- `Switch` and the slider track are pills while corners are soft or round, and square when corners are square.
- Toasts and tabs take no control, so they stay square under every corner setting.
- Borders are one-pixel hairlines in a token, drawn `ctx.LineWidth()` device pixels thick (the scale rounded, at least 1). Edges snap to device pixels, so neighbours share an edge.
- The open tab's outline is the separate `Edge` family, `edgeW` (2 px) thick (`EdgeT`, `EdgeL`, `EdgeR`, `EdgeB`).
- Bars are separated by `Border`, never `Rule`.
- Floating panels (dialogs, menus, toasts) cast a `Shadow` shifted by `shadowOffset` (4), not blurred.

## Icons

- Icons are SVG files drawn as vector paths in the element's color (`Color(...)`, muted by default), with the icon's own hover color. `HoverColor(token)` replaces that hover color where it would not contrast, as on a banner.
- The built-in set: `arrow-down`, `arrow-up`, `check`, `chevron-down`, `chevron-up`, `clipboard`, `copy`, `cross`, `dot`, `dots`, `external`, `folder`, `gear`, `grip`, `history`, `pencil`, `plus`, `prompt`, `redo`, `replay`, `save`, `scissors`, `search`, `trash`, `triangle-down`, `triangle-right`, `undo`, `unlink`. Wrappers exist for the common ones (`TrashIcon`, `GearIcon`, `CrossIcon`, ...).
- `widget.RegisterIcons(fsys)` adds or replaces icons from a folder of `*.svg` files. `Icon(name)` panics on an unknown name and lists the valid ones: that is a mistake in the program.
- A theme's `iconStroke` multiplies every stroke (0.8 thin, 1.3 bold, at most 4), so icon weight is a theme choice.
- Icons decorate in `Muted`; icons on the banner use `HeroMuted`. A menu reserves a 14 px icon column when any item has an icon, so labels line up.
- An icon from an open set goes in with its license file.

## States

| State | Look | Who does it |
|---|---|---|
| Rest | click targets are outlined, drawn after their children: buttons, fields and menus in `BorderStrong`, tables in `BorderSoft`, primary buttons in `AccentLine` | the control style |
| Hover | buttons brighten their fill (`Lighten`); rows and menu items fill with `Hover` (`HoverBg`); an unchecked checkbox outline goes from `Rule` to `Muted`; a scrollbar thumb from `Muted` to `Text`; the slider thumb to `AccentLine` | the widget |
| Selected | `Selected` fill on list, table and menu rows; the open tab as in [Tabs](#widgets-when-to-use-which); the chosen `SegmentChip` as a primary button | the app passes which one |
| Focus | a focused `Field` draws its outline again in `AccentLine` over the resting one, at line width; a focused `Area` or `Code` gets a square `AccentLine` outline | the widget |
| Disabled | `MenuItem.Disabled`: `Muted` text, no hover, not pickable. `TableRow.Muted`: `Muted` text. `DisabledSwitch`: the off colors, ignores clicks. | the app; there is no disabled option for buttons |
| Dragging | `core.CursorMove` from press to release, never on hover | the widget |
| Under a modal | no hover and no clicks below the topmost modal | the runtime |

For a button that cannot act now, build a muted `Button` without `OnClick` and say why in its tooltip.

## Widgets: when to use which

Everything below is in package `widget` unless it says `view`. The package docs of `view`, `widget` and `markup` index every element and option.

**Buttons.**

| Widget | Use |
|---|---|
| `PrimaryButton` | the one main action of a dialog, form or bar |
| `Button` | every other action; give it `HOf(core.RowOf)` (or `CompactOf` in a toolbar) |
| `LinkButton` | a quiet action inside text or a list; `AccentLine` text, `Hover` under the pointer |
| `IconButton` | an action that has a clear icon; `WOf(core.IconBtnOf)`, `HOf(core.CompactOf)` and a `Tooltip` |
| `SegmentChip`, `SegmentIconChip` | one of a few exclusive choices, side by side |

A button's `W` is a minimum: the button widens to show its label and cuts it with an ellipsis only when the row has no room. Button labels are padded 10 on each side.

**Text input.**

| Widget | Use |
|---|---|
| `Field` | one line; give it `Bg(core.Sunken)` and `HOf(core.RowOf)` (`CompactOf` inline). `Placeholder`, `OnEnter`, `OnTab`, `Styled` (colored spans), `TextMenu` (right-click Undo, Redo, Cut, Copy, Paste, Select all). Set `TextField.Masked` for secrets: the field shows bullets and never copies or cuts its text. |
| `Area` | several lines; keeps Tab for indenting, Escape leaves it |
| `Code` | read-only text with spans, folds, line numbers, selection and search matches |
| `NotedFields` | the fields of a form that refuses bad input, each with its refusal note |

**Choices.**

| Widget | Use |
|---|---|
| `Checkbox`, `CheckRow` | an option in a list of options; `CheckRow` makes the whole row the click target |
| `Switch` | a setting that takes effect at once; `DisabledSwitch` when it cannot change now |
| `Slider` | a number in a range where the effect can be seen while dragging (`OnSlide`), committed on release (`OnSlideEnd`); `SliderMid` puts a value in the middle |
| `SegmentChip` row | two to about five exclusive choices |
| `Menu` / `OpenMenu` | a longer list, or actions on something right-clicked |

All of them take `OnChange` (or `OnClick`) and leave the value to the app: the app flips its own bool and the next frame shows it.

**Rows, bars and lists.**

| Widget | Use |
|---|---|
| `ListRow` | a row of a list; `Selected` when chosen, `Hover` otherwise |
| `FormRow` | a label in a fixed column and its controls; rows with the same label width line up |
| `Toolbar` | a bar of controls over a pane: `toolbarH`, `gutter` inset, `Border` under it |
| `Note` | a one-line message under a control: `core.Error` for a problem, `core.Muted` for a hint |
| `Table` | rows with columns: uppercase small bold muted header on `Panel`, rows `rowH` on `Sunken`, a `BorderSoft` hairline between rows, the row under the pointer lit, cells padded `cellPad`. Opt-ins: `Zebra`, draggable column dividers (`State`, columns at least `MinColumnPx` 56), editable cells, a lead column, a detail under a row, row drag handles. |
| `KeyChips` | a key combination such as `Ctrl+Shift+Tab`, one chip per key |

**Tabs.** `TabBar` holds `Tabs` (or `ScrollTabs` when the labels may not fit). The bar is `tabBarH` with `tabGap` above the tabs and a `Border` line under the inactive ones; inactive labels are `Muted`. The open tab has the pane color `Bg`, its label in `Text` and a 2 px `AccentLine` outline on top, left and right, open at the bottom so it runs into its pane. `ActiveTab()` gives a hand-built tab the same look. No tab label is bold.

**Containers.**

| Widget | Use |
|---|---|
| `Scroll` | content taller than its box; an 8 px track, a thumb at least 24 long, drawn only when the content overflows; one wheel notch is 48 px |
| `HScroll` | a row wider than its box (tab strips, toolbars in a narrow window): arrows only at an end that can still scroll, wheel, sideways touchpad swipe, drag to pan |
| `Split` | two panes with a draggable `Border` divider (resize cursor); `Vertical`, `AsRatio`, `MinFirst`, `MinSecond` |
| `Resizable` | a box the user resizes by its edges |

**Dialogs.** A dialog is a modal panel inside the window: `Dialog` or `ResizableDialog`, centered over a `Shadow` scrim.

- **Title bar** (`DialogTitle`): `dialogTitleH`, padded `dialogPad`, the `hero` gradient with its top corners rounded to the panel, a `Border` line under it. From the left: the bold title in `HeroText` (a short noun), an optional `DialogSubtitle` (small, `HeroMuted`, gives way first), free room, then the close button (`CrossIcon`, tooltip "Close (Esc)"). No other controls in the bar.
- **Toolbar** (`DialogToolbar`): `dialogToolbarH`, padded `dialogPad`, gap `looseGap`, a `BorderSoft` line under it. A field that should take the room gets `Grow(1)`.
- **Body**: padded `dialogPad`.
- **Message dialog** (`MessageDialog`): `dialogW` wide; the body text in `Muted`, any form you add, an optional input line, then the buttons on the right (gap `looseGap`, at least `dialogButtonMinW` wide, the primary one lit). Enter in the input runs the primary button.
- **Closing.** The close button runs `OnDismiss`. A click outside a dialog is swallowed and leaves it open (`KeepOnOutsideClick`), so a stray click cannot throw away the user's work. Escape is the app's to wire (see [What the app owns](#what-the-app-owns)); make it do what the close button does.
- **Moving.** Pass `Moved(&state)` to let the user drag the dialog by its title bar; a double click on the bar centers it again. Without `Moved` the bar does not drag.

**Toasts.** `Toasts(&stack, width, height, bottomInset)` shows a `ToastStack` in the bottom right corner, above `bottomInset` (a status bar's height, or 0): 320×36 each, 6 apart, at most 5 (`ToastMax`), the oldest on top, on `Panel` with a shadow. Info toasts have a `Border` outline and go after 3 s (`ToastInfoFor`; call `Prune` each tick). Error toasts have `Error` text and outline, stay until clicked, and the same error is not shown twice. A click dismisses any toast. Use a toast to confirm something done or report an error the user did not wait for; use a dialog when the user must decide.

**Tooltips.** `view.Tooltip(text)` on any element; `TooltipLayer` shows the innermost one after the pointer rests 30 ticks (half a second at 60 ticks a second). The box is `Frame` with the `tooltip` outline and small `Text`, 16 px below and right of the pointer, kept inside the window, and above the pointer when there is no room below.

**Menus.** `OpenMenu(ctx, x, y, items)` opens the popup menu; mount `PopupMenuLayer(ctx)` once and call `PopupMenuKeys` each tick. The menu is a `Bg` panel with the `menu` style and a shadow, inset `MenuPad` (4), as wide as its widest row within `MenuMinW` 120 and `MenuMaxW` 420, kept inside the window. Rows are `compactH` high, padded 10; the keyboard's row is `Selected`, others `Hover` under the pointer. `MenuItem` has `Shortcut` (shown small and muted, "Ctrl" written "Cmd" on macOS), `Hint` (muted lines under the label), `Icon`, `Disabled` and `Separator` (9 high). A pick or a click outside closes it.

**Other.** `Image` draws a `render.Image` scaled to fit. `Paragraph` wraps small muted text, with the theme's `leading` between its lines. `view.HRule`, `view.VRule` and `view.Separator` draw dividers.

## Keyboard and focus

- **What takes focus.** Only the text widgets: `Field`, `Area` and `Code`. Buttons, checkboxes, switches, sliders, tabs and table rows have no keyboard path in the toolkit.
- **Moving focus.** A click in a field focuses it; a click elsewhere blurs it. The app moves focus by setting a state's `Focused`; the state that asked last wins, and only one state has focus (`ctx.FocusOwner`).
- **Tab and Shift+Tab** move focus forward and back in tree order, wrapping, inside the topmost `FocusScope` (every dialog is one). Focus moved by Tab is scrolled into view in any enclosing `Scroll`. Ctrl+Tab does not move focus. An `Area` keeps Tab (indent, Shift+Tab dedent), and so does a `Field` with `OnTab`. There is no other focus order than tree order.
- **Escape.** A `Field` or `Code` clears its selection and keeps focus; an `Area` gives up focus. Closing menus and dialogs is the app's.
- **Enter** runs a field's `OnEnter` (not with Ctrl, which is left to app shortcuts) and picks the highlighted menu row through `PopupMenuKeys`. Up and Down move the menu highlight, skipping separators and disabled rows.
- **Editing keys** follow each platform: Ctrl+arrows and Ctrl+Backspace act on words (Option on macOS), Command+arrows act on the line on macOS, Ctrl+Home and Ctrl+End go to the document's ends. Ctrl (Command on macOS) with A, C, X, V, Z, Y (or Shift+Z) is select all, copy, cut, paste, undo and redo.
- **Key repeat.** A held editing key fires at the press, again after 25 ticks, then every 3, unless `Sampler.Repeat` holds the desktop's timing (`desktop/keyrepeat`).
- **App shortcuts.** `engine.ParseBinding("Ctrl+Shift+K")`, `Binding.Matches(ctx.In, key)` and `Binding.String()` handle combinations. Stored shortcut names come from `render.Key.String()` and stay stable. Show the shortcut in the tooltip or menu item of the action.

## Pointer

- **Clicks.** A left click reaches every `OnClick` under the pointer, innermost first, then its ancestors. A right click reaches only the innermost `OnRightClick`, so one right click opens one menu. `OnDoubleClick` runs on the second click within 400 ms and 4 px; a third makes a triple click (a field selects its line).
- **Clipping hides the pointer.** While the pointer is outside a `Clip` or `Scroll` element, its children are not hovered or clicked; the click still reaches fields inside so they can blur.
- **Wheel.** The innermost scroller under the pointer whose content overflows takes the wheel; an enclosing one does not move too.
- **Modal.** While a modal layer is visible, nothing under it gets clicks, wheel or hover.
- **Cursor.** Split dividers and resizable edges show resize cursors; a drag shows `CursorMove` only from press to release; an idle drag handle shows the default cursor, so it reads as a label, not a control.
- **Resizing.** A resizable box is grabbed within `edgeGrab` (6) inside or 4 outside its border.
- **Drags** follow the pointer while the button is held, also after it leaves a clipping parent. Only the element that took the press follows it there; every other element gets no click, drag or `OnDragOver` through the clip. A `Custom` with its own drag reads `ctx.PointerX`/`PointerY` while it runs.

## Layers

The window's root is a `Stack`; children later in it are drawn over earlier ones and get input first. Use this order:

1. the content
2. a dialog, when one is open (modal)
3. `Toasts` (not modal)
4. `PopupMenuLayer(ctx)` (modal; closes on a click outside)
5. `TooltipLayer(&state)`, always last

- At most one popup menu is open; it lives in `ctx.Menu`, so any element can open one. `OpenMenu` replaces an open one.
- A `Modal` element swallows every click; a click outside runs its `OnDismiss` unless it has `KeepOnOutsideClick` (dialogs do). `DismissMargin(px)` counts clicks near its edge as inside.
- `FloatAt(x, y)` places a floating element and keeps it inside its `Stack`; `FloatCenter()` centers it; `FloatTR(top, right)` pins it to the top right.
- An app that skips drawing while idle keeps drawing while `TooltipState.Waiting()` is true, and when `ctx.CaretReads` moved during the last draw (the caret blinks). `view.ModalLayer`, `DrawBelow` and `DrawFrom` let it keep the part under a modal in an image and redraw only the layers.

## Scaling

- Widgets, layout and hit tests work in **logical pixels**. `ctx` multiplies by `ctx.Scale` when it draws, so HiDPI stays sharp, and `Input.Scaled` maps the pointer back.
- `render.Host.Layout` receives the window size and the device scale factor. Set `ctx.SetScale(device × your interface scale)`; the canvas is the window in device pixels.
- On Linux under Xwayland the window system reports a factor of 1 on a scaled monitor. `desktop/displayscale.New().Factor()` asks the compositor (Hyprland, Sway) or `GDK_SCALE`/`QT_SCALE_FACTOR`; it returns 0 while unknown, so fall back to the renderer's factor. It never blocks a frame.
- Lines are whole device pixels at every scale; edges snap to device pixels.
- Text size is a separate setting from scale (`Prefs.Fonts`). Controls written with size roles grow with it.

## Theming pipeline

A `theming.Catalog` holds the built-in themes (`theming.Builtins()`) and the user's theme files; `Catalog.Theme(prefs)` builds the `core.Theme` for the preferences, and `ctx.SetTheme` applies it.

- **Theme files** are JSON: `name`, `extends` (`dark` or `light`), `variables` (colors referred to as `"$name"`), a flat `colors` table of token to `#rrggbb` or `#rrggbbaa`, `iconStroke`, `metrics` (see below), and `controls` (per control: `radius`, `gradient` as a strength 0..0.5 or `[top, bottom]`, `border` token). `theming.ParseTheme` merges a file over a base; unknown fields and variables, trailing data, a negative radius and a border token with no color are errors.
- **Sizes.** The optional `metrics` block sets the knobs and pins named sizes; the names, defaults and ranges are in [Sizes](#sizes). Font sizes stay the user's.

  ```json
  {
    "extends": "dark",
    "metrics": {
      "unit": 3,
      "leading": 2,
      "steps": { "hair": 1, "near": 2, "tight": 4, "loose": 4, "label": 5, "gutter": 6, "dialog": 8 },
      "sizes": { "gutter": 14, "rowH": 32 }
    }
  }
  ```

  Every key is optional. The knobs feed the rules; a size under `sizes` is fixed at those pixels whatever the knobs and the text size say, and the sizes derived from it follow it (a pinned `rowH` moves `toolbarH` and the other bars). Over a base theme, a knob the file states replaces the base's and pins merge by name, the file's winning; a base pin survives a change of a knob, and no syntax removes it. A key out of range (an explicit 0 unit or step included), an unknown key or size name, a `null` size and a number that is not an integer are errors. In Go the same is `core.Theme.Metrics` and `Metrics.Pin`.
- **User themes.** `Catalog.LoadUser(dir)` reads every `*.json` in a folder: the file name is the id, `name` the label. A file that does not parse is skipped and reported; a missing folder is no problem. `Catalog.SetUser` lists them after the built-ins.
- **Preferences** (`theming.Prefs`), applied in this order by `Catalog.Theme`: the theme (an unknown id falls back to the first built-in) and the three font sizes; the corner factor and Flat; with `Contrast: "high"`, the four customizable colors moved to 7:1 against their text, and text, muted text, `Error`, `OK`, the accent line, borders, rules and row fills moved in their own hue until each reaches its ratio on the surfaces it is drawn on (`Bg`, `Sunken` and `Frame` for text and lines, `Sunken` for `Selected` and `Hover`, `Bg` for `Chip`): text 17:1, muted 12:1, `Error` and `OK` 8:1, accent line 6:1, `Border` and `Rule` 7:1, `BorderSoft` 3:1; surfaces unchanged; then the user's color overrides for that theme (Selection, Find match, Active match, Accent), each ignored unless it parses and reaches `RequiredRatio` against its text (4.5, or 7 with high contrast).
- **Controls draw from their style**, so a theme restyles any control without code.
- **Look.** The built-in palettes have a rose or burgundy accent on warm near-black (`Dark`) or cream (`Light`) surfaces, outlined controls, flat fills and one gradient, behind dialog titles.

## Accessibility

What the toolkit gives:

- **Contrast floors**, checked by tests for both built-in themes: `Text` at least 4.5:1 on `Bg`, `Sunken`, `Frame`, `Alt`, `Selected`, `Selection`, `FindMatch` and `FindActive`; `Muted` at least 3:1 on `Bg`, `Sunken` and `Frame`; `ChipText` and `AccentText` 4.5:1 on their fills; borders, rules, the accent line, icons and row fills have their own floors. Other pairs (text on `Panel`, `Hover` or `Chip`) are not tested. Keep essential text in `Text`.
- **High contrast** (`Prefs.Contrast = "high"`), as in [Theming pipeline](#theming-pipeline).
- **Text size** from 10 to 24 px for the UI face, independent of the interface scale; controls grow with it.
- **A steady caret** (`ctx.Caret.Steady = true`) for users who do not want blinking. The caret is the only animation; toasts appear and go without motion.
- **Tooltips** on any element (`Tooltip`). The toolkit puts one only on the dialog close button: give every `IconButton` its own.

What it does not give, and an app must not claim: no screen reader support and no accessibility tree; no keyboard path to buttons, checkboxes, switches, sliders, tabs or table rows; no focus order other than tree order; no reduced-motion or system-contrast detection. An app that needs these builds them, or offers keyboard shortcuts for every action.

## Composing a screen

1. **Start from the state.** Put every value the screen shows and every widget state (`TextField`, `ScrollState`, `SplitState`, ...) in the app's struct.
2. **Pick the frame.** A `Col` on `Bg`: a `Toolbar` or `TabBar` at the top, then the body with `Grow(1)`. Use `Split` for side-by-side panes, `Scroll` for anything that can grow.
3. **Inset and space.** Inset a pane by `gutter`, space its blocks by `looseGap`, controls in a row by `tightGap`, items that read as one by `nearGap`.
4. **Size the rows.** Rows, buttons and fields `HOf(core.RowOf)`; icon buttons and toolbar controls `CompactOf`; one-line notes `Note`.
5. **Choose widgets** from [the tables above](#widgets-when-to-use-which): one `PrimaryButton` per place, tooltips on icon buttons, `Note` for a refusal under its field.
6. **Mount the layers** in the root `Stack` in the order of [Layers](#layers), and wire Escape, the caret and the menu keys in `Update`.
7. **Check it** in both built-in themes, with high contrast, at a large text size and at scale 2, in a narrow window.

The same tree can be written as a template (`markup`): attributes are the Go option names in kebab-case (`pad-x`, `border-t`), sizes come from `{{rowH}}` and the other scope names, and a `<slot>` takes any Go node. Use whichever reads better; both build the same elements.

## Known limitations

- A child of a `Row` is measured at the row's whole inner width, and a `Col` child's side margins are not taken off its measured width, so a `Paragraph` with `Grow(1)` in a row, or with `MarginX`, can be too short and draw over what follows.
- A table's header and its last row (when selected, hovered or zebra-filled) are square fills inside the table's rounded outline.
- Tooltip delay counts ticks, so it is longer while the app updates at a lower rate.
- `core.SystemClipboard` knows no tool on Windows: copy and paste fail there unless the app sets its own `ctx.Clipboard`.
- Menu hint lines are a fixed 16 px each and do not follow the small text size.
