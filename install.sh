#!/bin/sh
set -e

REPO="aporicho/gt"
BIN_NAME="gt"
INSTALL_DIR="${HOME}/.local/bin"

# ── 颜色与输出 ────────────────────────────────────────────────────────────────

if [ -t 1 ]; then
  BOLD='\033[1m' DIM='\033[2m'
  GREEN='\033[32m' CYAN='\033[36m' RED='\033[31m'
  RESET='\033[0m'
else
  BOLD='' DIM='' GREEN='' CYAN='' RED='' RESET=''
fi

info() { printf "  ${CYAN}→${RESET} %s\n" "$1"; }
ok()   { printf "  ${GREEN}✓${RESET} %s\n" "$1"; }
fail() { printf "  ${RED}✗ %s${RESET}\n" "$1"; exit 1; }

# ── 进度条 ────────────────────────────────────────────────────────────────────

draw_bar() {
  pct=$1 w=40
  filled=$((pct * w / 100))
  bar="" i=0
  while [ $i -lt $w ]; do
    if [ $i -lt $filled ]; then bar="${bar}█"; else bar="${bar}░"; fi
    i=$((i + 1))
  done
  printf "\r  ${CYAN}%s${RESET} %3d%%" "$bar" "$pct"
}

download() {
  url=$1 dest=$2

  total=$(curl -fsSLI "$url" 2>/dev/null \
    | grep -i '^content-length' | tail -1 | tr -dc '0-9')

  tmp=$(mktemp)
  trap 'rm -f "$tmp"' EXIT

  curl -fSL "$url" -o "$tmp" 2>/dev/null &
  pid=$!

  if [ -n "$total" ] && [ "$total" -gt 0 ] 2>/dev/null; then
    while kill -0 "$pid" 2>/dev/null; do
      current=$(wc -c < "$tmp" 2>/dev/null | tr -d ' ')
      [ -z "$current" ] && current=0
      pct=$((current * 100 / total))
      [ "$pct" -gt 100 ] && pct=100
      draw_bar "$pct"
      sleep 0.1
    done
    draw_bar 100
    printf "\n"
  else
    # fallback: spinner
    set -- '⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏'
    idx=1
    while kill -0 "$pid" 2>/dev/null; do
      eval c=\${$idx}
      printf "\r  %s 下载中..." "$c"
      idx=$((idx % $# + 1))
      sleep 0.08
    done
    printf "\r                    \n"
  fi

  wait "$pid" || { rm -f "$tmp"; fail "下载失败"; }
  mkdir -p "$(dirname "$dest")"
  mv "$tmp" "$dest"
  trap - EXIT
}

# ── 检测系统 ──────────────────────────────────────────────────────────────────

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)  ;;
  darwin) ;;
  *)      fail "不支持的系统: $OS" ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)               fail "不支持的架构: $ARCH" ;;
esac

# ── 获取版本 ──────────────────────────────────────────────────────────────────

TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | sed 's/.*"tag_name": *"\(.*\)".*/\1/')
[ -z "$TAG" ] && fail "无法获取最新版本"

BINARY="${BIN_NAME}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${TAG}/${BINARY}"

# ── 安装 ──────────────────────────────────────────────────────────────────────

printf "\n  ${BOLD}gt installer${RESET}\n\n"
info "系统: ${OS}/${ARCH}"
info "版本: ${TAG}"
printf "\n"

mkdir -p "$INSTALL_DIR"
download "$URL" "${INSTALL_DIR}/${BIN_NAME}"
chmod +x "${INSTALL_DIR}/${BIN_NAME}"
ok "安装到 ${INSTALL_DIR}/${BIN_NAME}"

ln -sf "${BIN_NAME}" "${INSTALL_DIR}/gtc"
ok "创建 gtc → gt symlink"

# ── 配置 shell 函数 ──────────────────────────────────────────────────────────

SHELL_NAME=$(basename "${SHELL:-/bin/sh}")
case "$SHELL_NAME" in
  zsh)  RC_FILE="$HOME/.zshrc" ;;
  bash) RC_FILE="$HOME/.bashrc" ;;
  *)    RC_FILE="" ;;
esac

if [ -z "$RC_FILE" ]; then
  printf "\n"
  printf "  未能识别 shell (%s)，请手动配置：\n" "$SHELL_NAME"
  echo '  # >>> gt >>>'
  echo '  export PATH="$HOME/.local/bin:$PATH"'
  echo '  __gt_cd() { local tmp="/tmp/gt_lastdir"; [ -f "$tmp" ] || return; local dir="$(cat "$tmp")"; rm -f "$tmp"; cd "$dir"; }'
  echo '  gt()  { if [ $# -eq 0 ]; then command gt  && __gt_cd; else command gt  "$@"; fi; }'
  echo '  gtc() { if [ $# -eq 0 ]; then command gtc && __gt_cd; else command gtc "$@"; fi; }'
  echo '  # <<< gt <<<'
  exit 0
fi

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
