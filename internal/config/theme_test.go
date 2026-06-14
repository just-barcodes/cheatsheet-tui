package config

import "testing"

// TestPresetsAreValidAndCohesive guards every built-in theme: colors must parse,
// the core roles must be set, and themed presets must keep a single accent hue
// (section headers/footer keys share the chrome accent) over a real background.
func TestPresetsAreValidAndCohesive(t *testing.T) {
	for name, c := range presets {
		th := Theme{Colors: c}
		if err := th.validate(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if c.Accent == "" || c.Keycap == "" || c.Foreground == "" {
			t.Errorf("%s: missing a core color", name)
		}
		if name == "default" {
			continue // the terminal-native default is exempt from the rules below
		}
		if c.AccentBright != c.Accent {
			t.Errorf("%s: accent_bright %q != accent %q (clashing hue)", name, c.AccentBright, c.Accent)
		}
		if c.Background == "" {
			t.Errorf("%s: themed preset has no background", name)
		}
	}
}
