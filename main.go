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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	flag "github.com/spf13/pflag"

	"github.com/just-barcodes/cheatsheet-tui/internal/config"
	"github.com/just-barcodes/cheatsheet-tui/internal/tui"
)

//go:embed cheatsheets/*.yaml
var builtinRoot embed.FS

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

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
	flag.CommandLine.Init("cheatsheet", flag.ContinueOnError)
	dir := flag.StringP("dir", "d", "", "directory of cheatsheet .yaml files (overrides the default search)")
	doInit := flag.Bool("init", false, "copy the built-in cheatsheets into your config dir, then exit")
	themePath := flag.StringP("theme", "t", "", "path to a theme.yaml (overrides ~/.config/cheatsheet/theme.yaml)")
	showVersion := flag.BoolP("version", "v", false, "print the version and exit")
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0) // --help is a successful outcome, not an error
		}
		fmt.Fprintf(os.Stderr, "cheatsheet: %v\n", err)
		flag.Usage()
		os.Exit(2)
	}

	if *showVersion {
		fmt.Println("cheatsheet " + version)
		return
	}

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

	themeSrc := config.ThemeLocator{Flag: *themePath, ConfigDir: cfgDir}.Resolve()
	if themeSrc.Path != "" {
		if themeSrc.MustExist && !fileExists(themeSrc.Path) {
			fatal(fmt.Errorf("theme file not found: %s", themeSrc.Path))
		}
		theme, err := config.LoadThemeFile(themeSrc.Path)
		if err != nil {
			fatal(err)
		}
		tui.ApplyTheme(theme)
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
	if dst := filepath.Join(dir, config.ThemeFileName); !fileExists(dst) {
		if err := os.WriteFile(dst, []byte(themeSample), 0o644); err != nil {
			return err
		}
	}

	fmt.Printf("Wrote %d cheatsheet(s) to %s", written, dir)
	if skipped > 0 {
		fmt.Printf(" (%d already present, left untouched)", skipped)
	}
	fmt.Printf("\nEdit them there, then just run: cheatsheet\n")
	fmt.Printf("Recolor the UI by editing %s\n", filepath.Join(dir, "theme.yaml"))
	return nil
}

// themeSample is a ready-to-edit theme.yaml seeded by --init. The values are the
// built-in defaults, so it changes nothing until edited; delete a line to fall
// back to that default.
const themeSample = `# cheatsheet colors — edit a value, or delete a line to keep the default.
# Each color is a hex string ("#A78BFA") or a 0–255 terminal color number.
colors:
  accent: "#A78BFA"        # headings, active border, search prompt
  accent_bright: "#C4B5FD" # section titles, footer keys
  keycap: "#22D3EE"        # the hotkeys themselves
  foreground: "#E5E7EB"    # descriptions and body text
  muted: "#6B7280"         # hints, counts, inactive text
  border: "#3F3F46"        # inactive pane borders, scrollbar track
  selection: "#312E81"     # highlighted row background
`

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
