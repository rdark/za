# Za - Zettelkasten Augmentation

[![CI](https://github.com/rdark/za/workflows/CI/badge.svg)](https://github.com/rdark/za/actions?query=workflow%3ACI)
[![Release](https://github.com/rdark/za/workflows/Release/badge.svg)](https://github.com/rdark/za/actions?query=workflow%3ARelease)

A CLI tool for managing daily journal entries and standup notes in a zettelkasten-style knowledge base.

## Features

- **Automatic work extraction** from journal to standup notes
- **Smart link fixing** for relative date references (Yesterday/Tomorrow)
- **Gap handling** for weekends and holidays
- **Template-based note generation** via external tools (e.g., [zk](https://github.com/zk-org/zk))

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
za generate-standup          # Create standup with yesterday's work
za journal-work-done         # Extract work for Slack
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
```

## Usage

### Generate Notes

```bash
za generate-journal              # Creates journal with fixed links
za generate-standup              # Creates standup with work from yesterday
za generate-standup --no-work    # Skip work extraction
```

### Extract Work

```bash
za journal-work-done             # Today's work
za journal-work-done 2025-01-15  # Specific date (with fallback)
za standup-work-done             # From standup
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
