package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/just-barcodes/cheatsheet-tui/internal/config"
	"github.com/just-barcodes/cheatsheet-tui/internal/search"
	"github.com/just-barcodes/cheatsheet-tui/internal/tui"
)

// featureState holds the world shared between steps in a scenario.
type featureState struct {
	dir      string
	lastFile string
	sheet    config.Cheatsheet
	sheets   []config.Cheatsheet
	loadErr  error

	items   []search.Item
	results []search.Item

	loc      config.Locator
	existing map[string]bool
	resolved config.Source

	model    tui.Model
	rendered string
}

// --- loading steps ---

func (s *featureState) aCheatsheetFileWithContent(name string, content *godog.DocString) error {
	s.dir = s.tempDir()
	s.lastFile = filepath.Join(s.dir, name)
	return os.WriteFile(s.lastFile, []byte(content.Content), 0o644)
}

func (s *featureState) iLoadThatCheatsheet() error {
	s.sheet, s.loadErr = config.LoadFile(s.lastFile)
	return nil
}

func (s *featureState) theCheatsheetNameIs(want string) error {
	if s.loadErr != nil {
		return s.loadErr
	}
	if s.sheet.Name != want {
		return fmt.Errorf("name = %q, want %q", s.sheet.Name, want)
	}
	return nil
}

func (s *featureState) itHasNSection(n int) error {
	if got := len(s.sheet.Sections); got != n {
		return fmt.Errorf("sections = %d, want %d", got, n)
	}
	return nil
}

func (s *featureState) sectionHasNBindings(title string, n int) error {
	for _, sec := range s.sheet.Sections {
		if sec.Title == title {
			if got := len(sec.Bindings); got != n {
				return fmt.Errorf("section %q bindings = %d, want %d", title, got, n)
			}
			return nil
		}
	}
	return fmt.Errorf("section %q not found", title)
}

func (s *featureState) bindingHasDescription(keys, desc string) error {
	for _, sec := range s.sheet.Sections {
		for _, b := range sec.Bindings {
			if b.Keys == keys {
				if b.Desc != desc {
					return fmt.Errorf("binding %q desc = %q, want %q", keys, b.Desc, desc)
				}
				return nil
			}
		}
	}
	return fmt.Errorf("binding %q not found", keys)
}

func (s *featureState) loadingFailsWithAnError() error {
	if s.loadErr == nil {
		return fmt.Errorf("expected a load error, got nil")
	}
	return nil
}

func (s *featureState) aDirectoryWithCheatsheets(table *godog.Table) error {
	s.dir = s.tempDir()
	for _, row := range table.Rows[1:] {
		file := row.Cells[0].Value
		name := row.Cells[1].Value
		body := fmt.Sprintf("name: %s\nsections: []\n", name)
		if err := os.WriteFile(filepath.Join(s.dir, file), []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (s *featureState) iLoadTheDirectory() error {
	s.sheets, s.loadErr = config.LoadDir(s.dir)
	return s.loadErr
}

func (s *featureState) iGetNCheatsheets(n int) error {
	if got := len(s.sheets); got != n {
		return fmt.Errorf("cheatsheets = %d, want %d", got, n)
	}
	return nil
}

func (s *featureState) theCheatsheetsAreOrdered(order string) error {
	var got []string
	for _, c := range s.sheets {
		got = append(got, c.Name)
	}
	want := splitList(order)
	if strings.Join(got, ", ") != strings.Join(want, ", ") {
		return fmt.Errorf("order = %v, want %v", got, want)
	}
	return nil
}

// --- search steps ---

func (s *featureState) theFollowingHotkeys(table *godog.Table) error {
	head := table.Rows[0].Cells
	col := map[string]int{}
	for i, c := range head {
		col[c.Value] = i
	}
	s.items = nil
	for _, row := range table.Rows[1:] {
		s.items = append(s.items, search.Item{
			Sheet:   row.Cells[col["sheet"]].Value,
			Section: row.Cells[col["section"]].Value,
			Keys:    row.Cells[col["keys"]].Value,
			Desc:    row.Cells[col["desc"]].Value,
		})
	}
	return nil
}

func (s *featureState) iSearchFor(query string) error {
	s.results = search.Filter(s.items, query)
	return nil
}

func (s *featureState) theResultsContainKey(keys string) error {
	if !s.hasKey(keys) {
		return fmt.Errorf("expected results to contain key %q; got %v", keys, s.resultKeys())
	}
	return nil
}

func (s *featureState) theResultsDoNotContainKey(keys string) error {
	if s.hasKey(keys) {
		return fmt.Errorf("expected results NOT to contain key %q; got %v", keys, s.resultKeys())
	}
	return nil
}

func (s *featureState) iGetNResults(n int) error {
	if got := len(s.results); got != n {
		return fmt.Errorf("results = %d, want %d", got, n)
	}
	return nil
}

func (s *featureState) resultNHasKey(n int, keys string) error {
	if n < 1 || n > len(s.results) {
		return fmt.Errorf("result %d out of range (%d results)", n, len(s.results))
	}
	if got := s.results[n-1].Keys; got != keys {
		return fmt.Errorf("result %d key = %q, want %q", n, got, keys)
	}
	return nil
}

// --- location/resolution steps ---

func (s *featureState) theDirFlagIs(path string) error {
	s.loc.Flag = path
	return nil
}

func (s *featureState) theEnvVarIs(path string) error {
	s.loc.Env = path
	return nil
}

func (s *featureState) aConfigDirThatExists(path string) error {
	s.loc.ConfigDir = path
	if s.existing == nil {
		s.existing = map[string]bool{}
	}
	s.existing[path] = true
	return nil
}

func (s *featureState) aConfigDirThatDoesNotExist(path string) error {
	s.loc.ConfigDir = path
	return nil
}

func (s *featureState) iResolveTheLocation() error {
	s.loc.DirExists = func(p string) bool { return s.existing[p] }
	s.resolved = s.loc.Resolve()
	return nil
}

func (s *featureState) cheatsheetsLoadFrom(path string) error {
	if s.resolved.Builtin {
		return fmt.Errorf("resolved to built-in, want path %q", path)
	}
	if s.resolved.Path != path {
		return fmt.Errorf("resolved path = %q, want %q", s.resolved.Path, path)
	}
	return nil
}

func (s *featureState) theBuiltinCheatsheetsAreUsed() error {
	if !s.resolved.Builtin {
		return fmt.Errorf("expected built-in cheatsheets, got path %q", s.resolved.Path)
	}
	return nil
}

// --- helpers ---

func (s *featureState) hasKey(keys string) bool {
	for _, r := range s.results {
		if r.Keys == keys {
			return true
		}
	}
	return false
}

func (s *featureState) resultKeys() []string {
	var ks []string
	for _, r := range s.results {
		ks = append(ks, r.Keys)
	}
	return ks
}

func (s *featureState) tempDir() string {
	d, err := os.MkdirTemp("", "cheatsheet-bdd")
	if err != nil {
		panic(err)
	}
	return d
}

func splitList(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	s := &featureState{}

	ctx.Step(`^a cheatsheet file "([^"]*)" with content:$`, s.aCheatsheetFileWithContent)
	ctx.Step(`^I load that cheatsheet$`, s.iLoadThatCheatsheet)
	ctx.Step(`^the cheatsheet name is "([^"]*)"$`, s.theCheatsheetNameIs)
	ctx.Step(`^it has (\d+) section$`, s.itHasNSection)
	ctx.Step(`^section "([^"]*)" has (\d+) bindings$`, s.sectionHasNBindings)
	ctx.Step(`^binding "([^"]*)" has description "([^"]*)"$`, s.bindingHasDescription)
	ctx.Step(`^loading fails with an error$`, s.loadingFailsWithAnError)

	ctx.Step(`^a directory with cheatsheets:$`, s.aDirectoryWithCheatsheets)
	ctx.Step(`^I load the directory$`, s.iLoadTheDirectory)
	ctx.Step(`^I get (\d+) cheatsheets$`, s.iGetNCheatsheets)
	ctx.Step(`^the cheatsheets are ordered "([^"]*)"$`, s.theCheatsheetsAreOrdered)

	ctx.Step(`^the following hotkeys:$`, s.theFollowingHotkeys)
	ctx.Step(`^I search for "([^"]*)"$`, s.iSearchFor)
	ctx.Step(`^the results contain key "([^"]*)"$`, s.theResultsContainKey)
	ctx.Step(`^the results do not contain key "([^"]*)"$`, s.theResultsDoNotContainKey)
	ctx.Step(`^I get (\d+) results$`, s.iGetNResults)
	ctx.Step(`^result (\d+) has key "([^"]*)"$`, s.resultNHasKey)

	ctx.Step(`^a cheatsheet "([^"]*)" with (\d+) bindings$`, s.aCheatsheetWithBindings)
	ctx.Step(`^a cheatsheet "([^"]*)" with a long description ending in "([^"]*)"$`, s.aCheatsheetWithLongDescription)
	ctx.Step(`^a terminal (\d+) columns wide and (\d+) rows tall$`, s.aTerminalSized)
	ctx.Step(`^I view the cheatsheet$`, s.iViewTheCheatsheet)
	ctx.Step(`^the hotkeys are laid out in (\d+) columns?$`, s.theHotkeysAreLaidOutInColumns)
	ctx.Step(`^binding "([^"]*)" is visible$`, s.bindingIsVisible)
	ctx.Step(`^binding "([^"]*)" is not visible$`, s.bindingIsNotVisible)
	ctx.Step(`^the screen shows the column count "([^"]*)"$`, s.theScreenShowsColumnCount)
	ctx.Step(`^the screen asks for a larger window$`, s.theScreenAsksForALargerWindow)
	ctx.Step(`^I set the column count to (\d+)$`, s.iSetTheColumnCountTo)
	ctx.Step(`^I cycle the column count (\d+) times$`, s.iCycleTheColumnCount)

	ctx.Step(`^the --dir flag is "([^"]*)"$`, s.theDirFlagIs)
	ctx.Step(`^the CHEATSHEET_DIR env var is "([^"]*)"$`, s.theEnvVarIs)
	ctx.Step(`^a config directory "([^"]*)" that exists$`, s.aConfigDirThatExists)
	ctx.Step(`^a config directory "([^"]*)" that does not exist$`, s.aConfigDirThatDoesNotExist)
	ctx.Step(`^I resolve the cheatsheet location$`, s.iResolveTheLocation)
	ctx.Step(`^cheatsheets load from "([^"]*)"$`, s.cheatsheetsLoadFrom)
	ctx.Step(`^the built-in cheatsheets are used$`, s.theBuiltinCheatsheetsAreUsed)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("feature tests failed")
	}
}
