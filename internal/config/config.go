// Package config loads cheatsheets from simple YAML files.
package config

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Binding is a single hotkey and what it does.
type Binding struct {
	Keys string `yaml:"keys"`
	Desc string `yaml:"desc"`
}

// Section groups related bindings under a heading (e.g. "Movement").
type Section struct {
	Title    string    `yaml:"title"`
	Bindings []Binding `yaml:"bindings"`
}

// Cheatsheet is the set of hotkeys for one program/app/type.
type Cheatsheet struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Icon        string    `yaml:"icon"`
	Sections    []Section `yaml:"sections"`
}

// Load parses a single cheatsheet from r.
func Load(r io.Reader) (Cheatsheet, error) {
	var c Cheatsheet
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		return Cheatsheet{}, err
	}
	return c, nil
}

// LoadFile parses a single cheatsheet YAML file.
func LoadFile(path string) (Cheatsheet, error) {
	f, err := os.Open(path)
	if err != nil {
		return Cheatsheet{}, err
	}
	defer f.Close()
	c, err := Load(f)
	if err != nil {
		return Cheatsheet{}, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	return c, nil
}

// LoadDir parses every *.yaml/*.yml file in dir, sorted by sheet name.
func LoadDir(dir string) ([]Cheatsheet, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var sheets []Cheatsheet
	for _, e := range entries {
		if e.IsDir() || !isYAML(e.Name()) {
			continue
		}
		c, err := LoadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		sheets = append(sheets, c)
	}
	sortByName(sheets)
	return sheets, nil
}

// LoadFS parses every *.yaml/*.yml file at the root of fsys, sorted by sheet
// name. Used to read the embedded built-in cheatsheets.
func LoadFS(fsys fs.FS) ([]Cheatsheet, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	var sheets []Cheatsheet
	for _, e := range entries {
		if e.IsDir() || !isYAML(e.Name()) {
			continue
		}
		data, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, err
		}
		c, err := Load(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		sheets = append(sheets, c)
	}
	sortByName(sheets)
	return sheets, nil
}

func sortByName(sheets []Cheatsheet) {
	sort.SliceStable(sheets, func(i, j int) bool {
		return sheets[i].Name < sheets[j].Name
	})
}

func isYAML(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".yaml" || ext == ".yml"
}
