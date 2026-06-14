package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/just-barcodes/cheatsheet-tui/internal/config"
)

// The UI palette and the styles built from it. These are process-global and set
// once at startup: applyPalette installs the defaults at init, and ApplyTheme
// re-installs them with the user's overrides before the program runs.
var (
	titleStyle, titleDescStyle  lipgloss.Style
	paneStyle, paneActiveStyle  lipgloss.Style
	sidebarItem, sidebarItemSel lipgloss.Style
	sectionStyle                lipgloss.Style
	scrollThumb, scrollTrack    lipgloss.Style
	kbdStyle, descStyle         lipgloss.Style
	rowSelStyle                 lipgloss.Style
	searchPrompt, searchText    lipgloss.Style
	placeholder, countStyle     lipgloss.Style
	footerStyle, footerKey      lipgloss.Style

	// Background painting (set by applyPalette). hasBackground is false for
	// terminal-native themes; bgOpenSeq is the SGR that turns the background on.
	hasBackground bool
	bgOpenSeq     string
)

func init() { applyPalette(defaultPalette()) }

// palette is a compact, modern set: violet accent, cyan keycaps, muted grays.
type palette struct {
	bg       lipgloss.Color // UI background; empty keeps the terminal's own
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
	set(&p.bg, t.Colors.Background)
	set(&p.accent, t.Colors.Accent)
	set(&p.accentHi, t.Colors.AccentBright)
	set(&p.keyColor, t.Colors.Keycap)
	set(&p.fg, t.Colors.Foreground)
	set(&p.dim, t.Colors.Muted)
	set(&p.subtle, t.Colors.Border)
	set(&p.selBg, t.Colors.Selection)
	applyPalette(p)
}

// applyPalette rebuilds every style from p. Text drawn on top of a colored fill
// (the title/sidebar accent bar, the selected row) picks black or white by the
// fill's luminance, so both dark and light palettes stay readable.
func applyPalette(p palette) {
	onAccent := readableOn(p.accent)
	onSelect := readableOn(p.selBg)

	hasBackground = p.bg != ""
	bgOpenSeq = ""
	if hasBackground {
		bgOpenSeq = backgroundOpenSeq(p.bg)
	}

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

	// The selected row is rendered as one styled block (see renderItemRows), so
	// the highlight bar carries both its fill and its text style here.
	rowSelStyle = lipgloss.NewStyle().Background(p.selBg).Foreground(onSelect).Bold(true)

	searchPrompt = lipgloss.NewStyle().Bold(true).Foreground(p.accent)
	searchText = lipgloss.NewStyle().Foreground(p.fg)
	placeholder = lipgloss.NewStyle().Foreground(p.dim).Italic(true)
	countStyle = lipgloss.NewStyle().Foreground(p.dim)

	footerStyle = lipgloss.NewStyle().Foreground(p.dim)
	footerKey = lipgloss.NewStyle().Foreground(p.accentHi).Bold(true)
}

// readableOn returns near-black or near-white text, whichever contrasts better
// with the bg fill. Non-hex colors (an ANSI index) fall back to white, assuming
// the common dark terminal.
func readableOn(bg lipgloss.Color) lipgloss.Color {
	const (
		dark  = lipgloss.Color("#0B0B12")
		light = lipgloss.Color("#FFFFFF")
	)
	r, g, b, ok := parseHexColor(string(bg))
	if !ok {
		return light
	}
	// Perceptual luminance on a 0-255 scale.
	if 0.2126*float64(r)+0.7152*float64(g)+0.0722*float64(b) > 150 {
		return dark
	}
	return light
}

// backgroundOpenSeq returns just the SGR that switches the background on for
// color c, honoring the active color profile. It is derived by rendering a
// single space and taking everything before it.
func backgroundOpenSeq(c lipgloss.Color) string {
	s := lipgloss.NewStyle().Background(c).Render(" ")
	if i := strings.IndexByte(s, ' '); i > 0 {
		return s[:i]
	}
	return "" // no-color profile: nothing to set
}

func parseHexColor(s string) (r, g, b uint8, ok bool) {
	if len(s) == 4 && s[0] == '#' { // expand #RGB to #RRGGBB
		s = string([]byte{'#', s[1], s[1], s[2], s[2], s[3], s[3]})
	}
	if len(s) != 7 || s[0] != '#' {
		return 0, 0, 0, false
	}
	v, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	return uint8(v >> 16), uint8(v >> 8), uint8(v), true
}
