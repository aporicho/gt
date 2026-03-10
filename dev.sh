#!/bin/bash
# 本地编译 + 安装，方便开发测试
set -e
cd "$(dirname "$0")"

# ── 颜色与输出 ────────────────────────────────────────────────────────────────

if [ -t 1 ]; then
  BOLD='\033[1m' GREEN='\033[32m' CYAN='\033[36m' RED='\033[31m' RESET='\033[0m'
else
  BOLD='' GREEN='' CYAN='' RED='' RESET=''
fi

info() { printf "  ${CYAN}→${RESET} %s\n" "$1"; }
ok()   { printf "  ${GREEN}✓${RESET} %s\n" "$1"; }
fail() { printf "  ${RED}✗ %s${RESET}\n" "$1"; exit 1; }

spin() {
  msg=$1; pid=$2
  set -- '⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏'
  idx=1
  while kill -0 "$pid" 2>/dev/null; do
    eval c=\${$idx}
    printf "\r  ${CYAN}%s${RESET} %s" "$c" "$msg"
    idx=$((idx % $# + 1))
    sleep 0.08
  done
  wait "$pid" 2>/dev/null && printf "\r  ${GREEN}✓${RESET} %s\n" "$msg" \
                          || { printf "\r"; fail "$msg"; }
}

# ── 编译 ──────────────────────────────────────────────────────────────────────

printf "\n  ${BOLD}gt dev build${RESET}\n\n"

make build > /dev/null 2>&1 &
spin "编译中..." $!


install -m 755 gt ~/.local/bin/gt
ok "安装到 ~/.local/bin/gt"

ln -sf gt ~/.local/bin/gtc
ok "创建 gtc → gt symlink"

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
    [ -f "$tmp" ] || return
    local dir="$(cat "$tmp")"
    rm -f "$tmp"
    cd "$dir"
}
gt() {
    if [ $# -eq 0 ]; then command gt && __gt_cd; else command gt "$@"; fi
}
gtc() {
    if [ $# -eq 0 ]; then command gtc && __gt_cd; else command gtc "$@"; fi
}
# <<< gt <<<
BLOCK
  ok "更新 shell 配置 → ${RC_FILE}"
  printf "\n  运行 ${BOLD}source ${RC_FILE}${RESET} 生效\n\n"
fi
