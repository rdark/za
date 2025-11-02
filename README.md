# Za - Zettelkasten Augmentation

[![CI](https://github.com/rdark/za/workflows/CI/badge.svg)](https://github.com/rdark/za/actions?query=workflow%3ACI)
[![Release](https://github.com/rdark/za/workflows/Release/badge.svg)](https://github.com/rdark/za/actions?query=workflow%3ARelease)

A CLI tool for managing daily journal entries and standup notes in a zettelkasten-style knowledge base.

## Features

- **Automatic work extraction** - Populate standups with yesterday's completed work and today's goals
- **GitHub integration** - Automatically include PRs created yesterday and open/unreviewed PRs in standups
- **Goals management** - Copy unfinished goals between journal entries
- **Smart link fixing** - Resolve relative date references (Yesterday/Tomorrow) and cross-references
- **Slack-ready updates** - Generate concise daily updates in Slack-compatible format
- **Gap handling** - Handles weekends and holidays automatically
- **Template-based generation** - Integrates with external tools like [zk](https://github.com/zk-org/zk)

## Installation

### Download Binary

Download the appropriate binary for your platform from the [releases page](https://github.com/rdark/za/releases), make it executable, and move it to your PATH:

```bash
# Linux/macOS
chmod +x za-*
sudo mv za-* /usr/local/bin/za

# Or add to your PATH
mv za-* ~/bin/za  # Make sure ~/bin is in your PATH
```

### Build from Source

```bash
go install github.com/rdark/za@latest
```

## Quick Start

```bash
# Generate configuration file
za generate-config

# Edit .za.yaml with your paths and settings
# Then start using:

za generate-journal           # Create today's journal
za generate-standup          # Create standup with yesterday's work and today's goals
za standup-slack             # Generate Slack-ready daily update
za fix-links journal/2025-01-15.md  # Fix stale links
```

## Configuration

Create `.za.yaml` in your notes directory:

```yaml
journal:
  dir: ./journal
  work_done_sections: ["work completed", "worked on"]
  create:
    cmd: "zk new --title 'Daily Log {date}' --print-path journal/"

standup:
  dir: ./standup
  work_done_section: "Worked on yesterday"
  create:
    cmd: "zk new --title 'Standup {date}' --print-path standup/"

search_window_days: 30

# GitHub integration (optional)
# Requires GitHub CLI (gh) to be installed and authenticated
github:
  enabled: true
  org: "my-org"  # GitHub organization to search for PRs
```

### GitHub Integration

The GitHub integration is optional and requires:
1. [GitHub CLI (gh)](https://cli.github.com/) installed and authenticated
2. Configuration in `.za.yaml` with `github.enabled: true` and your organization name

When enabled, `generate-standup` will automatically:
- Add PRs created yesterday (in any state) to "Worked on yesterday"
- Add PRs opened in the last 7 days that are still open and unreviewed to "Working on today"

## Usage

### Generate Notes

```bash
za generate-journal              # Creates journal with fixed links
za generate-standup              # Creates standup with yesterday's work and today's goals
za generate-standup --no-work    # Skip work extraction
```

### Slack Updates

```bash
za standup-slack                 # Generate update for today
za standup-slack 2025-01-15      # Generate update for specific date
```

Outputs a concise summary of yesterday's completed work and today's planned goals in Slack-compatible format:

```
previous:
* Completed feature X
* Fixed bug Y
next:
* Review code changes
* Deploy to staging
```

### Fix Links

```bash
za fix-links journal/2025-01-15.md --dry-run  # Preview
za fix-links journal/2025-01-15.md            # Apply
```

Fixes temporal links (Yesterday/Tomorrow) and cross-references (Journal/Standup) to point to actual existing files.

## File Format

Notes use date-based filenames (`YYYY-MM-DD.md`) with markdown + YAML frontmatter:

```markdown
---
title: Daily Log 2025-01-15
---

# Work Completed

- Implemented feature X
- Fixed bug Y

[Yesterday](2025-01-14.md) | [Standup](../standup/2025-01-15.md)
```

## License

MIT License - see LICENSE file for details

## Built With

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration
- [Goldmark](https://github.com/yuin/goldmark) - Markdown parsing
