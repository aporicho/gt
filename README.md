# gtc

目录书签管理工具，用于快速跳转到常用目录。

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/aporicho/gtc/main/install.sh | bash
```

支持平台：macOS (Apple Silicon / Intel)、Linux (x64 / ARM64)，无需 Go 环境。

在 `~/.zshrc` 或 `~/.bashrc` 中添加 shell 函数：

```sh
gtc() {
    if [ $# -eq 0 ]; then
        local dir=$(command gtc)
        [ -n "$dir" ] && cd "$dir"
    else
        command gtc "$@"
    fi
}
```

## 用法

```
gtc              交互式选择书签目录并跳转
gtc add          添加当前目录到书签
gtc add <路径>   添加指定目录到书签
gtc list         列出所有书签
```

## 操作

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

书签保存于 `~/.config/gtc/bookmarks`，每行一个路径。
