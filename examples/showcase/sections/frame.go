package sections

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	"github.com/gascoigne-ai/uitoolkit/theming"
	. "github.com/gascoigne-ai/uitoolkit/view"
	. "github.com/gascoigne-ai/uitoolkit/widget"
)

// Fixed widths of the frame, one per role.
const (
	navW       = 180
	overlayW   = 160
	overlayTop = 44 // under the top bar
)

// Update runs one tick in a window of width by height logical pixels: the
// keys of the frame itself (F3, Escape, the popup menu's), then the view's
// input. ctx.In and ctx.Tick are set by the caller.
func (s *Showcase) Update(ctx *core.Ctx, width, height int) error {
	if err := s.syncTheme(ctx); err != nil {
		return err
	}
	ctx.Cursor = core.CursorDefault
	s.widgets.toasts.Prune(s.now())
	if ctx.In.Pressed(render.KeyF3) {
		s.Overlay = !s.Overlay
	}
	if ctx.In.Escape {
		s.escape(ctx)
	}
	PopupMenuKeys(ctx, ctx.In)
	root := s.View(ctx, width, height)
	Layout(ctx, root, width, height)
	Update(ctx, root)
	return s.syncTheme(ctx) // a handler may have changed the theme
}

// escape closes what is on top: the popup menu, else the dialog.
func (s *Showcase) escape(ctx *core.Ctx) {
	if MenuOpen(ctx) {
		CloseMenu(ctx)
		return
	}
	s.dismissDialog()
}

// Draw paints one frame of width by height logical pixels on canvas.
func (s *Showcase) Draw(ctx *core.Ctx, width, height int, canvas render.Canvas) {
	s.beginFrame(ctx)
	root := s.View(ctx, width, height)
	Layout(ctx, root, width, height)
	Draw(ctx, root, canvas)
}

// beginFrame keeps what the overlay shows: the draw counts since the last
// frame began, which it zeroes, and the time since then.
func (s *Showcase) beginFrame(ctx *core.Ctx) {
	now := s.now()
	if !s.frameStart.IsZero() {
		s.frameTime = now.Sub(s.frameStart)
	}
	s.frameStart = now
	if ctx.Stats == nil {
		return
	}
	s.stats, *ctx.Stats = *ctx.Stats, core.DrawStats{}
}

// View is the whole window: the top bar over the nav and the chosen section,
// then the layers over them, in a window of width by height logical pixels.
func (s *Showcase) View(ctx *core.Ctx, width, height int) Node {
	return Stack(
		Col(Bg(core.Bg),
			s.topBar(),
			Row(Grow(1),
				s.Nav(ctx),
				Col(Grow(1), Clip(), s.SectionView(ctx)),
			),
		),
		If(s.Overlay, s.statsPanel()),
		s.layers(ctx, width, height),
		TooltipLayer(&s.widgets.tooltip),
	)
}

// SectionView is the chosen section in the chosen form.
func (s *Showcase) SectionView(ctx *core.Ctx) Node {
	return Forms[s.Form].section(Sections[s.Section])(s, ctx)
}

// layers are the dialog, the toasts and the popup menu in the chosen form,
// each only while there is one. The toasts place themselves in a window of
// width by height, so the layers must be direct children of its Stack.
func (s *Showcase) layers(ctx *core.Ctx, width, height int) Node {
	form := Forms[s.Form]
	var dialog, toasts, menu Node
	if s.dialog != noDialog {
		dialog = form.dialog(dialogViews[s.dialog])(s, ctx)
	}
	if len(s.widgets.toasts) > 0 {
		toasts = form.toasts(s, ctx, width, height)
	}
	if MenuOpen(ctx) {
		menu = form.menu(s, ctx)
	}
	return Frag(dialog, toasts, menu)
}

// Nav lists the sections down the left side.
func (s *Showcase) Nav(ctx *core.Ctx) Node {
	rows := For(Sections, func(index int, section Section) Node {
		return ListRow(index == s.Section, HOf(core.RowOf), PadX(core.Gutter), AlignItems(AlignCenter),
			OnClick(func() { s.Select(ctx, index) }),
			Text(section.Label, Grow(1), Ellipsis()),
		)
	})
	return Col(W(navW), Bg(core.Sunken), BorderR(core.Border), PadY(core.LooseGap), Clip(), rows)
}

// topBar holds the controls of the whole window. It scrolls sideways, so no
// control is out of reach in a narrow window.
func (s *Showcase) topBar() Node {
	return HScroll(&s.widgets.topBar, HOf(core.ToolbarOf), Bg(core.Frame), BorderB(core.Border),
		PadX(core.Gutter), Gap(core.TightGap), AlignItems(AlignCenter),
		Caption("Theme"),
		For(s.catalog.All(), func(_ int, choice theming.ThemeChoice) Node {
			return barChip(choice.Label, s.Prefs.Theme == choice.ID, func() { s.SetThemeID(choice.ID) })
		}),
		Caption("Contrast", MarginL(core.Gutter)),
		For(theming.ContrastChoices(), func(_ int, choice theming.ContrastChoice) Node {
			return barChip(choice.Label, s.Prefs.Contrast == choice.ID, func() { s.SetContrast(choice.ID) })
		}),
		Caption("Scale", MarginL(core.Gutter)),
		For(ScaleSteps, func(index int, step float64) Node {
			return barChip(percent(step), s.Scale == index, func() { s.SetScale(index) })
		}),
		Caption("Form", MarginL(core.Gutter)),
		For(Forms, func(index int, form Form) Node {
			return barChip(form.Label, s.Form == index, func() { s.SetForm(index) })
		}),
	)
}

func barChip(label string, selected bool, pick func()) Node {
	return SegmentChip(label, selected, HOf(core.CompactOf), OnClick(pick))
}

func percent(step float64) string {
	return strconv.FormatFloat(step*100, 'f', -1, 64) + "%"
}

// statsPanel is the overlay in the top right corner.
func (s *Showcase) statsPanel() Node {
	lines := []string{
		fmt.Sprintf("frame %.1f ms", float64(s.frameTime)/float64(time.Millisecond)),
		fmt.Sprintf("elements %d", s.stats.Elements),
		fmt.Sprintf("mask fills %d", s.stats.MaskFills),
		fmt.Sprintf("ring fills %d", s.stats.RingFills),
		fmt.Sprintf("widths %d", s.stats.Widths),
		fmt.Sprintf("advances %d", s.stats.Advances),
	}
	return Col(FloatTR(overlayTop, core.Gutter), W(overlayW), Pad(core.LooseGap), Bg(core.Panel), Control("menu"), Shadow(),
		For(lines, func(_ int, line string) Node { return Text(line, Mono(), HOf(core.NoteOf)) }),
	)
}
