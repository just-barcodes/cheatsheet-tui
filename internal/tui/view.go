package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/just-barcodes/cheatsheet-tui/internal/search"
)

const (
	sidebarWidth = 22
	chromeRows   = 5  // title + search + footer + two pane borders
	colGap       = 2  // blank cells between newspaper columns
	maxColumns   = 3  // hotkeys never split into more than three columns
	rowIndent    = 2  // leading spaces before each binding row
	keyGap       = 1  // cells between the key column and the description
	minDescWidth = 12 // narrowest a description may get before we drop a column
	maxKeyWidth  = 24 // keys wider than this truncate rather than starve the desc
)

// listHeight is how many body rows the main pane can show.
func (m Model) listHeight() int {
	h := m.height - chromeRows
	if h < 1 {
		return 1
	}
	return h
}

// mainWidth is the outer width of the hotkey pane; mainInnerWidth is the space
// left for content once the pane border and padding are removed. The 4 covers
// the sidebar's and the main pane's two vertical borders, so sidebar + main
// exactly fills the terminal width (no 2-column overhang on the right).
func (m Model) mainWidth() int      { return max(m.width-sidebarWidth-4, 20) }
func (m Model) mainInnerWidth() int { return m.mainWidth() - 4 }

// View implements tea.Model.
func (m Model) View() string {
	if !m.ready {
		return "loading…"
	}
	if len(m.sheets) == 0 {
		return "No cheatsheets found. Add a .yaml file to your cheatsheets directory."
	}

	lay := m.layout(m.visible(), m.mainInnerWidth(), m.listHeight())
	body := lipgloss.JoinHorizontal(lipgloss.Top, m.sidebar(), m.mainPane(lay))
	view := strings.Join([]string{
		m.titleBar(),
		m.searchBar(),
		body,
		m.footer(),
	}, "\n")
	return paintBackground(view, m.width, m.height)
}

// paintBackground fills the whole terminal with the theme background. lipgloss
// emits a reset (\x1b[0m) at the end of every styled span, which drops the
// background for the cells that follow; this re-asserts it after each reset and
// at every line start, then pads each line to full width and the screen to full
// height. Spans that set their own background (the title/sidebar bars, the
// selected row) are single styled blocks, so their fill survives untouched and
// the screen background simply resumes after them.
func paintBackground(view string, w, h int) string {
	if !hasBackground {
		return view
	}
	const reset = "\x1b[0m"
	lines := strings.Split(view, "\n")
	out := make([]string, 0, max(h, len(lines)))
	for _, ln := range lines {
		if lipgloss.Width(ln) > w { // a line wider than the screen would wrap
			ln = lipgloss.NewStyle().MaxWidth(w).Render(ln)
		}
		painted := bgOpenSeq + strings.ReplaceAll(ln, reset, reset+bgOpenSeq)
		if pad := w - lipgloss.Width(ln); pad > 0 {
			painted += strings.Repeat(" ", pad)
		}
		out = append(out, painted+reset)
	}
	for len(out) < h {
		out = append(out, bgOpenSeq+strings.Repeat(" ", w)+reset)
	}
	return strings.Join(out, "\n")
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

// line is one rendered row of the main pane: either a section header, a blank
// section gap, or one physical row of a binding (itemIdx >= 0). A binding whose
// description wraps contributes several rows that share its itemIdx.
type line struct {
	text    string
	itemIdx int
}

// paneLayout is the resolved geometry of the hotkey pane for one render.
type paneLayout struct {
	cols, colWidth, contentWidth, capacity int
	lines                                  []line
	overflow                               bool
	tooSmall                               bool
}

// layout resolves how the visible items pack into newspaper columns. Because
// descriptions wrap, the line count depends on the column width, so the lines
// are built here at the chosen width rather than probed up front.
func (m Model) layout(items []search.Item, innerWidth, h int) paneLayout {
	keyW := keyColumnWidth(items)
	cols := m.columnCount(innerWidth, keyW)
	if cols < 1 {
		return paneLayout{tooSmall: true}
	}

	contentWidth := innerWidth
	colWidth := colWidthFor(contentWidth, cols)
	lines := m.buildLines(items, colWidth, keyW)

	// Use only as many columns as the content fills; fewer columns make the
	// ones in use wider rather than leaving a blank column on the right.
	if needed := max((len(lines)+h-1)/h, 1); cols > needed {
		cols = needed
		colWidth = colWidthFor(contentWidth, cols)
		lines = m.buildLines(items, colWidth, keyW)
	}

	// Reserve a scrollbar column only when the paged capacity (cols*h) overflows
	// and a readable column still survives the two cells it costs.
	overflow := len(lines) > cols*h
	if overflow {
		if c := m.columnCount(contentWidth-2, keyW); c >= 1 {
			contentWidth -= 2
			cols = c
			colWidth = colWidthFor(contentWidth, cols)
			lines = m.buildLines(items, colWidth, keyW)
			overflow = len(lines) > cols*h
		} else {
			overflow = false // too tight for a scrollbar; use the full width
		}
	}
	return paneLayout{cols, colWidth, contentWidth, cols * h, lines, overflow, false}
}

func colWidthFor(width, cols int) int {
	return (width - colGap*(cols-1)) / cols
}

// LayoutColumns reports how many newspaper columns the hotkey pane renders at
// the current terminal size. Exposed so behavioral tests can assert the layout.
func (m Model) LayoutColumns() int {
	if !m.ready {
		return 0
	}
	return m.layout(m.visible(), m.mainInnerWidth(), m.listHeight()).cols
}

// columnCount is how many newspaper columns to render. With no user override
// (m.cols == 0) it auto-fits the width; otherwise it honors the requested
// count. Either way it never gives a column narrower than is readable, and
// returns 0 when not even one column fits (the caller shows a resize prompt).
func (m Model) columnCount(innerWidth, keyW int) int {
	minW := rowIndent + keyW + keyGap + minDescWidth
	fit := (innerWidth + colGap) / (minW + colGap)
	if fit < 1 {
		return 0
	}
	if m.cols > 0 {
		return min(m.cols, min(fit, maxColumns))
	}
	return min(fit, maxColumns)
}

func (m Model) mainPane(lay paneLayout) string {
	style := paneStyle
	if m.mode == modeSearch {
		style = paneActiveStyle
	}
	width := m.mainWidth()
	h := m.listHeight()

	if lay.tooSmall {
		msg := placeholder.Render("Window too small — enlarge it to view hotkeys")
		return style.Width(width).Height(h).Render(msg)
	}

	visibleLines, start := m.window(lay.lines, lay.capacity)
	content := flowColumns(visibleLines, lay.cols, h, lay.colWidth)
	if len(lay.lines) == 0 {
		content = placeholder.Render("no matches")
	}
	if lay.overflow {
		bar := scrollbar(h, len(lay.lines), lay.capacity, start)
		content = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(lay.contentWidth+1).Render(content), bar)
	}
	return style.Width(width).Height(h).Render(content)
}

// flowColumns lays the windowed lines into up to cols newspaper columns of h
// rows each — the first h lines fill column one, the next h fill column two,
// and so on — then joins the columns side by side.
func flowColumns(lines []line, cols, h, colWidth int) string {
	blocks := make([]string, 0, cols)
	for c := range cols {
		lo := c * h
		if lo >= len(lines) {
			break
		}
		hi := min(lo+h, len(lines))
		var b strings.Builder
		for i := lo; i < hi; i++ {
			b.WriteString(lines[i].text)
			if i < hi-1 {
				b.WriteByte('\n')
			}
		}
		style := lipgloss.NewStyle().Width(colWidth)
		if c < cols-1 {
			style = style.MarginRight(colGap)
		}
		blocks = append(blocks, style.Render(b.String()))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, blocks...)
}

// scrollbar renders a vertical track of height h with a proportional thumb.
// visible is how many of the total lines the paged columns show at once.
func scrollbar(h, total, visible, start int) string {
	thumbLen := max(h*visible/total, 1)
	maxStart := total - visible
	thumbPos := 0
	if maxStart > 0 {
		thumbPos = start * (h - thumbLen) / maxStart
	}
	var b strings.Builder
	for i := range h {
		if i > 0 {
			b.WriteByte('\n')
		}
		if i >= thumbPos && i < thumbPos+thumbLen {
			b.WriteString(scrollThumb.Render("█"))
		} else {
			b.WriteString(scrollTrack.Render("░"))
		}
	}
	return b.String()
}

// keyColumnWidth sizes the keycap column to the widest key on screen so keys
// stay on one line, clamped to a floor (short keys still read as a column) and
// a ceiling (a pathologically long key truncates rather than starve the desc).
func keyColumnWidth(items []search.Item) int {
	widest := 0
	for _, it := range items {
		if w := lipgloss.Width(it.Keys); w > widest {
			widest = w
		}
	}
	return min(max(widest, 6), maxKeyWidth)
}

// buildLines flattens visible items into display lines, inserting a section
// header whenever the (sheet, section) grouping changes. A binding whose
// description wraps contributes one line per wrapped row.
func (m Model) buildLines(items []search.Item, colWidth, keyW int) []line {
	var lines []line
	prevGroup := ""
	for idx, it := range items {
		group := it.Sheet + "\x00" + it.Section
		if group != prevGroup {
			if prevGroup != "" {
				lines = append(lines, line{text: "", itemIdx: -1}) // section gap
			}
			header := it.Section
			if m.mode == modeSearch {
				header = it.Sheet + " · " + it.Section
			}
			lines = append(lines, line{text: sectionStyle.Render(header), itemIdx: -1})
			prevGroup = group
		}
		for _, row := range renderItemRows(it, idx == m.cursor, colWidth, keyW) {
			lines = append(lines, line{text: row, itemIdx: idx})
		}
	}
	return lines
}

// renderItemRows renders one binding into one-or-more physical rows: the key
// sits beside the first line of the description, and continuation lines of a
// wrapped description align under it with a blank key column.
//
// The selected row is emitted as a single styled block so its highlight bar is
// one continuous fill — both prettier and a precondition for clean background
// painting, which re-asserts the screen background after every style reset.
func renderItemRows(it search.Item, selected bool, colWidth, keyW int) []string {
	descW := max(colWidth-keyW-rowIndent-keyGap, minDescWidth)
	indent := strings.Repeat(" ", rowIndent)
	gap := strings.Repeat(" ", keyGap)

	if selected {
		descLines := strings.Split(plainWrap(it.Desc, descW), "\n")
		rows := make([]string, len(descLines))
		for i, dl := range descLines {
			keyCell := strings.Repeat(" ", keyW)
			if i == 0 {
				keyCell = plainCell(it.Keys, keyW)
			}
			rows[i] = rowSelStyle.Width(colWidth).Render(indent + keyCell + gap + dl)
		}
		return rows
	}

	descLines := strings.Split(descStyle.Width(descW).Render(it.Desc), "\n")
	rows := make([]string, len(descLines))
	for i, dl := range descLines {
		keyCell := strings.Repeat(" ", keyW)
		if i == 0 {
			keyCell = kbdStyle.Width(keyW).MaxHeight(1).Render(it.Keys)
		}
		rows[i] = lipgloss.JoinHorizontal(lipgloss.Top, indent, keyCell, gap, dl)
	}
	return rows
}

// plainWrap word-wraps s to width w and pads each line to w cells, with no
// styling, so it can be wrapped in a single background-bearing style without
// inner resets breaking the fill.
func plainWrap(s string, w int) string {
	return ansiSeq.ReplaceAllString(lipgloss.NewStyle().Width(w).Render(s), "")
}

// plainCell pads or truncates s to a single w-cell line, unstyled.
func plainCell(s string, w int) string {
	return ansiSeq.ReplaceAllString(lipgloss.NewStyle().Width(w).MaxHeight(1).Render(s), "")
}

var ansiSeq = regexp.MustCompile("\x1b\\[[0-9;]*m")

// window returns the slice of display lines that fits the paged columns plus
// its start offset. Scroll position is derived purely from the cursor: the
// selected row is kept centred once the list is long enough to scroll, which is
// stateless and always keeps context visible above and below.
func (m Model) window(lines []line, capacity int) ([]line, int) {
	if len(lines) <= capacity {
		return lines, 0
	}
	cursorLine := 0
	for i, ln := range lines {
		if ln.itemIdx == m.cursor {
			cursorLine = i
			break
		}
	}
	start := min(max(cursorLine-capacity/2, 0), len(lines)-capacity)
	return lines[start : start+capacity], start
}

func (m Model) footer() string {
	var keys []struct{ k, label string }
	if m.mode == modeSearch {
		keys = []struct{ k, label string }{
			{"type", "filter"}, {"↑↓", "move"}, {"esc", "exit"}, {"^c", "quit"},
		}
	} else {
		// The column control always reflects the chosen setting — "auto" or the
		// pinned count — not how many columns happen to be visible right now.
		cols := "auto"
		if m.cols > 0 {
			cols = fmt.Sprintf("%d", m.cols)
		}
		keys = []struct{ k, label string }{
			{"j/k", "move"}, {"h/l·tab", "sheet"}, {"g/G", "top/bottom"},
			{"c", "cols:" + cols}, {"/", "search"}, {"q", "quit"},
		}
	}
	parts := make([]string, len(keys))
	for i, e := range keys {
		parts[i] = footerKey.Render(e.k) + " " + footerStyle.Render(e.label)
	}
	return footerStyle.Render(strings.Join(parts, footerStyle.Render("  ·  ")))
}
