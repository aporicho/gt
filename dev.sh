#!/bin/bash
# 本地编译 + 安装，方便开发测试
set -e
cd "$(dirname "$0")"
make build
install -m 755 gt ~/.local/bin/gt
ln -sf gt ~/.local/bin/gtc

# ── 配置 shell 函数 ──────────────────────────────────────────────────────────

SHELL_NAME=$(basename "${SHELL:-/bin/sh}")
case "$SHELL_NAME" in
  zsh)  RC_FILE="$HOME/.zshrc" ;;
  bash) RC_FILE="$HOME/.bashrc" ;;
  *)    RC_FILE="" ;;
esac

if [ -n "$RC_FILE" ]; then
  sed -i.bak '/# >>> gt >>>/,/# <<< gt <<</d' "$RC_FILE" && rm -f "${RC_FILE}.bak"
  cat >> "$RC_FILE" << 'BLOCK'
# >>> gt >>>
export PATH="$HOME/.local/bin:$PATH"
__gt_cd() {
    local tmp="/tmp/gt_lastdir"
    [ -f "$tmp" ] && cd "$(cat "$tmp")" && rm -f "$tmp"
}
gt() {
    if [ $# -eq 0 ]; then command gt && __gt_cd; else command gt "$@"; fi
}
gtc() {
    if [ $# -eq 0 ]; then command gtc && __gt_cd; else command gtc "$@"; fi
}
# <<< gt <<<
BLOCK
fi

echo "✅ 已安装到 ~/.local/bin/gt (gtc -> gt symlink)"
echo "运行 source ${RC_FILE} 生效"
