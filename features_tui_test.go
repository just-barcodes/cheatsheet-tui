package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/just-barcodes/cheatsheet-tui/internal/config"
	"github.com/just-barcodes/cheatsheet-tui/internal/tui"
)

// --- multi-column layout steps ---

func (s *featureState) aCheatsheetWithBindings(name string, n int) error {
	var binds []config.Binding
	for i := range n {
		binds = append(binds, config.Binding{
			Keys: fmt.Sprintf("k%d", i),
			Desc: fmt.Sprintf("action %d", i),
		})
	}
	sheet := config.Cheatsheet{
		Name:     name,
		Sections: []config.Section{{Title: "All", Bindings: binds}},
	}
	s.model = tui.New([]config.Cheatsheet{sheet})
	return nil
}

func (s *featureState) aCheatsheetWithLongDescription(name, sentinel string) error {
	desc := "This is a deliberately long binding description that must wrap " +
		"across several lines before it finally ends in " + sentinel
	sheet := config.Cheatsheet{
		Name: name,
		Sections: []config.Section{{
			Title:    "All",
			Bindings: []config.Binding{{Keys: "k0", Desc: desc}},
		}},
	}
	s.model = tui.New([]config.Cheatsheet{sheet})
	return nil
}

func (s *featureState) aTerminalSized(w, h int) error {
	next, _ := s.model.Update(tea.WindowSizeMsg{Width: w, Height: h})
	s.model = next.(tui.Model)
	return nil
}

func (s *featureState) iViewTheCheatsheet() error {
	s.rendered = s.model.View()
	return nil
}

func (s *featureState) theHotkeysAreLaidOutInColumns(n int) error {
	if got := s.model.LayoutColumns(); got != n {
		return fmt.Errorf("columns = %d, want %d", got, n)
	}
	return nil
}

func (s *featureState) bindingIsVisible(key string) error {
	if !strings.Contains(s.rendered, key) {
		return fmt.Errorf("expected binding %q visible in:\n%s", key, s.rendered)
	}
	return nil
}

func (s *featureState) bindingIsNotVisible(key string) error {
	if strings.Contains(s.rendered, key) {
		return fmt.Errorf("expected binding %q NOT visible in:\n%s", key, s.rendered)
	}
	return nil
}

func (s *featureState) theScreenShowsColumnCount(text string) error {
	if !strings.Contains(s.rendered, text) {
		return fmt.Errorf("expected column indicator %q in:\n%s", text, s.rendered)
	}
	return nil
}

func (s *featureState) theScreenAsksForALargerWindow() error {
	if !strings.Contains(s.rendered, "too small") {
		return fmt.Errorf("expected a resize prompt in:\n%s", s.rendered)
	}
	return nil
}

// iSetTheColumnCountTo presses the cycle key n times; from the auto default that
// lands on a requested count of n (for n in 1..3).
func (s *featureState) iSetTheColumnCountTo(n int) error {
	return s.iCycleTheColumnCount(n)
}

func (s *featureState) iCycleTheColumnCount(n int) error {
	for range n {
		next, _ := s.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		s.model = next.(tui.Model)
	}
	return nil
}
