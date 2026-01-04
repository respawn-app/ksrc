## Skills in this repo

Source of truth:
- `docs/skills/ksrc/SKILL.md`

## Install / publish

### Codex
Codex installs skills from GitHub paths via `$skill-installer`. Publish by keeping the skill folder in this repo and install with:
```
$skill-installer install https://github.com/respawn-app/ksrc/tree/main/docs/skills/ksrc
```

### Claude Code (plugin marketplace)
This repo hosts a Claude Code plugin marketplace so anyone can install the `ksrc` skill plugin:
```
/plugin marketplace add respawn-app/claude-plugin-marketplace
/plugin install ksrc@respawn-tools
```

Marketplace repo: `respawn-app/claude-plugin-marketplace`
Plugin source: `respawn-app/ksrc` (via marketplace `source` -> `github`).

Keep the plugin skill in sync with the source of truth:
```
scripts/sync-claude-plugin-skill.sh
```

### Claude Code (manual install)
Claude Code also discovers skills from the filesystem. For personal installation:
```
mkdir -p ~/.claude/skills/ksrc
cp docs/skills/ksrc/SKILL.md ~/.claude/skills/ksrc/SKILL.md
```

For project installation, copy `SKILL.md` into `.claude/skills/ksrc/` in your own repo.

### Claude API / Claude.ai
Custom skills for Claude API are uploaded via the Skills API and referenced by `skill_id`. Claude.ai supports uploading custom skills as zip files in Settings > Features. These surfaces do not sync automatically; publish separately when needed.
