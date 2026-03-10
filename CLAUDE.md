# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**gt** (Go To) is a Go TUI utility for bookmarking directories. It uses Charmbracelet's Bubble Tea framework for the interactive terminal UI. Two commands share one binary:
- `gt`: pick a bookmark → cd
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

修改代码后运行 `./dev.sh` 一键编译并安装到 `~/.local/bin/gt`（含 gtc symlink + shell 函数更新），随后可直接在终端测试：

```bash
./dev.sh        # 编译 + 安装 gt + gtc symlink + 更新 shell 函数
gt              # 测试 picker（只 cd）
gtc             # 测试 picker（cd + claude）
gt add          # 测试交互式目录浏览器
gt add /tmp     # 测试直接添加路径
gt list         # 检查书签文件
```

依赖：Go（`brew install go`）。

## Architecture

All files are in `package main`, split by responsibility:

| File | Purpose |
|---|---|
| `main.go` | 入口 + CLI 分发 + `runApp` + `launchClaude` |
| `config.go` | 配置目录路径 (`~/.config/gt/`) + `init()` |
| `bookmarks.go` | 书签 CRUD (`loadBookmarks` / `saveBookmarks`) |
| `frecency.go` | frecency 数据结构 + 排序算法 |
| `theme.go` | 主题定义 + 加载/保存主题偏好 |
| `style.go` | lipgloss renderer + `themeStyles` 缓存 |
| `model.go` | `appModel` 结构体 + `Update`/`View` 分发 + `visibleRows` |
| `picker.go` | picker 模式 `updatePicker` + `viewPicker` |
| `browser.go` | browser 模式 `updateBrowser` + `viewBrowser` + 目录浏览 helpers |

The Bubble Tea model (`appModel` struct) drives two interactive modes: picker and browser.

**路径传递机制**：Go 二进制将选中路径写入 `/tmp/gt_lastdir`，shell 包装函数（`__gt_cd`）读取后执行 `cd` 并删除该文件。选中时会清屏（`\033[H\033[2J`），取消时保留终端内容。

**Shell 函数**：安装在 `~/.zshrc` 或 `~/.bashrc` 中，用 `# >>> gt >>>` / `# <<< gt <<<` 标记包裹，每次安装/开发时先删旧配置再写新配置。`install.sh` 和 `dev.sh` 各自内联同一份 shell 配置逻辑。

**argv[0] dispatch** (in `main()`):
- `filepath.Base(os.Args[0]) == "gtc"` → after picking, write path + launch claude
- otherwise → just write path (for shell `cd`)

**Non-interactive subcommands** (dispatched in `main()` before launching the TUI):
- `gt add [path]` — adds cwd or given path to bookmarks
- `gt list` — prints all bookmarks

**Interactive TUI** (launched by `runApp`):
- Pick a bookmark → write path to `/tmp/gt_lastdir` (for shell `cd`)
- If invoked as `gtc`, also launch `claude` in that directory

Bookmarks: plain text, one path per line, at `~/.config/gt/bookmarks`.
Theme preference: stored as a name string at `~/.config/gt/theme`.

## Key Details

- All UI text is in Simplified Chinese
- Navigation: arrow keys or hjkl, Enter to confirm, q/Esc/Ctrl+C to quit
- In-TUI keys: `-` delete (two-press confirmation), `+` add cwd, `t` cycle theme
- `appModel` is a **value type** — `Update()` returns a modified copy, not a pointer
- Lipgloss renders to `os.Stderr`; stdout is not used for path passing
- **不要用 `os.TempDir()`**——macOS 上它返回 `/var/folders/.../T/` 而非 `/tmp/`，与 shell 函数不一致

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
