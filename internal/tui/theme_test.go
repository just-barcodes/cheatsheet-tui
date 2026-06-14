package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

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

func TestSelectedTextContrastsWithSelection(t *testing.T) {
	t.Cleanup(func() { ApplyTheme(config.Theme{}) })

	// A light selection background (Selenized light) needs dark text.
	light, _ := config.Preset("selenized-light")
	ApplyTheme(light)
	if got := rowSelStyle.GetForeground(); got != lipgloss.Color("#0B0B12") {
		t.Fatalf("light selection should use dark text, got %v", got)
	}

	// A dark selection background (Selenized dark) needs light text.
	dark, _ := config.Preset("selenized-dark")
	ApplyTheme(dark)
	if got := rowSelStyle.GetForeground(); got != lipgloss.Color("#FFFFFF") {
		t.Fatalf("dark selection should use light text, got %v", got)
	}
}

func TestThemeBackgroundPaintsEveryCell(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() {
		ApplyTheme(config.Theme{})
		lipgloss.SetColorProfile(termenv.Ascii)
	})

	sheets := []config.Cheatsheet{{Name: "Vim", Sections: []config.Section{{
		Title: "Move", Bindings: []config.Binding{{Keys: "dd", Desc: "delete line"}},
	}}}}
	const w, h = 80, 12

	// A terminal-native theme (no background) leaves the view unpainted.
	ApplyTheme(config.Theme{})
	plain := New(sheets)
	plainView := mustView(plain, w, h)
	if strings.Contains(plainView, "\x1b[48;2;16;60;72m") {
		t.Fatal("default theme should not paint a background")
	}

	// Selenized dark fills every line edge-to-edge with bg_0 (#103c48).
	dark, _ := config.Preset("selenized-dark")
	ApplyTheme(dark)
	view := mustView(New(sheets), w, h)
	bgOpen := "\x1b[48;2;16;60;72m"

	lines := strings.Split(view, "\n")
	if len(lines) != h {
		t.Fatalf("painted view has %d lines, want %d", len(lines), h)
	}
	strip := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	for i, ln := range lines {
		if !strings.HasPrefix(ln, bgOpen) {
			t.Fatalf("line %d does not start with the background", i)
		}
		if got := lipgloss.Width(strip.ReplaceAllString(ln, "")); got != w {
			t.Fatalf("line %d width = %d, want %d", i, got, w)
		}
	}
}

func mustView(m Model, w, h int) string {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model).View()
}

func TestApplyThemeDefaultsWhenEmpty(t *testing.T) {
	t.Cleanup(func() { ApplyTheme(config.Theme{}) })

	ApplyTheme(config.Theme{})
	if got := kbdStyle.GetForeground(); got != lipgloss.Color("#22D3EE") {
		t.Fatalf("keycap color = %v, want default #22D3EE", got)
	}
}
