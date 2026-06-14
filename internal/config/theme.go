package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

// ThemeFileName is the reserved theme filename. The cheatsheet loaders skip it
// so a theme.yaml sitting alongside the cheatsheets is never parsed as one.
const ThemeFileName = "theme.yaml"

// ThemeSource describes which theme file to read, if any.
type ThemeSource struct {
	Path      string // theme file to load; empty means "no file, use defaults"
	MustExist bool   // an explicit --theme path must exist; the default may be absent
}

// ThemeLocator decides where the theme is read from. Inputs are injected so the
// decision is pure and testable; main wires in the real flag and config dir.
type ThemeLocator struct {
	Flag      string // value of the --theme flag ("" if unset)
	ConfigDir string // e.g. ~/.config/cheatsheet ("" if unavailable)
}

// Resolve picks the theme source: an explicit --theme path wins and must exist;
// otherwise theme.yaml in the config dir is used if present; otherwise no file.
func (l ThemeLocator) Resolve() ThemeSource {
	switch {
	case l.Flag != "":
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
	Colors ThemeColors `yaml:"colors"`
}

// ThemeColors names each colour the UI is built from.
type ThemeColors struct {
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
	if err := t.validate(); err != nil {
		return Theme{}, err
	}
	return t, nil
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
