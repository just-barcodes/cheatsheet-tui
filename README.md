# ⌘ cheatsheet-tui

A fast, modern terminal cheat-sheet viewer. Define your hotkeys in simple YAML
files (one per program/app/type) and browse them in a searchable, vim-navigable
TUI that opens instantly.

![search across all cheatsheets, grouped by program and section](docs/preview.txt)

## Features

- **YAML cheatsheets** — one file per program/app/type (`vim`, `hyprland`,
  `system`, …), grouped into sections.
- **Instant fuzzy search** (`/`) across every cheatsheet at once.
- **Multi-column layout** — hotkeys flow into side-by-side columns on wide
  terminals; press `c` to choose the column count yourself.
- **Vim navigation** — `j/k` move, `h/l` / `tab` switch sheets, `g/G` top/bottom,
  `ctrl-d/ctrl-u` half-page.
- **Modern look** — rounded panes, colour-coded keycaps, a highlighted
  selection bar. Built on Bubble Tea + Lip Gloss.

## Isolated, reproducible toolchain

Nothing is installed to the system. The Go toolchain is managed by
[mise](https://mise.jdx.dev/) and **all** Go caches/binaries are redirected into
the repo (`.go/`) by `.mise.toml`, so `go get`/`go install` never write to the
shared `~/go` or `~/.cache`.

```sh
mise trust && task setup      # one-time: pinned toolchain + git hooks
mise exec -- go test ./...    # run the executable specifications
mise exec -- go build -o cheatsheet .
```

Common commands are wrapped in a [Taskfile](https://taskfile.dev):

```sh
task            # list tasks
task run        # build and launch the TUI
task test       # run all tests
task ci         # fmt check + vet + test + build (same gate as GitHub CI)
```

GitHub Actions runs the same `task ci` on every push/PR, and pushing a
`v*` tag cross-compiles linux/darwin (amd64/arm64) binaries via
`task release:build` and attaches them to a GitHub release.

## Install

On Arch Linux (once published to the AUR — see `packaging/aur/README.md`):

```sh
yay -S cheatsheet-tui
```

With a Go toolchain (installs to `$GOBIN`, usually `~/go/bin` — note the
binary is named `cheatsheet-tui` after the module, not `cheatsheet`):

```sh
go install github.com/just-barcodes/cheatsheet-tui@latest
```

Or grab a binary from the GitHub releases page, or build from source as below.

## Run

```sh
./cheatsheet                       # uses your cheatsheets, or the built-ins
./cheatsheet --dir ./cheatsheets   # load from an explicit directory
./cheatsheet --theme ~/dark.yaml   # use a theme file from anywhere
```

## Defining your own cheatsheets

Cheatsheets are read from the **first** of these that applies:

1. `--dir <path>` flag
2. `$CHEATSHEET_DIR` environment variable
3. your config dir — `~/.config/cheatsheet` (`$XDG_CONFIG_HOME/cheatsheet`)
4. the **built-in** cheatsheets bundled into the binary

To start customising, scaffold the built-ins into your config dir and edit them:

```sh
./cheatsheet --init     # writes vim/hyprland/system .yaml to ~/.config/cheatsheet
                        # (never overwrites files you already have)
$EDITOR ~/.config/cheatsheet/vim.yaml
./cheatsheet            # now reads your edited copies
```

Add a new `*.yaml` file to that directory and it shows up in the sidebar,
sorted by `name`.

## Theming

Drop a `theme.yaml` in your config dir (`~/.config/cheatsheet`) to recolor the
UI — no code, no rebuild. `--init` seeds one with the defaults; edit a value, or
delete a line to keep that default. Each color is a hex string (`"#A78BFA"`) or
a `0`–`255` terminal color number.

Built-in themes (`--list-themes`): `default`, `selenized-dark/-light`,
`solarized-dark/-light`, `dracula`, `nord`, `gruvbox-dark/-light`, `tokyo-night`,
`catppuccin-mocha`. Use one directly, or as a starting point:

```sh
./cheatsheet --theme selenized-dark   # or -t; a name or a path to a theme.yaml
```

```yaml
# ~/.config/cheatsheet/theme.yaml — base on a preset, then tweak a few colors
name: selenized-dark
colors:
  keycap: "#ffffff"
```

Set `background` to paint the whole UI with the theme's color (the Selenized
presets do this, so they look right on any terminal); omit it to keep your
terminal's own background and transparency.

`--theme` takes either a built-in name or a path to a theme.yaml anywhere, and
overrides the config-dir default.

```yaml
colors:
  accent: "#A78BFA"        # headings, active border, search prompt
  accent_bright: "#C4B5FD" # section titles, footer keys
  keycap: "#22D3EE"        # the hotkeys themselves
  foreground: "#E5E7EB"    # descriptions and body text
  muted: "#6B7280"         # hints, counts, inactive text
  border: "#3F3F46"        # inactive pane borders, scrollbar track
  selection: "#312E81"     # highlighted row background
```

> Terminal apps can't change the font or its size — that's controlled by your
> terminal emulator — so theming covers colors only.

## Keys

| Key | Action |
| --- | --- |
| `j` / `k` | Move selection down / up |
| `h` / `l`, `tab` | Previous / next cheatsheet |
| `g` / `G` | Jump to top / bottom |
| `ctrl-d` / `ctrl-u` | Half-page down / up |
| `c` | Cycle hotkey columns (auto → 1 → 2 → 3); current setting shown in the footer |
| `/` | Search all hotkeys |
| `esc` | Exit search |
| `q`, `ctrl-c` | Quit |

## Cheatsheet format

```yaml
name: Vim
description: Modal text editor
icon: "📝"
sections:
  - title: Movement
    bindings:
      - keys: "h j k l"
        desc: "Move left / down / up / right"
      - keys: "dd / yy"
        desc: "Delete / yank line"
```

Drop a new `*.yaml` file into the cheatsheets directory and it appears in the
sidebar, sorted by `name`.

## Development

Behaviour is driven by executable specifications (RED → GREEN → REFACTOR):

- `features/*.feature` — Gherkin specs for loading and search, run by
  [godog](https://github.com/cucumber/godog) via `features_test.go`.
- `internal/tui/model_test.go` — TUI navigation/search behaviour.

```sh
mise exec -- go test ./...
```
