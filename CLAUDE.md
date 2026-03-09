# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**gt** (Go To) is a single-file Go TUI utility for bookmarking directories. It uses Charmbracelet's Bubble Tea framework for the interactive terminal UI. Two commands share one binary:
- `gt`: pick a bookmark → cd (no claude)
- `gtc`: pick a bookmark → cd → launch `claude`

`gtc` is a symlink to `gt`; behavior is determined by `filepath.Base(os.Args[0])`.

## Commands

```bash
make build      # 编译当前平台
make install    # 编译并安装到 ~/.local/bin + 创建 gtc symlink
make fmt        # 格式化
make vet        # 格式化 + lint
make build-all  # 交叉编译所有平台到 dist/
```

## 开发流程

修改代码后运行 `./dev.sh` 一键编译并安装到 `~/.local/bin/gt`（含 gtc symlink），随后可直接在终端测试：

```bash
./dev.sh        # 编译 + 安装 gt + gtc symlink
gt              # 测试 picker（只 cd）
gtc             # 测试 picker（cd + claude）
gt add          # 测试交互式目录浏览器
gt add /tmp     # 测试直接添加路径
gt list         # 检查书签文件
```

依赖：Go（`brew install go`）。

## Architecture

Everything lives in `main.go`. The Bubble Tea model (`appModel` struct) drives a single interactive picker.

**argv[0] dispatch** (in `main()`):
- `filepath.Base(os.Args[0]) == "gtc"` → after picking, print path + launch claude
- otherwise → just print path (for shell `cd`)

**Non-interactive subcommands** (dispatched in `main()` before launching the TUI):
- `gt add [path]` — adds cwd or given path to bookmarks
- `gt list` — prints all bookmarks

**Interactive TUI** (launched by `runApp`):
- Pick a bookmark → print path to stdout (for shell `cd`)
- If invoked as `gtc`, also launch `claude` in that directory

Bookmarks: plain text, one path per line, at `~/.config/gt/bookmarks`.
Theme preference: stored as a name string at `~/.config/gt/theme`.

## Key Details

- All UI text is in Simplified Chinese
- Navigation: arrow keys or hjkl, Enter to confirm, q/Esc/Ctrl+C to quit
- In-TUI keys: `-` delete (two-press confirmation), `+` add cwd, `t` cycle theme
- `appModel` is a **value type** — `Update()` returns a modified copy, not a pointer
- Lipgloss renders to `os.Stderr`; the selected path is printed to `os.Stdout`

## Model Fields Reference

| Field | Purpose |
|---|---|
| `dirs` | ordered bookmark list |
| `pickerCursor` | highlighted index |
| `selected` | set on Enter; empty string means cancelled |
| `width`, `height` | terminal size, updated from `tea.WindowSizeMsg` |
| `themeIdx` | index into the `themes` slice |
| `cachedStyles` | pre-built `themeStyles`; rebuilt only at init and when `t` is pressed |
| `pickerOffset` | first visible index for the scroll viewport |
| `statusMsg` | transient feedback line, cleared on every keypress |
| `confirmDelete` | two-step `-` confirmation flag |

## Style Caching

`buildStyles(themeIdx int) themeStyles` constructs the six lipgloss styles once and returns a `themeStyles` struct. It is called at model creation inside `runApp` and again in `Update()` when the `"t"` key is pressed — never inside `View()`.

## Scrolling Viewport

Visible rows = `height - 3` (title row + status bar + margin). After every update, `pickerOffset` is clamped so the cursor stays in view. `View()` renders only `m.dirs[m.pickerOffset:end]`.
