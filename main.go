// Command cheatsheet is a fast terminal cheat-sheet viewer. It reads hotkeys
// from YAML files (one per program/app/type) and shows them in a searchable,
// vim-navigable TUI.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sam/cheatsheet-tui/internal/config"
	"github.com/sam/cheatsheet-tui/internal/tui"
)

func main() {
	dir := flag.String("dir", defaultDir(), "directory of cheatsheet .yaml files")
	flag.Parse()

	sheets, err := config.LoadDir(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cheatsheet: %v\n", err)
		os.Exit(1)
	}
	if len(sheets) == 0 {
		fmt.Fprintf(os.Stderr, "cheatsheet: no .yaml cheatsheets found in %s\n", *dir)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.New(sheets), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "cheatsheet: %v\n", err)
		os.Exit(1)
	}
}

// defaultDir resolves where cheatsheets live: $CHEATSHEET_DIR, else
// $XDG_CONFIG_HOME/cheatsheet (or ~/.config/cheatsheet), else ./cheatsheets.
func defaultDir() string {
	if d := os.Getenv("CHEATSHEET_DIR"); d != "" {
		return d
	}
	if cfg, err := os.UserConfigDir(); err == nil {
		p := filepath.Join(cfg, "cheatsheet")
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	return "cheatsheets"
}
