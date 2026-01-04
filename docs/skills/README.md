## Skills in this repo

Source of truth:
- `docs/skills/ksrc/SKILL.md`

## Install / publish

### Codex
Codex installs skills from GitHub paths via `$skill-installer`. Publish by keeping the skill folder in this repo and install with:
```
$skill-installer install https://github.com/respawn-app/ksrc/tree/main/docs/skills/ksrc
```

### Claude Code
Claude Code discovers skills from the filesystem. For personal installation:
```
mkdir -p ~/.claude/skills/ksrc
cp docs/skills/ksrc/SKILL.md ~/.claude/skills/ksrc/SKILL.md
```

For project installation, copy `SKILL.md` into `.claude/skills/ksrc/` in your own repo.

### Claude API / Claude.ai
Custom skills for Claude API are uploaded via the Skills API and referenced by `skill_id`. Claude.ai supports uploading custom skills as zip files in Settings > Features. These surfaces do not sync automatically; publish separately when needed.
