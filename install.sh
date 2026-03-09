#!/bin/sh
set -e

REPO="aporicho/gtc"
BIN_NAME="gtc"
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

echo "已安装到 ${INSTALL_DIR}/${BIN_NAME}"

# ── Detect shell config file ─────────────────────────────────────────────────

SHELL_NAME=$(basename "${SHELL:-/bin/sh}")
case "$SHELL_NAME" in
  zsh)  RC_FILE="$HOME/.zshrc" ;;
  bash) RC_FILE="$HOME/.bashrc" ;;
  *)    RC_FILE="" ;;
esac

if [ -z "$RC_FILE" ]; then
  echo ""
  echo "未能识别 shell ($SHELL_NAME)，请手动配置："
  echo '  export PATH="$HOME/.local/bin:$PATH"'
  echo '  gtc() { if [ $# -eq 0 ]; then local dir=$(command gtc); [ -n "$dir" ] && cd "$dir"; else command gtc "$@"; fi; }'
  exit 0
fi

CHANGED=false

# ── Add PATH if needed ───────────────────────────────────────────────────────

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo '' >> "$RC_FILE"
    echo '# gtc' >> "$RC_FILE"
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$RC_FILE"
    CHANGED=true
    echo "已添加 PATH 到 ${RC_FILE}"
    ;;
esac

# ── Add shell wrapper function ────────────────────────────────────────────────

if grep -q 'command gtc' "$RC_FILE" 2>/dev/null; then
  echo "shell 函数已存在，跳过"
else
  # if PATH was already there, we still need the comment header
  if [ "$CHANGED" = false ]; then
    echo '' >> "$RC_FILE"
    echo '# gtc' >> "$RC_FILE"
  fi
  cat >> "$RC_FILE" << 'FUNC'
gtc() {
    if [ $# -eq 0 ]; then
        local dir=$(command gtc)
        [ -n "$dir" ] && cd "$dir"
    else
        command gtc "$@"
    fi
}
FUNC
  CHANGED=true
  echo "已添加 shell 函数到 ${RC_FILE}"
fi

if [ "$CHANGED" = true ]; then
  echo ""
  echo "请运行以下命令使配置生效："
  echo "  source ${RC_FILE}"
fi
