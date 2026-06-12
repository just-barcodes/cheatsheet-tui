package tui

import "github.com/charmbracelet/lipgloss"

// A compact, modern palette: violet accent, cyan for keycaps, muted grays.
var (
	accent   = lipgloss.Color("#A78BFA") // violet
	accentHi = lipgloss.Color("#C4B5FD") // lighter violet
	keyColor = lipgloss.Color("#22D3EE") // cyan
	fg       = lipgloss.Color("#E5E7EB")
	dim      = lipgloss.Color("#6B7280")
	subtle   = lipgloss.Color("#3F3F46")
	selBg    = lipgloss.Color("#312E81") // deep indigo selection bar

	titleStyle = lipgloss.NewStyle().
			Bold(true).Foreground(lipgloss.Color("#0B0B12")).
			Background(accent).Padding(0, 1)

	titleDescStyle = lipgloss.NewStyle().Foreground(dim).Italic(true)

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).BorderForeground(subtle).
			Padding(0, 1)

	paneActiveStyle = paneStyle.BorderForeground(accent)

	sidebarItem    = lipgloss.NewStyle().Foreground(fg).Padding(0, 1)
	sidebarItemSel = lipgloss.NewStyle().Bold(true).
			Foreground(lipgloss.Color("#0B0B12")).Background(accent).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().Bold(true).Foreground(accentHi).
			MarginTop(1)

	kbdStyle = lipgloss.NewStyle().Bold(true).
			Foreground(keyColor)

	descStyle    = lipgloss.NewStyle().Foreground(fg)
	descSelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)

	rowSelStyle = lipgloss.NewStyle().Background(selBg)

	searchPrompt = lipgloss.NewStyle().Bold(true).Foreground(accent)
	searchText   = lipgloss.NewStyle().Foreground(fg)
	placeholder  = lipgloss.NewStyle().Foreground(dim).Italic(true)
	countStyle   = lipgloss.NewStyle().Foreground(dim)

	footerStyle = lipgloss.NewStyle().Foreground(dim)
	footerKey   = lipgloss.NewStyle().Foreground(accentHi).Bold(true)
)
