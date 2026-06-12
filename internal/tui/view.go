package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/sam/cheatsheet-tui/internal/search"
)

const (
	sidebarWidth = 22
	chromeRows   = 5 // title + search + footer + two pane borders
)

// listHeight is how many body rows the main pane can show.
func (m Model) listHeight() int {
	h := m.height - chromeRows
	if h < 1 {
		return 1
	}
	return h
}

// View implements tea.Model.
func (m Model) View() string {
	if !m.ready {
		return "loading…"
	}
	if len(m.sheets) == 0 {
		return "No cheatsheets found. Add a .yaml file to your cheatsheets directory."
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, m.sidebar(), m.mainPane())
	return strings.Join([]string{
		m.titleBar(),
		m.searchBar(),
		body,
		m.footer(),
	}, "\n")
}

func (m Model) titleBar() string {
	cur := m.current()
	left := titleStyle.Render(" ⌘ Cheatsheet ")
	desc := cur.Name
	if cur.Description != "" {
		desc += " — " + cur.Description
	}
	if m.mode == modeSearch {
		desc = "Search across all cheatsheets"
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, left, " ", titleDescStyle.Render(desc))
}

func (m Model) searchBar() string {
	if m.mode == modeSearch {
		q := searchText.Render(m.query) + searchPrompt.Render("▌")
		count := countStyle.Render(fmt.Sprintf("  %d matches", len(m.visible())))
		return searchPrompt.Render("/ ") + q + count
	}
	return placeholder.Render("/ press / to search all hotkeys")
}

func (m Model) sidebar() string {
	var b strings.Builder
	for i, sh := range m.sheets {
		label := sh.Name
		if sh.Icon != "" {
			label = sh.Icon + "  " + label
		}
		if i == m.sheetIdx && m.mode == modeNormal {
			b.WriteString(sidebarItemSel.Width(sidebarWidth - 2).Render(label))
		} else {
			b.WriteString(sidebarItem.Width(sidebarWidth - 2).Render(label))
		}
		if i < len(m.sheets)-1 {
			b.WriteByte('\n')
		}
	}
	style := paneStyle
	if m.mode == modeNormal {
		style = paneActiveStyle
	}
	return style.Width(sidebarWidth).Height(m.listHeight()).Render(b.String())
}

// line is one rendered row of the main pane: either a section header or a
// selectable binding (itemIdx >= 0).
type line struct {
	text    string
	itemIdx int
}

func (m Model) mainPane() string {
	items := m.visible()
	width := max(m.width-sidebarWidth-2, 20) // minus gap + pane border slack
	innerWidth := width - 4 // pane padding + border

	lines := m.buildLines(items, innerWidth)
	visibleLines := m.window(lines)

	var b strings.Builder
	for i, ln := range visibleLines {
		b.WriteString(ln.text)
		if i < len(visibleLines)-1 {
			b.WriteByte('\n')
		}
	}
	if len(items) == 0 {
		b.WriteString(placeholder.Render("no matches"))
	}

	style := paneStyle
	if m.mode == modeSearch {
		style = paneActiveStyle
	}
	return style.Width(width).Height(m.listHeight()).Render(b.String())
}

// buildLines flattens visible items into display lines, inserting a section
// header whenever the (sheet, section) grouping changes.
func (m Model) buildLines(items []search.Item, innerWidth int) []line {
	var lines []line
	prevGroup := ""
	for idx, it := range items {
		group := it.Sheet + "\x00" + it.Section
		if group != prevGroup {
			header := it.Section
			if m.mode == modeSearch {
				header = it.Sheet + " · " + it.Section
			}
			lines = append(lines, line{text: sectionStyle.Render(header), itemIdx: -1})
			prevGroup = group
		}
		lines = append(lines, line{text: m.renderRow(it, idx, innerWidth), itemIdx: idx})
	}
	return lines
}

func (m Model) renderRow(it search.Item, idx, innerWidth int) string {
	selected := idx == m.cursor
	keyW := 16
	key := kbdStyle.Width(keyW).Render(it.Keys)
	descW := max(innerWidth-keyW-1, 4)
	dStyle := descStyle
	if selected {
		dStyle = descSelStyle
	}
	desc := dStyle.MaxWidth(descW).Render(it.Desc)

	row := lipgloss.JoinHorizontal(lipgloss.Top, "  ", key, " ", desc)
	if selected {
		return rowSelStyle.Width(innerWidth).Render(row)
	}
	return row
}

// window returns the slice of display lines that fits the pane, keeping the
// selected binding visible. Scroll position is derived purely from the cursor.
func (m Model) window(lines []line) []line {
	h := m.listHeight()
	if len(lines) <= h {
		return lines
	}
	cursorLine := 0
	for i, ln := range lines {
		if ln.itemIdx == m.cursor {
			cursorLine = i
			break
		}
	}
	start := 0
	if cursorLine >= h {
		start = cursorLine - h + 1
	}
	end := start + h
	if end > len(lines) {
		end = len(lines)
		start = end - h
	}
	return lines[start:end]
}

func (m Model) footer() string {
	var keys []struct{ k, label string }
	if m.mode == modeSearch {
		keys = []struct{ k, label string }{
			{"type", "filter"}, {"↑↓", "move"}, {"esc", "exit"}, {"^c", "quit"},
		}
	} else {
		keys = []struct{ k, label string }{
			{"j/k", "move"}, {"h/l·tab", "sheet"}, {"g/G", "top/bottom"},
			{"/", "search"}, {"q", "quit"},
		}
	}
	parts := make([]string, len(keys))
	for i, e := range keys {
		parts[i] = footerKey.Render(e.k) + " " + footerStyle.Render(e.label)
	}
	return footerStyle.Render(strings.Join(parts, footerStyle.Render("  ·  ")))
}
