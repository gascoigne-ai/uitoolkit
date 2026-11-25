## render: the renderer contract

`render` says what the toolkit draws through and what a renderer runs: `Canvas`, `Image`,
`Offscreen`, `Path` and `Stroke`, `FontSource` and `Face`, `Device`, `InputState`, `Key`, `Cursor`
and `Host`. `core.Ctx`, `view` and `widget` draw through it and import no GPU library
(`TestCoreImportsNoEbiten`, `TestViewImportsNoEbiten`, `TestWidgetImportsNoEbiten`, and the same for
`engine`, `markup` and `desktop/systeminfo`). Each guard also rejects any import under `renderer/`. `render.Graphics` is the record a renderer fills for the
system report (name and version of the renderer, graphics library, GPU memory, monitor, window).

Invariants:

- **Standard library only.** `render` imports no GPU library and nothing of this module;
  `render/rendertest` imports only `render` besides. `TestRenderImportsOnlyStandardLibrary` checks
  both.
- **Sub clips without translating.** `Sub(clip)` is the same surface with bounds
  `Bounds().Intersect(clip)` in the same coordinates, because widgets draw in absolute ones.
- **Device pixels, no snapping.** A canvas draws the coordinates it gets; snapping is the caller's.
  `Box` holds floats; `image.Rectangle` is for whole-pixel bounds and clips.
- **Premultiplied color.** Every `color.RGBA`, and what `ReadPixels` returns, is
  alpha-premultiplied.
- **Own images only.** A renderer panics on an `Image` it did not make.
- **Face is comparable.** It is `Source` and `Size`, written as a literal, and is a cache map key.
  `FontSource` implementations are pointer types, and it has no method that makes a face.
- **Offscreen is live, then released.** After `Release` no method may be called.
- **Stroke.** The zero cap is butt and the zero join is miter; a zero `MiterLimit` makes every join
  a bevel.
- **Keys.** `KeyUnknown` is 0 and its name is "". Every other key up to `KeyMax` has a unique name,
  its constant without `Key`, from one table (`keyNames`). Names are frozen because users store
  them in shortcut settings; numbers are not. Aggregate modifiers (`KeyControl`...) and the left and
  right ones both exist.
- **Host.Draw false keeps the previous frame.**
- **Input is states and durations.** `InputState.Down` lists the keys that are down with how many
  ticks each has been; 1 is the tick it goes down, a tap within one tick included, so "just
  pressed" is derived and not reported twice. An aggregate has 0 ticks: it is held, never pressed.
- **Fakes.** `rendertest.Canvas` and its Subs record one list of ops in paint order, each with the
  bounds it was drawn in. Its `Offscreen` panics on use after `Release`, but still answers
  `Bounds` and `Ops` so a test can read what was drawn. It holds no pixels.

## render/text: measuring without a GPU

`render/text` implements `render.FontSource` over go-text/typesetting, so layout (wrap, ellipsis,
caret) gets the same widths on every renderer and in tests. Non-test code imports only `render`,
go-text/typesetting and `x/image/math/fixed` (`TestRenderTextImports`). `render.Face` stays
`{Source, Size}`: `text.FontSource` is just another pointer source.

- **Source.** `NewFontSource(data)` parses one font file; a collection or corrupt data is an error,
  never a panic (`FuzzNewFontSource`). Ascent and descent come from the horizontal extents.
- **Measure.** The widest line, lines split at `\n \v \f \r`, NEL, LS and PS (CRLF is one break).
  Widths are summed in 26.6 fixed point at a size truncated to 1/64 px, as Ebitengine's `text/v2`
  does, and bidi runs are not reordered, which cannot change a width.
- **Chunks.** A line is shaped in chunks, cut after a sentence terminator cluster (and its marks),
  at bidi level changes, and after 4 KiB at a space or tab, because that is where Ebitengine cuts and
  a cut drops kerning across it. `chunks.go` is our own code for the same rule; it is checked
  against Ebitengine's STerm and ATerm tables (two characters newer than Go's Unicode tables are
  added by hand).
- **Concurrency.** One mutex guards the shaper, segmenter and buffers.

## renderer/ebiten: the Ebitengine renderer (own module)

`renderer/ebiten` is the module `github.com/gascoigne-ai/uitoolkit/renderer/ebiten`; its package is named
`ebitenrender` (a package named `ebiten` would clash with `hajimehoshi/ebiten/v2` in every caller). It
implements `render` over Ebitengine (`vector`, `text/v2`). Of the toolkit it imports only `render`
(`TestEbitenRenderImportsOnlyRenderAndEbiten`, non-test files). `ReadInput` reports
the tick as a `render.InputState` and drops a key with no render name; what it means is `core`'s.

- **Masks.** Rounded fills and outlines are quarters of cached anti-aliased corner masks, in one
  `DrawTriangles`. A `Device` owns the cache (64 entries, dropped whole when full). It is keyed in
  device pixels, so a change of scale or theme leaves it valid and drops nothing.
- **Faces.** A `FontSource` keeps the `text/v2` face of each size (16, then dropped whole), so
  measuring and drawing text allocate no face a call (`alloc_test.go`).
- **Shapes.** A rounded shape with a radius below one device pixel is square. A box whose edges are
  between pixels is anti-aliased.
- **Foreign images.** A bare `*ebiten.Image` satisfies `render.Image` by its `Bounds`, so it compiles
  where an image is wanted, but `DrawImage` and `DrawMask` panic with `render: image from another
  renderer`; `Device.NewImage` makes the image.
- **Graphics.** `ReadGraphics` fills a `render.Graphics` (renderer "Ebitengine", its module
  version from the build info, graphics library, GPU memory, FPS, monitor, window) and works only
  while the run loop runs. An app passes it as `systeminfo.Info{Graphics: ebitenrender.ReadGraphics}`.
- **Run-loop `TestMain`.** Ebitengine reads pixels only inside its run loop, so `main_test.go` runs
  the whole test binary inside the first `Update` (no `t.Parallel`), after it points `HOME` and the
  `XDG_*` dirs at a temp dir.
- **Skip-clear.** `Run` turns off the clearing of the screen before each frame, so a frame the host
  does not draw (`Host.Draw` false) keeps the previous one. A program with its own `ebiten.Game`
  uses `Device.NewCanvas` and handles this itself.
- **Conformance.** `TestMeasureAndMetricsMatchEbitenText` (`textconformance_test.go`, it needs the run loop, so xvfb) compares `render/text` with `text.Measure` within
  0.01 px over a corpus, three Go fonts and `renderer/ebiten/testdata/LiberationSans-Regular.ttf`. The Go fonts have
  no kerning, so the cut cases only show on Liberation Sans (SIL OFL 1.1, licence in
  `renderer/ebiten/testdata/LiberationSans-LICENSE.txt`); a guard fails the test if either side stops kerning "AV".
- **Key tables.** `KeyFromEbiten` and `KeyToEbiten` read tables built once at start by matching the
  `String()` names of `ebiten.Key` and `render.Key`, because users store shortcuts by name.
  `TestKeyNamesRoundTripEveryEbitenKey` fails when either side gains a name the other lacks.

## core: the toolkit

`core` holds the primitives (`rules.md`, Toolkit API rule 11). It is tested without reading pixels:
the draw calls are checked against `rendertest.Canvas`.

- **Logical pixels.** Widgets, layout and hit tests work in logical pixels. `Ctx` drawing calls
  multiply by `Scale`, text measures divide by it, and `Input.Scaled` maps the pointer back.
  This keeps HiDPI sharp.
- **Drawing.** `NewCtx` takes a `render.Device`, which loads the fonts; a `Face` is a loaded font
  and a size, so `SetScale` and `SetTheme` only resize and `SetFonts` alone loads. `Ctx` draw calls
  turn logical rects into device pixels and call a `render.Canvas`. `Text` is two `DrawText`
  passes, the second in the color scaled by `StemBoost`. A square gradient is not snapped, so on a
  box between pixels its edges are anti-aliased.
- **No GPU library.** The `Ctx` draw calls take a `render.Canvas`; `Clip` returns a canvas limited
  to a logical rect (device pixels, cut to the canvas). No file of `core`, a test included, imports
  Ebitengine or `ebitenrender` (`TestCoreImportsNoEbiten`); its tests use a `testDevice` built on
  `render/text` and `rendertest`.
- **Theme and tokens**: the look, the tokens and the theme pipeline are in `style.md`.
- **Input.** `Sampler.Sample` turns a `render.InputState` into the frame's `Input` and counts
  clicks (a renderer's input, `ebitenrender.ReadInput`, goes in as the `render.InputState`). The pure `inputFrom` derives
  presses, shortcuts and edit keys with their repeat; `modifiedEdit` is the per-platform table that
  turns modifiers into word and line edits. `Input.Keys` holds `render.Key`.
- **Clipboard.** `SystemClipboard` runs the platform's clipboard tools: Ebitengine has no clipboard
  API, and no dependency is added for it.
- **Text editors live in `widget`** (`doc.go`): `TextField` (one line, undo), `TextArea` (a
  `TextField` in multiline mode plus view state) and `CodeView` (read-only spans, folds, selection).
  They share the word rule (`isWordRune`), the scrollbar math and drag (`widget/scrollbar.go`,
  `barDrag`) and the monospace geometry (`monoGrid`), so the three agree on words, caret and clicks.

## desktop: the OS the window runs on

`desktop` holds what a window needs from the OS and Ebitengine does not give: the Dock
icon (`SetDockIcon`), the display scale (`displayscale`), the key repeat timing (`keyrepeat`) and the
system file chooser (`filedialog`) and the system report the settings show (`systeminfo`) and the installed fonts (`fonts`). Each is split per OS (build tags or a table keyed by OS) and its
parsing is pure and tested without a display. It imports nothing of the toolkit or the app, except `systeminfo`, which imports `render` for `render.Graphics` (`systeminfo.Graphics` is an alias of it) and shows the renderer in the first row of the Rendering section.

- **Why in the toolkit.** Every program with a window needs the same few things, and they are about the
  window and the desktop, not about requests. An `internal/platform` layer once held them
  behind ports owned by `settings` and `workspace`, but no feature code used those ports: only
  `app` and a view did, so the layer is gone. The ports live where they are used: `internal/app/ports.go`
  and `ui/overlays/settingsdialog/formpage/fontlist` (fonts), next to the fakes of the tests.
- **Sub-packages.** `displayscale`, `keyrepeat`, `filedialog`, `systeminfo` and `fonts` each have their own `source`,
  `System` and `run`; one flat package would need prefixes on all of them.
- **Rejected.** A D-Bus file chooser portal (see the `filedialog` package doc). Revisit when a
  client library is already a dependency.

## view: the view DSL

A view is a Go function from state to a tree of elements, so the tree is a pure function of the
state: no diffing and no reactive system. Elements and options are indexed in the `view` and
`widget` package docs; naming is `rules.md` (Toolkit API rules). State a widget edits
(`ScrollState`, `SplitState`, `TextField`...) is owned by the app, so it survives a rebuild.

### The frame builds once

- `Component.endFrame` builds the tree, lays it out, runs `view.Update` and keeps it (`drawTree`).
  `Draw` lays it out again (drags and scrolls move) and paints it. It builds again only when
  `treeStale` is set; `setSize` drops `drawTree`.
- `treeStale` comes from explicit signals, not a revision counter, because one missed bump would
  show old content. `Ctx.Fired` counts the handlers that ran: `view` for its `On*` options, a
  widget where it calls a handler or changes state the app reads at build (a column width). Each
  step after `Update` that changes what the view shows sets the flag itself (listed in `endFrame`).
- A view that builds only the rows in view (the sidebar tree) reports from its `Tick` that a wheel
  step or a thumb drag moved the row range, since `Update` moved the scroll after the build.
  `endFrame` runs that `Tick` after `Update`, so `Draw` builds again in the same frame and never
  paints rows of the old position.
- `frametree_test.go` (app) runs a script of frames and asserts after each that the painted tree
  dumps the same as a fresh build. A path that fails it must set the flag; the test is never
  relaxed.

### Runtime rules

- **Events** (`view_test.go`): a left click reaches every `OnClick` under the pointer, innermost
  first, up through its ancestors; a right click only the innermost `OnRightClick`; the wheel the
  innermost scroller that uses it.
- **Clipping hides the pointer.** While the pointer is outside a clipping element (`Clip`,
  `Scroll`), its children cannot be hit; the click flag still passes, so fields can blur.
- **Modal owns the pointer.** While a `Modal` element is visible, nothing under it gets clicks,
  wheel or hover (`Ctx.HoverOff`), and a click outside runs `OnDismiss`, except for a dialog (`KeepOnOutsideClick`, set by `widget.Dialog`): it swallows the click and stays until its close button or Escape runs `Dismiss`, so a stray click on the window behind cannot throw away what the user is doing. Menus and pick lists keep closing on a click elsewhere. Toasts are not modal.
- **A modal splits the root's children.** `view.ModalLayer(ctx, root)` is the index of the first visible direct child of the root that is a `Modal` (menus included: a context menu alone gives the layer under it), or -1 when the modal is deeper, the root draws after its children (`Border`, a control style border, an `OverDrawer`, an edge) or is itself modal. `view.DrawBelow` paints the root's own part and the children before that index, `view.DrawFrom` the children from it on, both with hover off and the root's clip; one after the other on one image they paint what `Draw` paints, element for element (`El.draw` is `drawSelf`, the children, `drawAfter`). An app keeps the first part in an image (`ui/window` `baseSnapshot`). `Ctx.CaretReads` counts `CaretVisible` calls, so a kept drawing knows it shows the blink phase.
- **The runtime draws on a `render.Canvas`.** `Draw`, `DrawBelow` and `DrawFrom` take a
  `render.Canvas`; `drawSelf`, `drawAfter`, `kidsDst`
  (`Ctx.Clip`) and `drawKids` work on it, and culling compares with `kidsDst.Bounds()`. A clip
  off the canvas has empty bounds and `drawKids` draws nothing then (an empty box would otherwise
  "reach" a child near the origin). `Behavior.Draw`, `OverDrawer.DrawOver` and the text behavior
  take the same (clipped) canvas, so a behavior draws on a canvas of any renderer and text records
  `DrawText`. No file of `view` imports Ebitengine (`TestViewImportsNoEbiten`).
- **Layout writes no state** (`view.Layout` doc, `widget/layoutpure_test.go`); `Update` settles it.
- **Drawing culls** off-image children (`view.Draw` doc, `BenchmarkLongList`), so a `Behavior`
  keeps no state that only its `Draw` writes.
- **Measure once per pass**: an element is measured once per width in one `Layout` pass
  (`Ctx.LayoutGen`, `Ctx.Laying`).
- **Layout is a flexbox subset** (`flex.go`). When a row overflows, `Optional` children drop
  first, then `Shrink` children give up the rest, first to last, down to their minimum; a row that
  fits lays out as if none had it (why explicit: the `Shrink` doc).
- **Cursor.** Widgets ask for a shape in `Ctx.Cursor`, reset every frame and applied by the app.
  `Component.showGrab` asks for `core.CursorMove` while a row, tab or tree row is dragged, unless a
  resize shape was asked for.
- **Focus.** Tab cycles inside the topmost `FocusScope`; `revealFocused` scrolls every `Scroll`
  holding the newly focused field so the field is in view.

### Window work off the frame

The window's background work; the reasons are in the doc comments of the packages named.

- **Background writes** (`feature/workspace/serial`, `serial.Writer`): autosave clones what
  it writes on the UI goroutine and runs the write as a Cmd. `serial.Writer` (its `Queue` is the rule) runs one write per file
  and keeps only the newest waiting one, so an older write never lands after a newer one; the
  outcome returns on the bus. Files the user names (saved body, Postman export, screenshot) go the
  same way through `workspace.Folders.Write`. `FlushWorkspace`, `SaveState` and `openWorkspaceAt` call
  `Writer.Wait`, then write synchronously. The window does no file I/O of its own
  (`TestTheAppDoesNoFileAccessOfItsOwn`; why one port: the `workspace.Folders` doc).
- **Socket sends** (`ui/window/send.go`): messages and pings go through `Component.sendInOrder`, a per-tab `outbox.Outbox`
  (`internal/feature/send/outbox`: the queue and the in-flight flag, `Enqueue` and `Done`) run by a Cmd, so a slow network never stops a frame and messages keep their order.
- **Change keys** (`workspace/autosave.Keys`, held by `Component.keys`): autosave and the window title compare JSON forms.
  `Keys` caches them and marks them stale only on input, a bus message, a toast change or a geometry
  change. Save, load and flush use `CollectionNow`; a test that edits the model directly calls
  `keys.MarkStale`.

## Toolkit: API and templates

The package map is `doc.go`; import and naming rules are `rules.md` (Layers, Toolkit API
rules). `examples` are tested examples that use the exported API only.

- **Families stay whole.** An option family (`MinW`, `AlignStart`, `JustifyBetween`, `PadL`...) is
  exported even where the app uses part of it, since a half family breaks rule 1; helpers only the
  toolkit uses are unexported. Borders are one-pixel hairlines in a token; `EdgeT/L/R/B` are the
  separate two-pixel open-tab outline. The `Round` family is only the corner masks a panel needed
  (`RoundT`, `RoundB`, `RoundBL`; see Exceptions).
- **Mistakes fail loudly.** A component option on the wrong kind of element panics, naming the
  option and the element's control (`behaviorOf`). A template error names the file and line,
  suggests the nearest name and lists the valid ones.
- **Theme pipeline**: in `style.md`, Theming pipeline. `core`'s tests build a small `testTheme`,
  since `theming` imports `core` and never the reverse.
- **Templates** (`markup`; format and why in its package doc). The app writes each view as a
  template in a `const` at the end of its Go file, parsed once with `markup.Must`. A template has
  one root, so a view that shows one of two things is two templates and Go picks. Parse does all
  the work that does not depend on the scope (element specs, `for`, `{{path}}` splits of text and
  attributes, the options of static attributes, shared by every build; a param's value is split at
  build); a build reads only handlers, `if`, `for`, params and `{{...}}` from the scope. A build
  reads names from the caller's scope, then from fallback scopes (`Build(scope, fallbacks...)`),
  and writes into none of them; a `for` binds its variable and `$index` in a small chain, one
  binding per loop, so its cost does not grow with the size of the scope. The app builds its
  templates with `sizes.Build`, which passes the fixed widths and `Ctx.SizesScope()` as fallbacks. A path reads map keys and exported fields only: no methods,
  comparisons or filters, so anything computed is a field of a view model. A wrong scope type, a
  nil state pointer or a widget-only option on another element is a parse or build error, never a
  panic at layout. `Load`/`MustLoad` (templates in an `fs.FS`) are for other programs.
- **Slots.** `<slot name="path"/>` puts a scope `view.Node` into a template: another template's
  tree or a `view.Custom` component, whose measure, layout and input run as in a Go tree
  (`custom_test.go`).
- **Dumps pin views.** `view.Dump` prints a laid-out tree with every non-zero attribute (one
  name-and-getter table, `dumpedAttributes`). Pinned dumps keep a view from changing, and tests
  compare a template with its Go form this way.

### Exceptions to the toolkit API rules

Each is also said in the doc comment of what breaks the rule; a new exception adds a row here.

- Rule 1: `Round` is not a per-side family but the corner masks a panel needed (`RoundT`, `RoundB`,
  `RoundBL`); a mask is added with the panel that needs it, not ahead.
- Rule 7: state types are values (`NewTextField`, `NewTextArea`, `NewMenu`); `OpenMenu` stores a
  pointer to its own copy.
- Rule 11: `core.Ctx.Menu` is `any`, because `widget` owns `MenuState` and `core` cannot name it;
  that is the price of any element opening a menu without a host. `core.MaskKey.Source` is `any`
  for the same reason: it holds the pointer to the parsed icon, a `widget` type. `core.Ctx.FocusOwner` is typed,
  since its interface (`core.Focuser`, aliased by `view.Focuser`) uses no widget type.
- Exported mutable tables are copies (`theming.CornerChoices` and the other choice lists);
  `engine.keyNames` is unexported, since nothing outside `engine` reads it.
- Ports of at most about five methods (`common/architecture-style.md`): `render.Canvas` has 13.
  Each is one thing a GPU backend draws in its own way and a caller needs all of them, so a split
  would only make every caller ask for every part.

## Toolkit: widgets and drawing

- **Sideways strips.** `widget.HScroll` shows a row at natural width in a clipped box, with an
  arrow only at an end that can still scroll, the wheel, a touchpad swipe (`Input.WheelX`) and
  `KeepInView()` for one child. A press dragged a few pixels pans it, following `Ctx.PointerX`
  outside the strip. So the request strip (`ui/tabstrip`) reorders by left
  drag only while it fits (`ScrollState.Overflowing`); the middle button always reorders. The pane
  tab strips (`ScrollTabs`) use it too.
- **Tooltips.** `view.Tooltip(text)` on any element; `view.Draw` records the innermost hovered tip
  in `Ctx.TipText`, and `widget.TooltipLayer`, mounted last in the window's Stack, shows it after a
  rest. While a tip waits the app keeps drawing (`TooltipState.Waiting`). Their look is in `style.md`.
- **Icons** (look and sources: `style.md`, Icons). The SVG subset is in `svg.go`; `RegisterIcons(fs)`
  adds or replaces icons from a folder; one-color icons are white masks kept by the `core.Ctx` (`Ctx.Mask`, at most 512, dropped whole and
  released when full, and on `SetScale` and `SetTheme`) and tinted with `DrawMask`; icons with token
  paints are drawn as `render.Path`s every frame. The package keeps no GPU cache.
- **Images.** `widget.Image` takes a `render.Image` and draws it scaled to fit with
  `render.FilterLinear`; a nil picture, also a nil pointer in the interface, measures 0x0. The
  `<image img>` attribute takes a `render.Image` from the scope.
- **Sharp edges.** `Ctx.Fill`, `Outline`, `HLine` and `VLine` snap each logical edge to device
  pixels (`Ctx.snap`), so neighbours share an edge, and draw lines `Ctx.LineWidth()` thick. Rounded
  fills and outlines (`FillStyled`, `OutlineStyled`) snap the same way and are one canvas call,
  which `ebitenrender` draws from cached corner masks. `Ctx.Stats` counts those calls
  (`MaskFills`, `RingFills`) and the real measurements; `BenchmarkDrawStats` in
  `internal/ui/window` reports them.
- **Caret blink** follows the clock (`Ctx.CaretVisible`), not ticks, so its rate holds when the idle
  app drops from 60 to 30 updates. `engine.MinFPS` (10) draws at least every 100 ms, so a phase
  change shows.
- **Width cache.** `Ctx.Width` remembers a measured string by face (font source and size) and text (up to
  4096 entries, then the map is dropped); `SetScale`, `SetTheme` and `SetFonts` clear it. An
  element's own font makes a new face value each call, so the key is what the face is made of,
  not its address. `DrawStats.Widths` counts real measurements only. `Ctx.Advance` (the exact advance
  of one mono character) is cached the same way by face (`faceKey`, cleared in the same three
  places; `DrawStats.Advances` counts real measurements).
- **Sizes scope memo.** `Ctx.SizesScope` is `Sizes.Scope` made once per `Sizes` value: it holds the
  `Sizes` it was made from and is rebuilt when `Ctx.Sizes` differs, so an assignment to the exported
  field cannot leave it stale. The map is shared and read-only; `Sizes.Scope` still gives a fresh map.
- **CodeView layout cache.** The runes of every line, the digit count of the widest line number and
  the wrap layout are kept between calls
  (`codeview_cache.go`), so an Update or Draw with nothing changed costs no allocation per line (a
  40k-line body went from 6-18 ms to under 0.1 ms per Update). The layout is rebuilt when the
  lines (`SetLines`, or `Lines` reassigned), the folded lines, the wrap width or the font advance
  change. Spans or `Num` edited in place without `SetLines` are not seen. Update and Draw build
  one geometry and one content width per call and pass them down. A Draw cuts the spans of a
  wrapped row into substrings of the line's text in a buffer the view reuses, so a row costs no
  allocation of its own; what a frame still allocates is inside Ebitengine's draw queue.
- **Monospace columns.** `CodeView` and `TextArea` share `monoGrid`, which places columns by the
  exact advance so the caret, selection, matches and clicks stay on the text. A tab advances to
  the next stop of `core.TabStop` columns along its line (`core.ColumnAfter`); the wrap layout keeps
  each row's start column, so drawing, the caret, selection, matches, clicks, wrapping and the
  horizontal extent all count it the same way. The text keeps the real tab (display only).

## Modules and releases

The repo holds three modules; `go.work` (committed, for development only: consumers ignore it) joins them.

| Module | Path | Tag |
|---|---|---|
| toolkit (root) | `github.com/gascoigne-ai/uitoolkit` | `vX.Y.Z` |
| Ebitengine renderer | `github.com/gascoigne-ai/uitoolkit/renderer/ebiten` | `renderer/ebiten/vX.Y.Z` |
| examples | `github.com/gascoigne-ai/uitoolkit/examples` | none |

- **Root stands alone.** It never requires a renderer or a GPU library; `make deps-check` and the CI job `deps`
  run `go list -deps -test ./...` with `GOWORK=off` and fail on `hajimehoshi` or `wgpu`, check `go mod tidy -diff`
  and build with `CGO_ENABLED=0`. Root checks run with `GOWORK=off`; the other modules run in the workspace.
- **Release order.** Tag the root first, then bump the `require` of the root in `renderer/ebiten/go.mod` (and
  in `examples/go.mod`) to that tag, then tag the renderer.
- **Before the first tag.** The renderer and the examples require the toolkit (and the examples the renderer) at
  `v0.0.0-00010101000000-000000000000`, and `go.work` replaces exactly that version with the local directory,
  which Go needs even in workspace mode. So they build only inside the workspace, and `go mod tidy` for them
  needs a temporary `replace`. Replacing those requires with real tags and dropping the `replace` lines of
  `go.work` is part of the first release.
