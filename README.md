# ⌘ cheatsheet-tui

A fast, modern terminal cheat-sheet viewer. Define your hotkeys in simple YAML
files (one per program/app/type) and browse them in a searchable, vim-navigable
TUI that opens instantly.

![search across all cheatsheets, grouped by program and section](docs/preview.txt)

## Features

- **YAML cheatsheets** — one file per program/app/type (`vim`, `hyprland`,
  `system`, …), grouped into sections.
- **Instant fuzzy search** (`/`) across every cheatsheet at once.
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
mise trust && mise install   # one-time: fetch the pinned Go toolchain
mise exec -- go test ./...    # run the executable specifications
mise exec -- go build -o cheatsheet .
```

## Run

```sh
./cheatsheet                       # uses your cheatsheets, or the built-ins
./cheatsheet --dir ./cheatsheets   # load from an explicit directory
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

## Keys

| Key | Action |
| --- | --- |
| `j` / `k` | Move selection down / up |
| `h` / `l`, `tab` | Previous / next cheatsheet |
| `g` / `G` | Jump to top / bottom |
| `ctrl-d` / `ctrl-u` | Half-page down / up |
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
