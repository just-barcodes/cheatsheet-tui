package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ThemeFileName is the reserved theme filename. The cheatsheet loaders skip it
// so a theme.yaml sitting alongside the cheatsheets is never parsed as one.
const ThemeFileName = "theme.yaml"

// presets are the built-in named themes, selectable with `--theme <name>` or a
// `name:` line in theme.yaml. To stay cohesive (the hallmark of these palettes)
// each themed preset uses one accent hue for all chrome and section headers, and
// reserves cyan for the keycaps — never a third clashing hue.
//
// selenized-* use the Selenized palette (https://github.com/jan-warchol/selenized);
// solarized-* use Ethan Schoonover's Solarized (https://ethanschoonover.com/solarized).
var presets = map[string]ThemeColors{
	// The default has no background, so it stays terminal-native (and keeps any
	// transparency). The other presets set one so they look as designed.
	"default": {
		Accent: "#A78BFA", AccentBright: "#C4B5FD", Keycap: "#22D3EE",
		Foreground: "#E5E7EB", Muted: "#6B7280", Border: "#3F3F46", Selection: "#312E81",
	},
	"selenized-dark": {
		Background: "#103c48", Accent: "#4695f7", AccentBright: "#4695f7", Keycap: "#41c7b9",
		Foreground: "#adbcbc", Muted: "#72898f", Border: "#2d5b69", Selection: "#184956",
	},
	"selenized-light": {
		Background: "#fbf3db", Accent: "#0072d4", AccentBright: "#0072d4", Keycap: "#009c8f",
		Foreground: "#3a4d53", Muted: "#909995", Border: "#d5cdb6", Selection: "#ece3cc",
	},
	"solarized-dark": {
		Background: "#002b36", Accent: "#268bd2", AccentBright: "#268bd2", Keycap: "#2aa198",
		Foreground: "#839496", Muted: "#586e75", Border: "#586e75", Selection: "#073642",
	},
	"solarized-light": {
		Background: "#fdf6e3", Accent: "#268bd2", AccentBright: "#268bd2", Keycap: "#2aa198",
		Foreground: "#657b83", Muted: "#93a1a1", Border: "#93a1a1", Selection: "#eee8d5",
	},
	"dracula": {
		Background: "#282a36", Accent: "#bd93f9", AccentBright: "#bd93f9", Keycap: "#8be9fd",
		Foreground: "#f8f8f2", Muted: "#6272a4", Border: "#44475a", Selection: "#44475a",
	},
	"nord": {
		Background: "#2e3440", Accent: "#5e81ac", AccentBright: "#5e81ac", Keycap: "#88c0d0",
		Foreground: "#d8dee9", Muted: "#4c566a", Border: "#4c566a", Selection: "#434c5e",
	},
	"gruvbox-dark": {
		Background: "#282828", Accent: "#fe8019", AccentBright: "#fe8019", Keycap: "#8ec07c",
		Foreground: "#ebdbb2", Muted: "#928374", Border: "#665c54", Selection: "#504945",
	},
	"gruvbox-light": {
		Background: "#fbf1c7", Accent: "#af3a03", AccentBright: "#af3a03", Keycap: "#427b58",
		Foreground: "#3c3836", Muted: "#928374", Border: "#bdae93", Selection: "#d5c4a1",
	},
	"tokyo-night": {
		Background: "#1a1b26", Accent: "#7aa2f7", AccentBright: "#7aa2f7", Keycap: "#7dcfff",
		Foreground: "#c0caf5", Muted: "#565f89", Border: "#3b4261", Selection: "#292e42",
	},
	"catppuccin-mocha": {
		Background: "#1e1e2e", Accent: "#cba6f7", AccentBright: "#cba6f7", Keycap: "#94e2d5",
		Foreground: "#cdd6f4", Muted: "#6c7086", Border: "#45475a", Selection: "#313244",
	},
}

// Preset returns a built-in theme by name.
func Preset(name string) (Theme, bool) {
	c, ok := presets[name]
	if !ok {
		return Theme{}, false
	}
	return Theme{Colors: c}, true
}

// IsPreset reports whether name is a built-in theme.
func IsPreset(name string) bool {
	_, ok := presets[name]
	return ok
}

// PresetNames lists the built-in themes, sorted.
func PresetNames() []string {
	names := make([]string, 0, len(presets))
	for n := range presets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ThemeSource describes where the active theme comes from.
type ThemeSource struct {
	Preset    string // non-empty: use this built-in preset
	Path      string // theme file to load; empty means "no file, use defaults"
	MustExist bool   // an explicit --theme path must exist; the default may be absent
}

// ThemeLocator decides where the theme comes from. Inputs are injected so the
// decision is pure and testable; main wires in the real flag and config dir.
type ThemeLocator struct {
	Flag      string // value of the --theme flag ("" if unset)
	ConfigDir string // e.g. ~/.config/cheatsheet ("" if unavailable)
}

// Resolve picks the theme source: an explicit --theme wins — a built-in name
// selects that preset, anything else is a file path that must exist; otherwise
// theme.yaml in the config dir is used if present; otherwise no theme.
func (l ThemeLocator) Resolve() ThemeSource {
	switch {
	case l.Flag != "":
		if IsPreset(l.Flag) {
			return ThemeSource{Preset: l.Flag}
		}
		return ThemeSource{Path: l.Flag, MustExist: true}
	case l.ConfigDir != "":
		return ThemeSource{Path: filepath.Join(l.ConfigDir, ThemeFileName)}
	default:
		return ThemeSource{}
	}
}

// Theme overrides the UI color palette. Every field is optional; an empty value
// keeps the built-in default. Colors are either a hex string like "#A78BFA" or
// a 0–255 terminal color number like "63".
type Theme struct {
	Name   string      `yaml:"name"`   // optional built-in preset to start from
	Colors ThemeColors `yaml:"colors"` // overrides applied on top of the preset/defaults
}

// ThemeColors names each colour the UI is built from.
type ThemeColors struct {
	Background   string `yaml:"background"`    // the UI background; empty keeps the terminal's own
	Accent       string `yaml:"accent"`        // headings, active border, search prompt
	AccentBright string `yaml:"accent_bright"` // section titles, footer keys
	Keycap       string `yaml:"keycap"`        // the hotkeys themselves
	Foreground   string `yaml:"foreground"`    // descriptions and body text
	Muted        string `yaml:"muted"`         // hints, counts, inactive text
	Border       string `yaml:"border"`        // inactive pane borders, scrollbar track
	Selection    string `yaml:"selection"`     // highlighted row background
}

// LoadTheme parses a theme from r. An empty document yields the zero Theme,
// which the UI reads as "use every default".
func LoadTheme(r io.Reader) (Theme, error) {
	var t Theme
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&t); err != nil {
		if err == io.EOF {
			return Theme{}, nil
		}
		return Theme{}, err
	}
	if t.Name != "" {
		base, ok := Preset(t.Name)
		if !ok {
			return Theme{}, fmt.Errorf("unknown theme %q (available: %s)", t.Name, strings.Join(PresetNames(), ", "))
		}
		t.Colors = base.Colors.overlay(t.Colors)
		t.Name = ""
	}
	if err := t.validate(); err != nil {
		return Theme{}, err
	}
	return t, nil
}

// overlay returns base with every non-empty field of o laid on top.
func (base ThemeColors) overlay(o ThemeColors) ThemeColors {
	pick := func(over, def string) string {
		if over != "" {
			return over
		}
		return def
	}
	return ThemeColors{
		Background:   pick(o.Background, base.Background),
		Accent:       pick(o.Accent, base.Accent),
		AccentBright: pick(o.AccentBright, base.AccentBright),
		Keycap:       pick(o.Keycap, base.Keycap),
		Foreground:   pick(o.Foreground, base.Foreground),
		Muted:        pick(o.Muted, base.Muted),
		Border:       pick(o.Border, base.Border),
		Selection:    pick(o.Selection, base.Selection),
	}
}

// LoadThemeFile parses theme.yaml at path. A missing file is not an error — it
// simply means the defaults apply.
func LoadThemeFile(path string) (Theme, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Theme{}, nil
		}
		return Theme{}, err
	}
	defer f.Close()
	t, err := LoadTheme(f)
	if err != nil {
		return Theme{}, fmt.Errorf("%s: %w", path, err)
	}
	return t, nil
}

var hexColor = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// validate rejects malformed colours so a typo surfaces as an error instead of
// silently rendering as no colour.
func (t Theme) validate() error {
	fields := map[string]string{
		"background":    t.Colors.Background,
		"accent":        t.Colors.Accent,
		"accent_bright": t.Colors.AccentBright,
		"keycap":        t.Colors.Keycap,
		"foreground":    t.Colors.Foreground,
		"muted":         t.Colors.Muted,
		"border":        t.Colors.Border,
		"selection":     t.Colors.Selection,
	}
	for name, val := range fields {
		if val == "" || hexColor.MatchString(val) {
			continue
		}
		if n, err := strconv.Atoi(val); err == nil && n >= 0 && n <= 255 {
			continue
		}
		return fmt.Errorf("colors.%s: %q is not a hex color (#RGB/#RRGGBB) or a 0–255 number", name, val)
	}
	return nil
}
