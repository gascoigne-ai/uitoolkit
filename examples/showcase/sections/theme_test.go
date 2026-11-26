package sections

import (
	"strings"
	"testing"

	"github.com/gascoigne-ai/uitoolkit/core"
)

func TestLiveThemeEdit(t *testing.T) {
	showcase, ctx, _ := newTestShowcase(t)
	open(t, showcase, ctx, "theming")
	frame(t, showcase, ctx)
	showcase.ThemeJSON.Text = `{"colors": {"bg": "#102030"}}`
	showcase.ApplyThemeJSON()
	frame(t, showcase, ctx)
	edited, err := core.ParseHex("#102030")
	if err != nil {
		t.Fatal(err)
	}
	if showcase.ThemeErr != "" || ctx.Color(core.Bg) != edited {
		t.Fatalf("after a valid edit: error %q, bg %v, want %v", showcase.ThemeErr, ctx.Color(core.Bg), edited)
	}
	tests := []struct{ name, text string }{
		{"invalid color", `{"colors": {"bg": "nope"}}`},
		{"unknown field", `{"colour": {}}`},
		{"cut off", `{"colors": {`},
		{"empty", ""},
		{"too large", `{"name": "` + strings.Repeat("x", maxThemeJSON) + `"}`},
		{"deeply nested", `{"colors": ` + strings.Repeat("[", 30000) + strings.Repeat("]", 30000) + `}`},
		{"deeply nested gradient", `{"controls": {"a": {"gradient": ` + strings.Repeat("[", 30000) + strings.Repeat("]", 30000) + `}}}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			showcase.ThemeJSON.Text = test.text
			showcase.ApplyThemeJSON()
			frame(t, showcase, ctx)
			if showcase.ThemeErr == "" {
				t.Error("no error for a bad theme")
			}
			if ctx.Color(core.Bg) != edited || showcase.Theme().Color(core.Bg) != edited {
				t.Errorf("bg = %v, want the last good %v", ctx.Color(core.Bg), edited)
			}
		})
	}
}

func TestLiveThemeEditShowsTheErrorUnderTheEditor(t *testing.T) {
	for form := range []int{0, 1} {
		showcase, ctx, _ := newTestShowcase(t)
		open(t, showcase, ctx, "theming")
		showcase.Form = form
		showcase.ThemeJSON.Text = `{"colour": 1}`
		showcase.ApplyThemeJSON()
		dump := layoutDump(ctx, showcase.SectionView(ctx))
		editor, message := strings.Index(dump, "areaB"), strings.Index(dump, `unknown field`)
		if editor < 0 || message < editor {
			t.Errorf("form %d: the error is not under the editor:\n%s", form, dump)
		}
	}
}

// Typing in the editor applies the edit, in both forms.
func TestTypingInTheThemeEditorAppliesIt(t *testing.T) {
	for form := range []int{0, 1} {
		showcase, ctx, _ := newTestShowcase(t)
		open(t, showcase, ctx, "theming")
		showcase.Form = form
		frame(t, showcase, ctx)
		// At the start: theming.ParseTheme ignores whatever follows the first
		// JSON value, so a character typed after it is not an error.
		showcase.ThemeJSON.Focused, showcase.ThemeJSON.Cursor = true, 0
		ctx.In = core.Input{Typed: []rune("x")}
		frame(t, showcase, ctx)
		if showcase.ThemeErr == "" {
			t.Errorf("form %d: a typed character that breaks the JSON set no error", form)
		}
	}
}

func TestThemeEditStaysOnTheNewBaseWhenThePreferencesChange(t *testing.T) {
	showcase, _, _ := newTestShowcase(t)
	showcase.ThemeJSON.Text = `{"colors": {"bg": "#102030"}}`
	showcase.ApplyThemeJSON()
	showcase.SetThemeID("light")
	edited, _ := core.ParseHex("#102030")
	light := New(newClock().now)
	light.SetThemeID("light")
	if showcase.Theme().Color(core.Bg) != edited || showcase.Theme().Color(core.Text) != light.Theme().Color(core.Text) {
		t.Errorf("bg %v text %v, want the edit over the light theme", showcase.Theme().Color(core.Bg), showcase.Theme().Color(core.Text))
	}
	showcase.ThemeJSON.Text = "{"
	showcase.ApplyThemeJSON()
	showcase.SetThemeID("dark")
	if showcase.ThemeErr == "" || showcase.Theme().Color(core.Bg) != New(newClock().now).Theme().Color(core.Bg) {
		t.Errorf("with a bad edit a new theme shows plain and keeps the error: %q, bg %v", showcase.ThemeErr, showcase.Theme().Color(core.Bg))
	}
}
