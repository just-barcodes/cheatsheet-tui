package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sam/cheatsheet-tui/internal/config"
)

// sampleSheets are pre-sorted by name, exactly as config.LoadDir delivers them.
func sampleSheets() []config.Cheatsheet {
	return []config.Cheatsheet{
		{
			Name: "Hyprland", Description: "Window manager",
			Sections: []config.Section{{
				Title: "Window",
				Bindings: []config.Binding{
					{Keys: "SUPER+Q", Desc: "Close window"},
				},
			}},
		},
		{
			Name: "Vim", Description: "Modal editor",
			Sections: []config.Section{{
				Title: "Movement",
				Bindings: []config.Binding{
					{Keys: "dd", Desc: "Delete line"},
					{Keys: "yy", Desc: "Yank line"},
				},
			}},
		},
	}
}

// ready returns a sized model so View() renders real content.
func ready() Model {
	m := New(sampleSheets())
	next, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 30})
	return next.(Model)
}

func press(m Model, key string) Model {
	var msg tea.KeyMsg
	switch key {
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "backspace":
		msg = tea.KeyMsg{Type: tea.KeyBackspace}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
	next, _ := m.Update(msg)
	return next.(Model)
}

func typeStr(m Model, s string) Model {
	for _, r := range s {
		m = press(m, string(r))
	}
	return m
}

func TestInitialViewShowsFirstSheet(t *testing.T) {
	m := ready()
	v := m.View()
	if !strings.Contains(v, "Hyprland") || !strings.Contains(v, "Vim") {
		t.Fatalf("expected both sheet names in sidebar, got:\n%s", v)
	}
	// First sheet alphabetically is Hyprland; its binding should be visible.
	if !strings.Contains(v, "Close window") {
		t.Fatalf("expected first sheet's binding visible, got:\n%s", v)
	}
}

func TestSlashEntersSearchAndFilters(t *testing.T) {
	m := ready()
	m = press(m, "/")
	if m.mode != modeSearch {
		t.Fatalf("expected search mode after '/'")
	}
	m = typeStr(m, "delete")
	v := m.View()
	if !strings.Contains(v, "Delete line") {
		t.Fatalf("expected matching binding shown, got:\n%s", v)
	}
	if strings.Contains(v, "Close window") {
		t.Fatalf("expected non-matching binding hidden, got:\n%s", v)
	}
}

func TestEscExitsSearch(t *testing.T) {
	m := ready()
	m = press(m, "/")
	m = typeStr(m, "del")
	m = press(m, "esc")
	if m.mode != modeNormal {
		t.Fatalf("expected normal mode after esc")
	}
	if m.query != "" {
		t.Fatalf("expected query cleared after esc, got %q", m.query)
	}
}

func TestJKMovesCursorWithinBounds(t *testing.T) {
	m := ready()
	// Switch to Vim (2 bindings) so we can move.
	m = press(m, "tab")
	if got := m.current().Name; got != "Vim" {
		t.Fatalf("tab should select Vim, got %q", got)
	}
	if m.cursor != 0 {
		t.Fatalf("cursor should reset to 0 on sheet switch, got %d", m.cursor)
	}
	m = press(m, "j")
	if m.cursor != 1 {
		t.Fatalf("j should move cursor to 1, got %d", m.cursor)
	}
	// Can't move past the last item.
	m = press(m, "j")
	if m.cursor != 1 {
		t.Fatalf("cursor should clamp at last item, got %d", m.cursor)
	}
	m = press(m, "k")
	if m.cursor != 0 {
		t.Fatalf("k should move cursor to 0, got %d", m.cursor)
	}
	m = press(m, "k")
	if m.cursor != 0 {
		t.Fatalf("cursor should clamp at 0, got %d", m.cursor)
	}
}

func TestQuitKeys(t *testing.T) {
	m := ready()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatalf("expected quit command on 'q'")
	}
	if msg := cmd(); msg == nil {
		t.Fatalf("expected non-nil quit msg")
	} else if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", msg)
	}
}
