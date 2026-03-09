#!/bin/bash
# 本地编译 + 安装，方便开发测试
set -e
cd "$(dirname "$0")"
make build
install -m 755 gt ~/.local/bin/gt
ln -sf gt ~/.local/bin/gtc
echo "✅ 已安装到 ~/.local/bin/gt (gtc -> gt symlink)"
