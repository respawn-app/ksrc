# ksrc

## Benefit
Fast, one‑liner search and file read for Kotlin dependency sources without opening IDEs or vendoring sources. `ksrc` resolves dependency versions from your Gradle project, ensures source JARs are present in Gradle caches, and lets you search/read those sources directly.

## What it is
`ksrc` is a CLI that:
- Resolves project dependencies via Gradle (prefers project‑resolved versions).
- Locates/ensures source JARs in Gradle caches (no repo mutation).
- Runs `rg --search-zip` over source JARs and emits stable file‑ids.
- Supports `cat`/`open` by file‑id or path with line slicing.

## Install
We currently publish standalone binaries via GitHub Releases. Download the appropriate archive for your OS/arch and place `ksrc` on your `PATH`.

Install script: TBD

## Skills

### Claude Code plugin
Add the Respawn marketplace, then install the plugin:
```
/plugin marketplace add respawn-app/claude-plugin-marketplace
/plugin install ksrc@respawn-tools
```

### Codex skill
Install from the public GitHub path:
```
$skill-installer install https://github.com/respawn-app/ksrc/tree/main/skills/ksrc
```

## License
This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License.

See `LICENSE.txt`.
