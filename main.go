// Command cheatsheet is a fast terminal cheat-sheet viewer. It reads hotkeys
// from YAML files (one per program/app/type) and shows them in a searchable,
// vim-navigable TUI.
//
// Cheatsheets are read from the first of: the --dir flag, $CHEATSHEET_DIR, your
// config dir (~/.config/cheatsheet), or the built-in defaults. Run with --init
// to copy the built-ins into your config dir so you can edit your own.
package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sam/cheatsheet-tui/internal/config"
	"github.com/sam/cheatsheet-tui/internal/tui"
)

//go:embed cheatsheets/*.yaml
var builtinRoot embed.FS

// builtinFS exposes the embedded cheatsheets with the "cheatsheets/" prefix
// stripped, so files sit at the root like a real directory.
func builtinFS() fs.FS {
	sub, err := fs.Sub(builtinRoot, "cheatsheets")
	if err != nil {
		panic(err) // embed path is a compile-time constant; cannot fail
	}
	return sub
}

func main() {
	dir := flag.String("dir", "", "directory of cheatsheet .yaml files (overrides the default search)")
	doInit := flag.Bool("init", false, "copy the built-in cheatsheets into your config dir, then exit")
	flag.Parse()

	cfgDir := configDir()

	if *doInit {
		if err := scaffold(cfgDir); err != nil {
			fatal(err)
		}
		return
	}

	src := config.Locator{
		Flag:      *dir,
		Env:       os.Getenv("CHEATSHEET_DIR"),
		ConfigDir: cfgDir,
		DirExists: dirExists,
	}.Resolve()

	sheets, err := loadSource(src)
	if err != nil {
		fatal(err)
	}
	if len(sheets) == 0 {
		fmt.Fprintf(os.Stderr, "cheatsheet: no .yaml cheatsheets found in %s\n", src.Path)
		if cfgDir != "" {
			fmt.Fprintf(os.Stderr, "Run 'cheatsheet --init' to create starter cheatsheets in %s\n", cfgDir)
		}
		os.Exit(1)
	}

	p := tea.NewProgram(tui.New(sheets), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fatal(err)
	}
}

func loadSource(src config.Source) ([]config.Cheatsheet, error) {
	if src.Builtin {
		return config.LoadFS(builtinFS())
	}
	return config.LoadDir(src.Path)
}

// scaffold writes the built-in cheatsheets into dir, skipping any that already
// exist, so users get editable starting files without clobbering their own.
func scaffold(dir string) error {
	if dir == "" {
		return fmt.Errorf("could not determine a config directory")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entries, err := fs.ReadDir(builtinFS(), ".")
	if err != nil {
		return err
	}
	written, skipped := 0, 0
	for _, e := range entries {
		dst := filepath.Join(dir, e.Name())
		if fileExists(dst) {
			skipped++
			continue
		}
		data, err := fs.ReadFile(builtinFS(), e.Name())
		if err != nil {
			return err
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return err
		}
		written++
	}
	fmt.Printf("Wrote %d cheatsheet(s) to %s", written, dir)
	if skipped > 0 {
		fmt.Printf(" (%d already present, left untouched)", skipped)
	}
	fmt.Printf("\nEdit them there, then just run: cheatsheet\n")
	return nil
}

// configDir is ~/.config/cheatsheet (or the OS equivalent), or "" if unknown.
func configDir() string {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(cfg, "cheatsheet")
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "cheatsheet: %v\n", err)
	os.Exit(1)
}
