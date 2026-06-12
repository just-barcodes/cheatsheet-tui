package config

// Source describes where cheatsheets should be read from.
type Source struct {
	Path    string // directory to load; empty when Builtin is true
	Builtin bool   // use the embedded built-in cheatsheets
}

// Locator decides where cheatsheets come from. All inputs are injected so the
// decision is pure and testable; main wires in the real flag/env/config dir.
type Locator struct {
	Flag      string            // value of the --dir flag ("" if unset)
	Env       string            // value of $CHEATSHEET_DIR ("" if unset)
	ConfigDir string            // e.g. ~/.config/cheatsheet ("" if unavailable)
	DirExists func(string) bool // reports whether a directory exists
}

// Resolve picks the cheatsheet source by priority: explicit flag, then env var,
// then the config directory if it exists, otherwise the built-in cheatsheets.
func (l Locator) Resolve() Source {
	switch {
	case l.Flag != "":
		return Source{Path: l.Flag}
	case l.Env != "":
		return Source{Path: l.Env}
	case l.ConfigDir != "" && l.DirExists != nil && l.DirExists(l.ConfigDir):
		return Source{Path: l.ConfigDir}
	default:
		return Source{Builtin: true}
	}
}
