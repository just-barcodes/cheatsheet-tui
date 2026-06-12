package main

import (
	"testing"

	"github.com/just-barcodes/cheatsheet-tui/internal/config"
)

// The built-in cheatsheets ship inside the binary, so a malformed YAML file
// in cheatsheets/ would only surface at runtime. Guard them at test time.
func TestBuiltinCheatsheetsAreValid(t *testing.T) {
	sheets, err := loadSource(config.Source{Builtin: true})
	if err != nil {
		t.Fatalf("built-in cheatsheets failed to load: %v", err)
	}
	if len(sheets) == 0 {
		t.Fatal("no built-in cheatsheets embedded")
	}
	for _, s := range sheets {
		if s.Name == "" {
			t.Errorf("a built-in cheatsheet has no name")
		}
		if len(s.Sections) == 0 {
			t.Errorf("built-in cheatsheet %q has no sections", s.Name)
		}
	}
}
