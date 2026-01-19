# TCRS CLI

A Go-based command-line interface for the TCRS (Timecard Recording System).

## Project Structure

```
tcrs_cli/
├── main.go                  # Entry point
├── go.mod                   # Go module definition
├── Makefile                 # Build automation
│
├── cmd/                     # CLI commands (cobra)
│   ├── root.go              # Root command and global flags
│   ├── login.go             # tcrs login <user> <pass>
│   ├── logout.go            # tcrs logout
│   ├── status.go            # tcrs status
│   ├── projects.go          # tcrs projects [--date]
│   ├── week.go              # tcrs week [--date]
│   └── save.go              # tcrs save --date --file
│
├── internal/
│   ├── client/              # TCRS HTTP client
│   │   ├── client.go        # Main HTTP client
│   │   ├── session.go       # Cookie/session management
│   │   ├── parser.go        # HTML/JS parsing
│   │   └── errors.go        # Custom errors
│   │
│   └── config/              # Configuration
│       └── config.go        # Environment and config management
│
└── skills/tcrs/             # Claude Code skill
    └── SKILL.md
```

## Building

```bash
make build          # Build binary
make install        # Install to /usr/local/bin + skill
make cross-compile  # Build for all platforms
```

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/PuerkitoBio/goquery` - HTML parsing

## API Endpoints

- Login: `POST /servlet/VerifController` with `method=login`, `name=<user>`, `pw=<pass>`
- Logout: `GET /servlet/VerifController?method=logout`
- Projects: `GET /Timecard/timecard_week/daychoose.jsp?cho_date={date}`
- Save: `POST /Timecard/timecard_week/weekinfo_deal.jsp`

## Session Management

- Cookies stored in `~/.tcrs/<user_id>.cookies`
- Session info in `~/.tcrs/<user_id>.session`
- 12-hour session timeout
- Auto-loads saved session on client creation

## Development Notes

- Use `--json` flag for machine-readable output
- Week dates are expected to be Monday (week start)
- Activity IDs from the `act.append()` JavaScript calls
- Project IDs from `<option value="...">` in dropdowns
