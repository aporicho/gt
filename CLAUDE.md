# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**GTC** (Go Terminal Changer) is a single-file Go TUI utility for bookmarking and navigating to directories. It uses Charmbracelet's Bubble Tea framework for the interactive terminal UI.

## Commands

```bash
make build      # 编译当前平台
make install    # 编译并安装到 ~/.local/bin  (uses install(1) — atomic, safe while binary is running)
make fmt        # 格式化
make vet        # 格式化 + lint
make build-all  # 交叉编译所有平台到 dist/
```

## Architecture

Everything lives in `main.go`. The Bubble Tea model (`model` struct) drives a single interactive picker.

**Non-interactive subcommands** (dispatched in `main()` before launching the TUI):
- `gtc add [path]` — adds cwd or given path to bookmarks
- `gtc list` — prints all bookmarks

**Interactive TUI** (launched by `runPicker`):
- Default (`gtc`): pick a bookmark; selected path is printed to stdout for use with `cd $(gtc)`
- `gtc rm` was **removed** — use the `-` key inside the TUI instead

Bookmarks: plain text, one path per line, at `~/.config/gtc/bookmarks`.
Theme preference: stored as a name string at `~/.config/gtc/theme`.

## Key Details

- All UI text is in Simplified Chinese
- Navigation: arrow keys or hjkl, Enter to confirm, q/Esc/Ctrl+C to quit
- In-TUI keys: `-` delete (two-press confirmation), `+` add cwd, `t` cycle theme
- `model` is a **value type** — `Update()` returns a modified copy, not a pointer
- Lipgloss renders to `os.Stderr`; the selected path is printed to `os.Stdout`

## Model Fields Reference

| Field | Purpose |
|---|---|
| `dirs` | ordered bookmark list |
| `cursor` | highlighted index |
| `selected` | set on Enter; empty string means cancelled |
| `width`, `height` | terminal size, updated from `tea.WindowSizeMsg` |
| `themeIdx` | index into the `themes` slice |
| `cachedStyles` | pre-built `themeStyles`; rebuilt only at init and when `t` is pressed |
| `listOffset` | first visible index for the scroll viewport |
| `statusMsg` | transient feedback line, cleared on every keypress |
| `confirmDelete` | two-step `-` confirmation flag |

## Style Caching

`buildStyles(themeIdx int) themeStyles` constructs the six lipgloss styles once and returns a `themeStyles` struct. It is called at model creation inside `runPicker` and again in `Update()` when the `"t"` key is pressed — never inside `View()`.

## Scrolling Viewport

Visible rows = `height - 3` (title row + status bar + margin). After every update, `listOffset` is clamped so the cursor stays in view. `View()` renders only `m.dirs[m.listOffset:end]`.

## Note on README

`README.md` still documents `gtc rm` — update it if publishing.
