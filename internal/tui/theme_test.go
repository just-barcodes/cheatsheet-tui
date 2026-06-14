package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/just-barcodes/cheatsheet-tui/internal/config"
)

func TestApplyThemeOverridesColors(t *testing.T) {
	t.Cleanup(func() { ApplyTheme(config.Theme{}) }) // restore defaults for other tests

	ApplyTheme(config.Theme{Colors: config.ThemeColors{
		Keycap:    "#FF0000",
		Selection: "#00FF00",
	}})

	if got := kbdStyle.GetForeground(); got != lipgloss.Color("#FF0000") {
		t.Fatalf("keycap color = %v, want #FF0000", got)
	}
	if got := rowSelStyle.GetBackground(); got != lipgloss.Color("#00FF00") {
		t.Fatalf("selection background = %v, want #00FF00", got)
	}
	// An unset color keeps its default.
	if got := descStyle.GetForeground(); got != lipgloss.Color("#E5E7EB") {
		t.Fatalf("foreground = %v, want default #E5E7EB", got)
	}
}

func TestApplyThemeDefaultsWhenEmpty(t *testing.T) {
	t.Cleanup(func() { ApplyTheme(config.Theme{}) })

	ApplyTheme(config.Theme{})
	if got := kbdStyle.GetForeground(); got != lipgloss.Color("#22D3EE") {
		t.Fatalf("keycap color = %v, want default #22D3EE", got)
	}
}
