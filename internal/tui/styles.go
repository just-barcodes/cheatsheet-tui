package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/just-barcodes/cheatsheet-tui/internal/config"
)

// The UI palette and the styles built from it. These are process-global and set
// once at startup: applyPalette installs the defaults at init, and ApplyTheme
// re-installs them with the user's overrides before the program runs.
var (
	titleStyle, titleDescStyle        lipgloss.Style
	paneStyle, paneActiveStyle        lipgloss.Style
	sidebarItem, sidebarItemSel       lipgloss.Style
	sectionStyle                      lipgloss.Style
	scrollThumb, scrollTrack          lipgloss.Style
	kbdStyle, descStyle, descSelStyle lipgloss.Style
	rowSelStyle                       lipgloss.Style
	searchPrompt, searchText          lipgloss.Style
	placeholder, countStyle           lipgloss.Style
	footerStyle, footerKey            lipgloss.Style
)

func init() { applyPalette(defaultPalette()) }

// palette is a compact, modern set: violet accent, cyan keycaps, muted grays.
type palette struct {
	accent   lipgloss.Color // headings, active border, search prompt
	accentHi lipgloss.Color // section titles, footer keys
	keyColor lipgloss.Color // the hotkeys themselves
	fg       lipgloss.Color // descriptions and body text
	dim      lipgloss.Color // hints, counts, inactive text
	subtle   lipgloss.Color // inactive pane borders, scrollbar track
	selBg    lipgloss.Color // highlighted row background
}

func defaultPalette() palette {
	return palette{
		accent:   lipgloss.Color("#A78BFA"),
		accentHi: lipgloss.Color("#C4B5FD"),
		keyColor: lipgloss.Color("#22D3EE"),
		fg:       lipgloss.Color("#E5E7EB"),
		dim:      lipgloss.Color("#6B7280"),
		subtle:   lipgloss.Color("#3F3F46"),
		selBg:    lipgloss.Color("#312E81"),
	}
}

// ApplyTheme installs the palette with the user's colour overrides merged over
// the defaults. Call it once, before constructing the program.
func ApplyTheme(t config.Theme) {
	p := defaultPalette()
	set := func(dst *lipgloss.Color, v string) {
		if v != "" {
			*dst = lipgloss.Color(v)
		}
	}
	set(&p.accent, t.Colors.Accent)
	set(&p.accentHi, t.Colors.AccentBright)
	set(&p.keyColor, t.Colors.Keycap)
	set(&p.fg, t.Colors.Foreground)
	set(&p.dim, t.Colors.Muted)
	set(&p.subtle, t.Colors.Border)
	set(&p.selBg, t.Colors.Selection)
	applyPalette(p)
}

// applyPalette rebuilds every style from p. Foregrounds that exist purely for
// contrast against an accent fill (near-black on accent, white on selection)
// stay fixed so text never becomes unreadable.
func applyPalette(p palette) {
	const (
		onAccent = lipgloss.Color("#0B0B12")
		onSelect = lipgloss.Color("#FFFFFF")
	)

	titleStyle = lipgloss.NewStyle().
		Bold(true).Foreground(onAccent).
		Background(p.accent).Padding(0, 1)
	titleDescStyle = lipgloss.NewStyle().Foreground(p.dim).Italic(true)

	paneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(p.subtle).
		Padding(0, 1)
	paneActiveStyle = paneStyle.BorderForeground(p.accent)

	sidebarItem = lipgloss.NewStyle().Foreground(p.fg).Padding(0, 1)
	sidebarItemSel = lipgloss.NewStyle().Bold(true).
		Foreground(onAccent).Background(p.accent).
		Padding(0, 1)

	// No vertical margin: every rendered line must be exactly one terminal row
	// so the scroll window math in view.go stays exact.
	sectionStyle = lipgloss.NewStyle().Bold(true).Foreground(p.accentHi)

	scrollThumb = lipgloss.NewStyle().Foreground(p.accent)
	scrollTrack = lipgloss.NewStyle().Foreground(p.subtle)

	kbdStyle = lipgloss.NewStyle().Bold(true).Foreground(p.keyColor)

	descStyle = lipgloss.NewStyle().Foreground(p.fg)
	descSelStyle = lipgloss.NewStyle().Foreground(onSelect).Bold(true)

	rowSelStyle = lipgloss.NewStyle().Background(p.selBg)

	searchPrompt = lipgloss.NewStyle().Bold(true).Foreground(p.accent)
	searchText = lipgloss.NewStyle().Foreground(p.fg)
	placeholder = lipgloss.NewStyle().Foreground(p.dim).Italic(true)
	countStyle = lipgloss.NewStyle().Foreground(p.dim)

	footerStyle = lipgloss.NewStyle().Foreground(p.dim)
	footerKey = lipgloss.NewStyle().Foreground(p.accentHi).Bold(true)
}
