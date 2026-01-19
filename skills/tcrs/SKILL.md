# TCRS Skill

Manage TCRS (Timecard Recording System) timecards via CLI.

## Triggers

- `/tcrs` - Direct skill invocation
- "tcrs", "timecard", "工時卡", "工時", "打卡"
- "查看工時", "填工時", "送工時"

## Instructions

When the user wants to interact with TCRS (Timecard Recording System), use the `tcrs` CLI tool.

### Available Commands

1. **Login** - Authenticate with TCRS
   ```bash
   # With arguments
   tcrs login <user_id> <password>

   # With environment variables (TCRS_USER, TCRS_PASSWORD)
   tcrs login
   ```

2. **Logout** - Clear session
   ```bash
   tcrs logout
   ```

3. **Status** - Check login status
   ```bash
   tcrs status
   ```

4. **Projects** - List available projects and activities
   ```bash
   tcrs projects [--date YYYY-MM-DD]
   ```

5. **Week** - View week timecard
   ```bash
   tcrs week [--date YYYY-MM-DD]
   ```

6. **Save** - Save timecard entries
   ```bash
   tcrs save --date YYYY-MM-DD --file entries.json
   tcrs save --date YYYY-MM-DD -f -  # Read from stdin
   ```

### Global Flags

- `--json` - Output in JSON format (useful for parsing)
- `--verbose` or `-v` - Enable verbose output

### Workflow Examples

#### Check Current Status
```bash
tcrs status
```

#### View This Week's Timecard
```bash
tcrs week --json
```

#### List Projects for Today
```bash
tcrs projects --json
```

#### Save Timecard Entries
Create a JSON file with entries:
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

Then save:
```bash
tcrs save --date 2025-01-13 --file entries.json
```

### Notes

- **Required**: Set `TCRS_BASE_URL` environment variable before use
- Session cookies are stored in `~/.tcrs/`
- Sessions expire after 12 hours
- Week dates should be the Monday of the desired week
- Use `--json` flag when parsing output programmatically
