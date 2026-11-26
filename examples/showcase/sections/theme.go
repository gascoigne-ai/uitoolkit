package sections

import (
	"fmt"

	"github.com/gascoigne-ai/uitoolkit/core"
	"github.com/gascoigne-ai/uitoolkit/theming"
)

// maxThemeJSON bounds the theme text that is parsed: the editor's content is
// untrusted (a paste can be large) and is read on every edit. The built-in
// theme files are under 3 KiB.
const maxThemeJSON = 64 << 10

// startThemeJSON is the editor's first text: a theme file that changes
// nothing.
const startThemeJSON = `{
  "colors": {}
}`

// Theme is the theme in effect.
func (s *Showcase) Theme() core.Theme { return s.lastGood }

// SetThemeID picks a theme of the catalog.
func (s *Showcase) SetThemeID(id string) {
	s.Prefs.Theme = s.catalog.ThemeID(id)
	s.retheme()
}

// SetContrast picks a contrast level of theming.ContrastChoices.
func (s *Showcase) SetContrast(id string) {
	s.Prefs.Contrast = id
	s.retheme()
}

// retheme is the theme for new preferences. An edit that cannot be read is
// left out, since the theme it was last good on is no longer chosen.
func (s *Showcase) retheme() {
	s.lastGood = s.catalog.Theme(s.Prefs)
	s.ApplyThemeJSON()
	s.themeChanged = true
}

// ApplyThemeJSON reads ThemeJSON over the chosen theme. A text that cannot
// be read sets ThemeErr and keeps the theme in effect. The edit is laid
// over the theme after the contrast preference, so what the user types wins,
// high contrast included.
func (s *Showcase) ApplyThemeJSON() {
	theme, err := parseThemeEdit(s.ThemeJSON.Text, s.catalog.Theme(s.Prefs))
	if err != nil {
		s.ThemeErr = err.Error()
		return
	}
	s.lastGood, s.ThemeErr, s.themeChanged = theme, "", true
}

func parseThemeEdit(text string, base core.Theme) (core.Theme, error) {
	if len(text) > maxThemeJSON {
		return core.Theme{}, fmt.Errorf("theme file: %d bytes, the most is %d", len(text), maxThemeJSON)
	}
	return theming.ParseTheme([]byte(text), &base)
}

// syncTheme puts a changed theme on the context, which loads faces anew.
func (s *Showcase) syncTheme(ctx *core.Ctx) error {
	if !s.themeChanged {
		return nil
	}
	if err := ctx.SetTheme(s.lastGood); err != nil {
		return fmt.Errorf("set theme: %w", err)
	}
	s.themeChanged = false
	return nil
}
