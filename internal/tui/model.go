// Package tui renders the cheatsheet terminal UI with Bubble Tea.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sam/cheatsheet-tui/internal/config"
	"github.com/sam/cheatsheet-tui/internal/search"
)

type mode int

const (
	modeNormal mode = iota
	modeSearch
)

// Model is the root Bubble Tea model.
type Model struct {
	sheets   []config.Cheatsheet
	all      []search.Item   // every binding, flattened (for search)
	perSheet [][]search.Item // bindings per sheet, aligned with sheets

	sheetIdx int // selected sheet in the sidebar
	cursor   int // selected row within the visible list

	mode  mode
	query string

	width, height int
	ready         bool
}

// New builds a Model from already-sorted cheatsheets.
func New(sheets []config.Cheatsheet) Model {
	m := Model{sheets: sheets}
	for _, sh := range sheets {
		var items []search.Item
		for _, sec := range sh.Sections {
			for _, b := range sec.Bindings {
				it := search.Item{Sheet: sh.Name, Section: sec.Title, Keys: b.Keys, Desc: b.Desc}
				items = append(items, it)
				m.all = append(m.all, it)
			}
		}
		m.perSheet = append(m.perSheet, items)
	}
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// current returns the selected cheatsheet.
func (m Model) current() config.Cheatsheet {
	if len(m.sheets) == 0 {
		return config.Cheatsheet{}
	}
	return m.sheets[m.sheetIdx]
}

// visible returns the rows shown in the main pane for the current state.
func (m Model) visible() []search.Item {
	if m.mode == modeSearch {
		return search.Filter(m.all, m.query)
	}
	if len(m.perSheet) == 0 {
		return nil
	}
	return m.perSheet[m.sheetIdx]
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		return m, nil
	case tea.KeyMsg:
		if m.mode == modeSearch {
			return m.updateSearch(msg)
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "/":
		m.mode = modeSearch
		m.query = ""
		m.cursor = 0
		return m, nil
	case "j", "down":
		m.moveCursor(1)
	case "k", "up":
		m.moveCursor(-1)
	case "g", "home":
		m.cursor = 0
	case "G", "end":
		m.cursor = m.lastIndex()
	case "ctrl+d":
		m.moveCursor(m.pageStep())
	case "ctrl+u":
		m.moveCursor(-m.pageStep())
	case "tab", "l", "right":
		m.switchSheet(1)
	case "shift+tab", "h", "left":
		m.switchSheet(-1)
	}
	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeNormal
		m.query = ""
		m.cursor = 0
		return m, nil
	case tea.KeyEnter, tea.KeyDown:
		m.moveCursor(1)
		return m, nil
	case tea.KeyUp:
		m.moveCursor(-1)
		return m, nil
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyBackspace:
		if n := len(m.query); n > 0 {
			m.query = m.query[:n-1]
		}
		m.clampCursor()
		return m, nil
	case tea.KeyRunes, tea.KeySpace:
		m.query += string(msg.Runes)
		if msg.Type == tea.KeySpace {
			m.query += " "
		}
		m.cursor = 0
		return m, nil
	}
	return m, nil
}

// --- cursor / navigation helpers ---

func (m *Model) moveCursor(delta int) {
	m.cursor += delta
	m.clampCursor()
}

func (m *Model) clampCursor() {
	last := m.lastIndex()
	if m.cursor > last {
		m.cursor = last
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m Model) lastIndex() int {
	n := len(m.visible())
	if n == 0 {
		return 0
	}
	return n - 1
}

func (m *Model) switchSheet(delta int) {
	if len(m.sheets) == 0 {
		return
	}
	m.sheetIdx = (m.sheetIdx + delta + len(m.sheets)) % len(m.sheets)
	m.cursor = 0
}

// pageStep is how far ctrl-d/ctrl-u jump.
func (m Model) pageStep() int {
	return max(m.listHeight()/2, 1)
}
