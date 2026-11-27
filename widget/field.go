package widget

import (
	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/render"
	. "github.com/gascoigne-ai/uitoolkit/view"
)

// FieldPad sets a field's inner padding, for fields inset in a box that
// already has padding of its own, so text lines up with its neighbours.
func FieldPad(pixels int) Option {
	return func(element *El) {
		field := behaviorOf[*fieldB](element, "FieldPad")
		field.inset, field.insetOf = pixels, nil
	}
}

// fieldPadOf is FieldPad for a padding that follows the context's Sizes. It
// may be 0, unlike FieldPad, where 0 is the default.
func fieldPadOf(pad func(sizes core.Sizes) int) Option {
	return func(element *El) { behaviorOf[*fieldB](element, "FieldPad").insetOf = pad }
}

// fieldB is a single-line text input editing a TextField.
type fieldB struct {
	Base
	field       *TextField
	placeholder string
	segments    Styler
	onEnter     func()
	onTab       func(back bool)
	inset       int                        // FieldPad, 0 for the default
	insetOf     func(sizes core.Sizes) int // fieldPadOf; wins over inset while set
}

// textInset is the space between a field's edge and its text: the themed
// Sizes.FieldPadX, or what FieldPad or fieldPadOf set.
func (b *fieldB) textInset(sizes core.Sizes) int {
	if b.insetOf != nil {
		return b.insetOf(sizes)
	}
	if b.inset > 0 {
		return b.inset
	}
	return sizes.FieldPadX
}

func (b *fieldB) FocusTarget() Focuser { return b.field }

// KeepsTab: a field with OnTab uses Tab itself.
func (b *fieldB) KeepsTab() bool { return b.onTab != nil }

func (b *fieldB) Update(element *El, ctx *core.Ctx) {
	f, in, r := b.field, ctx.In, element.Rect()
	at := func(posX int) int { return ctx.CursorAt(f.Display(), posX-(r.X+b.textInset(ctx.Sizes))+f.ScrollX) }
	if in.Click {
		f.Focused = r.Contains(in.X, in.Y)
		if f.Focused {
			f.PressN(at(in.X), in.Shift, max(in.Clicks, 1))
		}
	}
	if f.Dragging {
		if in.Held {
			f.Drag(at(ctx.PointerX))
		} else {
			f.Dragging = false
		}
	}
	if !f.Focused {
		return
	}
	if in.Escape {
		f.HasSel = false // focus stays: the window's Escape steps decide what else closes
	}
	changed := f.RunShortcuts(ctx)
	typed := in.Typed
	if in.Ctrl {
		typed = nil // Ctrl+letter is a shortcut, not text
	}
	if f.Apply(typed, in.Edits, in.Shift) || changed {
		ctx.Fired++
		element.Changed()
	}
	if in.Enter && !in.Ctrl && b.onEnter != nil { // Ctrl+Enter is a shortcut of the app
		ctx.Fired++
		b.onEnter()
	}
	if in.Tab && b.onTab != nil {
		ctx.Fired++
		b.onTab(in.Shift)
	}
	f.Cursor = f.clamp(f.Cursor) // OnChange may have replaced the text
	f.Reveal(ctx.Width(ctx.F.UI, string([]rune(f.Display())[:f.Cursor])), r.W-2*b.textInset(ctx.Sizes), ctx.Sizes.RevealMargin)
}

func (b *fieldB) Draw(element *El, ctx *core.Ctx, dst render.Canvas) {
	f, r := b.field, element.Rect()
	dst = ctx.Clip(dst, r)
	x := r.X + b.textInset(ctx.Sizes) - f.ScrollX
	if lo, hi, ok := f.Selection(); ok {
		runes := []rune(f.Display())
		x0 := x + ctx.Width(ctx.F.UI, string(runes[:lo]))
		x1 := x + ctx.Width(ctx.F.UI, string(runes[:hi]))
		inset := ctx.Sizes.SelectionInset
		ctx.Fill(dst, core.Rect{X: x0, Y: r.Y + inset, W: x1 - x0, H: max(r.H-2*inset, 0)}, ctx.Color(core.Selection))
	}
	segs := []Span{{Text: f.Display(), Style: core.Text}}
	if b.segments != nil && !f.Masked {
		segs = b.segments(f.Text)
	}
	for _, s := range segs {
		ctx.TextMid(dst, ctx.F.UI, s.Text, x, r, ctx.Color(s.Style))
		x += ctx.Width(ctx.F.UI, s.Text)
	}
	if f.Text == "" {
		ctx.TextMid(dst, ctx.F.UI, b.placeholder, r.X+b.textInset(ctx.Sizes), r, ctx.Color(core.Muted))
	}
	if f.Focused && ctx.CaretVisible() {
		shown, cursor := []rune(f.Display()), f.clamp(f.Cursor)
		cx := r.X + b.textInset(ctx.Sizes) - f.ScrollX + ctx.Width(ctx.F.UI, string(shown[:cursor]))
		drawCaret(ctx, dst, ctx.F.UI, caretGlyph(shown, cursor), cx, r)
	}
}

// DrawOver is the focus ring, over the field's outline.
func (b *fieldB) DrawOver(element *El, ctx *core.Ctx, dst render.Canvas) {
	if b.field.Focused {
		ctx.OutlineStyled(dst, element.Rect(), ctx.Color(core.AccentLine), ctx.Th.Style(element.ControlName()).Radius)
	}
}

// Field is a single-line text input bound to f (Vue's v-model: the
// widget edits f in place).
func Field(textField *TextField, opts ...Option) *El {
	return Custom(&fieldB{field: textField}, With([]Item{Control("field")}, opts)...)
}

// Placeholder is the hint shown while the field is empty.
func Placeholder(text string) Option {
	return func(element *El) { behaviorOf[*fieldB](element, "Placeholder").placeholder = text }
}

// Styled colors parts of the field's text with styler, which gets the
// whole text.
func Styled(styler Styler) Option {
	return func(element *El) { behaviorOf[*fieldB](element, "Styled").segments = styler }
}

// OnTab runs when Tab is pressed in the focused field; back is true with Shift.
func OnTab(callback func(back bool)) Option {
	return func(element *El) { behaviorOf[*fieldB](element, "OnTab").onTab = callback }
}

// OnEnter runs when Enter is pressed in the focused field.
func OnEnter(callback func()) Option {
	return func(element *El) { behaviorOf[*fieldB](element, "OnEnter").onEnter = callback }
}
