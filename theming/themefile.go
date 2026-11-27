package theming

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gascoigne-ai/uitoolkit/core"
	"image/color"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"
)

// core.Theme files are JSON in the style of editor themes: named variables
// that colors refer to as "$name", a flat token-to-color table, and a
// table of per-control styles (corner radius, gradient, outline).
//
//	{
//	  "name": "dark",
//	  "variables": { "rose": "#e88aa6" },
//	  "colors":    { "accent": "$rose", "shadow": "#00000066" },
//	  "controls":  { "button": { "radius": 6, "gradient": 0.08, "border": "border" },
//	                 "hero":   { "gradient": ["#4f1428", "#2a0d18"] } }
//	}
//
// A file may leave out whatever a base theme already defines; colors and
// controls are merged over it.

// maxIconStroke is the largest iconStroke a theme file may state. Thin icons
// are 0.8; above four times the drawn width the strokes of an icon merge into
// a blob, and a huge value paints the icon solid or gives +Inf vertices.
const maxIconStroke = 4

type controlFile struct {
	Radius   float64         `json:"radius"`
	Gradient json.RawMessage `json:"gradient"`
	Border   string          `json:"border"`
}

type themeFile struct {
	Name    string `json:"name"`
	Extends string `json:"extends"`
	// FontSize, SmallSize and MonoSize are read and ignored: the font size
	// settings decide the sizes (FontSizes.Resolved), and files written when a theme
	// stated them keep loading because the parser rejects unknown keys.
	FontSize  float64 `json:"fontSize"`
	SmallSize float64 `json:"smallSize"`
	MonoSize  float64 `json:"monoSize"`
	// IconStroke multiplies the stroke width of every icon: 0.8 for thin
	// icons, 1.3 for bold ones; absent or 0 keeps the base theme's. The most
	// is maxIconStroke.
	IconStroke float64                `json:"iconStroke"`
	Variables  map[string]string      `json:"variables"`
	Colors     map[string]string      `json:"colors"`
	Controls   map[string]controlFile `json:"controls"`
	// Metrics are the size rules, see core.Metrics. The block is optional.
	Metrics *metricsFile `json:"metrics"`
	// Feel are the timings and pointer distances, see core.Feel. The block is
	// optional.
	Feel feelFile `json:"feel"`
}

// themeExtends is the name of the built-in theme a file says it builds on
// ("extends"), "" when it does not say. The caller picks the base theme
// and passes it to ParseTheme.
func themeExtends(data []byte) (string, error) {
	var f struct {
		Extends string `json:"extends"`
	}
	if err := json.Unmarshal(data, &f); err != nil {
		return "", fmt.Errorf("theme file: %w", err)
	}
	return f.Extends, nil
}

// ParseTheme reads a theme file over base (nil for none). The text is one
// JSON object, with nothing but white space around it: another top-level
// value, null included, and any data after the object are errors. Unknown
// fields are errors, so a typo in a theme file is not silently ignored. An
// iconStroke outside 0 and maxIconStroke is an error, as is a value of the
// metrics or feel blocks outside its bounds. Over a base, the keys of those
// blocks merge one by one and the file wins.
//
// A base keeps its RadiusScale, 0 included: it is the user's corner choice,
// and 0 is square corners, so reading an edit must not round them again.
// Without a base the scale is 1.
func ParseTheme(data []byte, base *core.Theme) (core.Theme, error) {
	// Decoding null into a struct is no error, so the object is checked first.
	if !bytes.HasPrefix(bytes.TrimLeft(data, " \t\r\n"), []byte("{")) {
		return core.Theme{}, errors.New("theme file: not a JSON object")
	}
	var f themeFile
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return core.Theme{}, fmt.Errorf("theme file: %w", err)
	}
	// Decode reads one value; only the end of the text may follow it.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return core.Theme{}, errors.New("theme file: data after the theme")
	}
	if f.IconStroke < 0 || f.IconStroke > maxIconStroke {
		return core.Theme{}, fmt.Errorf("theme file: iconStroke must be within 0 and %d", maxIconStroke)
	}
	var th core.Theme
	if base != nil {
		th = base.With(nil)
		th.Controls = copyControls(base.Controls)
	} else {
		th = core.Theme{Colors: map[core.Token]color.RGBA{}, Controls: map[string]core.ControlStyle{}, RadiusScale: 1}.WithTextSize(FontSizes{}.Resolved())
	}
	if th.Controls == nil {
		th.Controls = map[string]core.ControlStyle{}
	}
	if f.Name != "" {
		th.Name = f.Name
	}
	if f.FontSize != 0 || f.SmallSize != 0 || f.MonoSize != 0 {
		slog.Debug("theme file states text sizes; ignored, the font size settings decide", "theme", f.Name)
	}
	if f.IconStroke > 0 {
		th.IconStroke = f.IconStroke
	}

	if f.Metrics != nil {
		metrics, err := f.Metrics.apply(th.Metrics)
		if err != nil {
			return core.Theme{}, fmt.Errorf("theme file: metrics: %w", err)
		}
		th.Metrics = metrics
	}
	feel, err := f.Feel.apply(th.Feel)
	if err != nil {
		return core.Theme{}, fmt.Errorf("theme file: feel: %w", err)
	}
	th.Feel = feel

	vars := map[string]color.RGBA{}
	for _, name := range sortedKeys(f.Variables) {
		c, err := resolveColor(f.Variables[name], vars)
		if err != nil {
			return core.Theme{}, fmt.Errorf("theme file: variable %q: %w", name, err)
		}
		vars[name] = c
	}
	for _, name := range sortedKeys(f.Colors) {
		c, err := resolveColor(f.Colors[name], vars)
		if err != nil {
			return core.Theme{}, fmt.Errorf("theme file: color %q: %w", name, err)
		}
		th.Colors[core.Token(name)] = c
	}
	for _, name := range sortedControlKeys(f.Controls) {
		cf := f.Controls[name]
		st := core.ControlStyle{Radius: cf.Radius, Border: core.Token(cf.Border)}
		if st.Radius < 0 {
			return core.Theme{}, fmt.Errorf("theme file: control %q: negative radius", name)
		}
		if err := parseGradient(cf.Gradient, vars, &st); err != nil {
			return core.Theme{}, fmt.Errorf("theme file: control %q: %w", name, err)
		}
		if st.Border != "" {
			if _, ok := th.Colors[st.Border]; !ok {
				return core.Theme{}, fmt.Errorf("theme file: control %q: border token %q has no color", name, st.Border)
			}
		}
		th.Controls[name] = st
	}
	return th, nil
}

// resolveColor is a "$variable" or a literal color.
func resolveColor(text string, vars map[string]color.RGBA) (color.RGBA, error) {
	if name, ok := strings.CutPrefix(text, "$"); ok {
		c, ok := vars[name]
		if !ok {
			return color.RGBA{}, fmt.Errorf("unknown variable $%s", name)
		}
		return c, nil
	}
	return core.ParseHexAlpha(text)
}

// parseGradient reads a control's gradient: absent or 0 for none, a
// number for a shade strength, or [top, bottom] colors.
func parseGradient(raw json.RawMessage, vars map[string]color.RGBA, st *core.ControlStyle) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var strength float64
	if err := json.Unmarshal(raw, &strength); err == nil {
		if strength < 0 || strength > 0.5 {
			return errors.New("gradient strength must be within 0 and 0.5")
		}
		st.Shade = strength
		return nil
	}
	var ends []string
	if err := json.Unmarshal(raw, &ends); err != nil || len(ends) != 2 {
		return errors.New("gradient must be a number or [top, bottom]")
	}
	top, err := resolveColor(ends[0], vars)
	if err != nil {
		return err
	}
	bottom, err := resolveColor(ends[1], vars)
	if err != nil {
		return err
	}
	st.Top, st.Bottom = &top, &bottom
	return nil
}

func copyControls(other map[string]core.ControlStyle) map[string]core.ControlStyle {
	out := make(map[string]core.ControlStyle, len(other))
	maps.Copy(out, other)
	return out
}

func sortedKeys(other map[string]string) []string {
	keys := slices.Sorted(maps.Keys(other))
	return keys
}

func sortedControlKeys(other map[string]controlFile) []string {
	keys := slices.Sorted(maps.Keys(other))
	return keys
}

//go:embed themes/*.json
var builtinThemes embed.FS

func builtin(name string) core.Theme {
	data, err := builtinThemes.ReadFile("themes/" + name + ".json")
	if err != nil {
		panic(err)
	}
	th, err := ParseTheme(data, nil)
	if err != nil {
		panic(fmt.Sprintf("built-in theme %s: %v", name, err))
	}
	return th
}

// Dark is warm near-black surfaces with a rose accent. The palette and
// control styles live in themes/dark.json.
func Dark() core.Theme { return builtin("dark") }

// Light is cream surfaces, white lists and a burgundy accent
// (themes/light.json).
func Light() core.Theme { return builtin("light") }
