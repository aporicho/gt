#!/bin/sh
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION_FILE="$ROOT/VERSION"

usage() {
    echo "用法: scripts/release.sh <patch|minor|major|x.y.z>"
    echo "示例: scripts/release.sh patch"
    exit 1
}

bump() {
    old=$1 type=$2
    major=$(echo "$old" | cut -d. -f1)
    minor=$(echo "$old" | cut -d. -f2)
    patch=$(echo "$old" | cut -d. -f3)
    case "$type" in
        major) echo "$((major+1)).0.0" ;;
        minor) echo "${major}.$((minor+1)).0" ;;
        patch) echo "${major}.${minor}.$((patch+1))" ;;
        *)     echo "$type" ;;
    esac
}

[ -z "$1" ] && usage

OLD=$(cat "$VERSION_FILE" | tr -d '[:space:]')
NEW=$(bump "$OLD" "$1")

echo "\n📦 发布 v${OLD} → v${NEW}\n"

printf '%s\n' "$NEW" > "$VERSION_FILE"

git -C "$ROOT" add VERSION
git -C "$ROOT" commit -m "release: v${NEW}"
git -C "$ROOT" tag "v${NEW}"
git -C "$ROOT" push origin main
git -C "$ROOT" push origin "v${NEW}"

echo "\n✅ v${NEW} 已发布！"
echo "   查看进度: gh run list --repo aporicho/gt"
