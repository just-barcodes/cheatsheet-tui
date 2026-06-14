package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"

	"github.com/just-barcodes/cheatsheet-tui/internal/config"
)

// --- theme steps ---

func (s *featureState) aThemeFileWithContent(content *godog.DocString) error {
	s.dir = s.tempDir()
	s.themePath = filepath.Join(s.dir, "theme.yaml")
	return os.WriteFile(s.themePath, []byte(content.Content), 0o644)
}

func (s *featureState) noThemeFileExists() error {
	s.dir = s.tempDir()
	s.themePath = filepath.Join(s.dir, "theme.yaml") // never created
	return nil
}

func (s *featureState) iLoadThatTheme() error {
	s.theme, s.themeErr = config.LoadThemeFile(s.themePath)
	return nil
}

func (s *featureState) theAccentColorIs(want string) error {
	if s.themeErr != nil {
		return s.themeErr
	}
	if got := s.theme.Colors.Accent; got != want {
		return fmt.Errorf("accent = %q, want %q", got, want)
	}
	return nil
}

func (s *featureState) theKeycapColorIs(want string) error {
	if s.themeErr != nil {
		return s.themeErr
	}
	if got := s.theme.Colors.Keycap; got != want {
		return fmt.Errorf("keycap = %q, want %q", got, want)
	}
	return nil
}

func (s *featureState) theAccentColorIsUnset() error {
	if s.themeErr != nil {
		return s.themeErr
	}
	if got := s.theme.Colors.Accent; got != "" {
		return fmt.Errorf("accent = %q, want unset", got)
	}
	return nil
}

func (s *featureState) theForegroundColorIsUnset() error {
	if s.themeErr != nil {
		return s.themeErr
	}
	if got := s.theme.Colors.Foreground; got != "" {
		return fmt.Errorf("foreground = %q, want unset", got)
	}
	return nil
}

func (s *featureState) loadingTheThemeFails() error {
	if s.themeErr == nil {
		return fmt.Errorf("expected a theme load error, got nil")
	}
	return nil
}

// --- theme source resolution steps ---

func (s *featureState) theThemeFlagIs(path string) error {
	s.themeLoc.Flag = path
	return nil
}

func (s *featureState) aConfigDirForTheTheme(path string) error {
	s.themeLoc.ConfigDir = path
	return nil
}

func (s *featureState) iResolveTheThemeSource() error {
	s.themeSrc = s.themeLoc.Resolve()
	return nil
}

func (s *featureState) theThemeLoadsFrom(path string) error {
	if s.themeSrc.Path != path {
		return fmt.Errorf("theme path = %q, want %q", s.themeSrc.Path, path)
	}
	return nil
}

func (s *featureState) theThemeFileIsRequiredToExist() error {
	if !s.themeSrc.MustExist {
		return fmt.Errorf("expected the theme file to be required to exist")
	}
	return nil
}

func (s *featureState) aMissingThemeFileIsAllowed() error {
	if s.themeSrc.MustExist {
		return fmt.Errorf("expected a missing theme file to be allowed")
	}
	return nil
}

func (s *featureState) noThemeFileIsLoaded() error {
	if s.themeSrc.Path != "" {
		return fmt.Errorf("expected no theme file, got %q", s.themeSrc.Path)
	}
	return nil
}
