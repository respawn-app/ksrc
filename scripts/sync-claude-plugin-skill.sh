#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
src="$root_dir/docs/skills/ksrc/SKILL.md"
dest="$root_dir/plugins/ksrc/skills/ksrc/SKILL.md"

if [[ ! -f "$src" ]]; then
  echo "Missing source skill: $src" >&2
  exit 1
fi

mkdir -p "$(dirname "$dest")"
cp "$src" "$dest"

echo "Synced ksrc skill to plugins/ksrc/skills/ksrc/SKILL.md"
