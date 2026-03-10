# gt — Go To

目录书签 TUI，快速跳转到常用项目。

- **`gt`** — 选择书签目录 → cd
- **`gtc`** — 选择书签目录 → cd → 自动启动 `claude`

两个命令共享同一个二进制，`gtc` 是指向 `gt` 的 symlink，通过 `argv[0]` 区分行为。

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/aporicho/gt/main/install.sh | sh
```

支持平台：macOS (Apple Silicon / Intel)、Linux (x64 / ARM64)，无需 Go 环境。

安装脚本会自动配置 shell 函数到 `~/.zshrc` 或 `~/.bashrc`，重复安装会自动替换旧配置。首次安装后请**开启新终端**生效。

使用 `gtc` 需要已安装 [Claude Code](https://docs.anthropic.com/en/docs/claude-code)。

## 用法

```
gt               选择书签目录，cd 过去
gtc              选择书签目录，cd 后自动启动 claude
gt add           交互式浏览并添加目录到书签
gt add <路径>    添加指定目录到书签
gt list          列出所有书签
gt update        更新到最新版本
gt --version     显示版本号
```

## 快捷键

| 按键 | 功能 |
|------|------|
| ↑ / k | 向上移动 |
| ↓ / j | 向下移动 |
| Enter | 确认选择 |
| - | 删除当前书签（需再按一次确认） |
| + | 添加当前目录到书签 |
| t | 切换主题 |
| q / Esc / Ctrl+C | 退出 |

## 书签存储

书签保存于 `~/.config/gt/bookmarks`，每行一个路径。
