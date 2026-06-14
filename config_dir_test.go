package main

import (
	"path/filepath"
	"runtime"
	"testing"
)

// configDir must follow the XDG Base Directory spec on Linux: honor
// $XDG_CONFIG_HOME, and fall back to $HOME/.config when it is unset. These
// guard against anyone replacing os.UserConfigDir with a hardcoded ~/.config.
func TestConfigDirHonorsXDGConfigHome(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG semantics are Linux-specific; os.UserConfigDir differs elsewhere")
	}
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")
	if got, want := configDir(), filepath.Join("/tmp/xdg-config", "cheatsheet"); got != want {
		t.Fatalf("configDir() = %q, want %q", got, want)
	}
}

func TestConfigDirFallsBackToHomeDotConfig(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG semantics are Linux-specific; os.UserConfigDir differs elsewhere")
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/tmp/home-only")
	if got, want := configDir(), filepath.Join("/tmp/home-only", ".config", "cheatsheet"); got != want {
		t.Fatalf("configDir() = %q, want %q", got, want)
	}
}
