# TCRS CLI

A Go-based command-line interface for the TCRS (Timecard Recording System).

## Features

- Single binary deployment (zero dependencies)
- Session persistence with automatic cookie management
- JSON output support for scripting
- Claude Code skill integration

## Installation

```bash
# Build and install
make install

# Or just build
make build
```

## Usage

### Environment Setup

```bash
export TCRS_BASE_URL="http://example.com/TCRS"
export TCRS_USER="your_user_id"
export TCRS_PASSWORD="your_password"
```

### Authentication

```bash
# Login (uses TCRS_USER and TCRS_PASSWORD env vars)
tcrs login

# Check status
tcrs status

# Logout
tcrs logout
```

### Viewing Data

```bash
# List projects and activities for today
tcrs projects

# List projects for a specific date
tcrs projects --date 2025-01-20

# View current week timecard
tcrs week

# View specific week (use Monday date)
tcrs week --date 2025-01-13
```

### Saving Timecard

```bash
# Save from JSON file
tcrs save --date 2025-01-13 --file entries.json

# Save from stdin
cat entries.json | tcrs save --date 2025-01-13 -f -
```

### JSON Format for Save

```json
{
  "entries": [
    {
      "project_id": "12345",
      "activity_id": "5",
      "progress": 0,
      "days": [
        {"hours": 8, "note": "", "progress": 0},
        {"hours": 8, "note": "", "progress": 0},
        {"hours": 8, "note": "", "progress": 0},
        {"hours": 8, "note": "", "progress": 0},
        {"hours": 8, "note": "", "progress": 0},
        {"hours": 0, "note": "", "progress": 0},
        {"hours": 0, "note": "", "progress": 0}
      ]
    }
  ]
}
```

### Global Flags

- `--json` - Output in JSON format
- `--verbose` / `-v` - Enable verbose output

## Environment Variables

- `TCRS_BASE_URL` - **Required**. TCRS server URL (e.g., `http://example.com/TCRS`)
- `TCRS_USER` - User ID for login (optional, can use argument instead)
- `TCRS_PASSWORD` - Password for login (optional, can use argument instead)
- `TCRS_CACHE_DIR` - Session cache directory (default: `~/.tcrs`)

## Development

```bash
# Build
make build

# Run tests
make test

# Cross-compile for all platforms
make cross-compile

# Install skill only
make skill
```

## Claude Code Skill

Install the skill to your project:

```bash
cp -r skills/tcrs <your-project>/.claude/skills/tcrs
```

After installation, the TCRS skill is available in Claude Code:

- Trigger: `/tcrs` or keywords like "timecard", "工時卡"
- Commands are executed via the `tcrs` CLI

## License

MIT
