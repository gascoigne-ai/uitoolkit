package core

import (
	"testing"

	"github.com/gascoigne-ai/uitoolkit/render"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/gofont/goregular"
)

// proportional reports whether the face gives "i" and "W" different
// widths, which a monospace font never does.
func proportional(ctx *Ctx, face render.Face) bool {
	return ctx.Width(face, "iiiiii") != ctx.Width(face, "WWWWWW")
}

func TestSetFontsReplacesTheInterfaceAndCodeFaces(t *testing.T) {
	ctx, err := NewCtx(testTheme(), testDevice{})
	if err != nil {
		t.Fatal(err)
	}
	if !proportional(ctx, ctx.F.UI) || proportional(ctx, ctx.F.Mono) {
		t.Fatal("the built-in fonts are not a proportional UI font and a monospace code font")
	}
	if err := ctx.SetFonts(FontData{Regular: gomono.TTF, Bold: gomonobold.TTF}, FontData{Regular: goregular.TTF}); err != nil {
		t.Fatal(err)
	}
	if proportional(ctx, ctx.F.UI) || proportional(ctx, ctx.F.Small) || proportional(ctx, ctx.F.Bold) {
		t.Fatal("the interface faces kept the old font")
	}
	if !proportional(ctx, ctx.F.Mono) {
		t.Fatal("the code face kept the old font")
	}
}

func TestSetFontsWithEmptyDataGoesBackToTheBuiltInFonts(t *testing.T) {
	ctx, _ := NewCtx(testTheme(), testDevice{})
	if err := ctx.SetFonts(FontData{Regular: gomono.TTF, Bold: gomonobold.TTF}, FontData{Regular: goregular.TTF}); err != nil {
		t.Fatal(err)
	}
	if err := ctx.SetFonts(FontData{}, FontData{}); err != nil {
		t.Fatal(err)
	}
	if !proportional(ctx, ctx.F.UI) || proportional(ctx, ctx.F.Mono) {
		t.Fatal("not back to the built-in fonts")
	}
}

func TestSetFontsKeepsTheOldFacesWhenTheDataIsNotAFont(t *testing.T) {
	ctx, _ := NewCtx(testTheme(), testDevice{})
	before := ctx.F
	if err := ctx.SetFonts(FontData{Regular: []byte("nope")}, FontData{}); err == nil {
		t.Fatal("garbage was accepted as a font")
	}
	if ctx.F != before {
		t.Fatal("the faces changed although the font was refused")
	}
}

func TestTheChosenFontsSurviveAThemeOrScaleChange(t *testing.T) {
	ctx, _ := NewCtx(testTheme(), testDevice{})
	if err := ctx.SetFonts(FontData{Regular: gomono.TTF, Bold: gomonobold.TTF}, FontData{}); err != nil {
		t.Fatal(err)
	}
	if err := ctx.SetTheme(testTheme()); err != nil {
		t.Fatal(err)
	}
	if err := ctx.SetScale(2); err != nil {
		t.Fatal(err)
	}
	if proportional(ctx, ctx.F.UI) {
		t.Fatal("the chosen font was dropped by a theme or scale change")
	}
}
