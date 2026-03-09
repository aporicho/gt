#!/bin/sh
set -e

REPO="aporicho/gt"
BIN_NAME="gt"
INSTALL_DIR="${HOME}/.local/bin"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *)
    echo "不支持的系统: $OS"
    exit 1
    ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)
    echo "不支持的架构: $ARCH"
    exit 1
    ;;
esac

# Get latest release tag
TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | sed 's/.*"tag_name": *"\(.*\)".*/\1/')

if [ -z "$TAG" ]; then
  echo "无法获取最新版本"
  exit 1
fi

BINARY="${BIN_NAME}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${TAG}/${BINARY}"

echo "正在安装 ${BIN_NAME} ${TAG} (${OS}/${ARCH})..."

mkdir -p "$INSTALL_DIR"
curl -fsSL "$URL" -o "${INSTALL_DIR}/${BIN_NAME}"
chmod +x "${INSTALL_DIR}/${BIN_NAME}"
ln -sf "${BIN_NAME}" "${INSTALL_DIR}/gtc"

echo "已安装到 ${INSTALL_DIR}/${BIN_NAME} (gtc -> gt symlink)"

# ── 配置 shell 函数 ──────────────────────────────────────────────────────────

SHELL_NAME=$(basename "${SHELL:-/bin/sh}")
case "$SHELL_NAME" in
  zsh)  RC_FILE="$HOME/.zshrc" ;;
  bash) RC_FILE="$HOME/.bashrc" ;;
  *)    RC_FILE="" ;;
esac

if [ -z "$RC_FILE" ]; then
  echo ""
  echo "未能识别 shell ($SHELL_NAME)，请手动配置："
  echo '  # >>> gt >>>'
  echo '  export PATH="$HOME/.local/bin:$PATH"'
  echo '  __gt_cd() { local tmp="/tmp/gt_lastdir"; [ -f "$tmp" ] && cd "$(cat "$tmp")" && rm -f "$tmp"; }'
  echo '  gt()  { if [ $# -eq 0 ]; then command gt  && __gt_cd; else command gt  "$@"; fi; }'
  echo '  gtc() { if [ $# -eq 0 ]; then command gtc && __gt_cd; else command gtc "$@"; fi; }'
  echo '  # <<< gt <<<'
  exit 0
fi

# 删除旧配置
sed -i.bak '/# >>> gt >>>/,/# <<< gt <<</d' "$RC_FILE" && rm -f "${RC_FILE}.bak"

# 写入新配置
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

echo "已更新 shell 配置 → ${RC_FILE}"
echo "运行 source ${RC_FILE} 生效"
