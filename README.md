# babeliocli — a read-only, agent-friendly CLI for Babelio

[![CI](https://github.com/yoanbernabeu/babeliocli/actions/workflows/ci.yml/badge.svg)](https://github.com/yoanbernabeu/babeliocli/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/yoanbernabeu/babeliocli)](https://goreportcard.com/report/github.com/yoanbernabeu/babeliocli)

`babeliocli` is an unofficial, read-only command-line client for [babelio.com](https://www.babelio.com) — the French book community. It is built for shell scripting and AI coding agents: every command emits JSON by default (`-f text` is available for humans), errors go to stderr, and exit codes are meaningful.

Babelio does not expose a public API, so `babeliocli` authenticates through the regular web form (or imports cookies from your browser for Google/Facebook SSO) and parses the HTML of the logged-in pages.

## Features

- **Read your shelves** (lus, à lire, en cours, pense-bête, abandonnés, critiqués, non critiqués) with exact counts
- **List books** per shelf with title, author IDs, rating, reading status, start/end dates, number of readers
- **Search** the catalogue by title, author, or free text
- **Book details**: synopsis, publisher, number of pages, average rating, review count, genres
- **Reviews**: paginated list of critiques with author, date (raw + ISO), rating, and full body
- **JSON-first output**, `-f text` for a readable terminal view
- **Classic login** (username / password) or **cookie import** for SSO-linked accounts
- **Session file** stored under `$XDG_CONFIG_HOME/babeliocli/session.json` (mode 0600)
- **Expired-session detection**: commands return a clear error instead of silently returning empty data
- Proper **ISO-8859-1 → UTF-8** charset handling (Babelio still serves legacy Latin-1 HTML)

## Install

One-liner (Linux / macOS, amd64 / arm64):

```bash
curl -fsSL https://raw.githubusercontent.com/yoanbernabeu/babeliocli/main/install.sh | sh
```

The script downloads the matching binary from the latest GitHub release and installs it to `/usr/local/bin` (override with `BABELIOCLI_INSTALL_DIR=…`).

Or with Go:

```bash
go install github.com/yoanbernabeu/babeliocli@latest
```

Or download a prebuilt binary from the [releases page](https://github.com/yoanbernabeu/babeliocli/releases).

## Quick Start

```bash
# 1. Authenticate (username/password)
babeliocli login --username you@example.com

# ...or import cookies if you sign in with Google/Facebook:
babeliocli session import \
  --phpsessid 'XXX' \
  --bbac 'YYY' \
  --id-user 'NNN' \
  --username 'your_pseudo'

# 2. Verify
babeliocli whoami

# 3. List shelves and books
babeliocli shelves -f text
babeliocli books --shelf lus --limit 10 -f text

# 4. Search and explore
babeliocli search "becky chambers" -f text
babeliocli book /livres/Simmons-Les-Cantos-dHyperion-tome-1--Hyperion-1/5603 -f text
babeliocli reviews /livres/Simmons-Les-Cantos-dHyperion-tome-1--Hyperion-1/5603 --limit 5 -f text
```

## JSON output for scripting and agents

All commands emit JSON by default. Pipe through `jq` or consume directly from a coding agent:

```bash
# Count books read per month
babeliocli books --shelf lus | jq -r '.books[] | .read_end[0:7]' | sort | uniq -c

# Top 10 authors across your library
babeliocli books --shelf all | jq -r '.books[].author' | sort | uniq -c | sort -rn | head

# Get the average rating of every "Hyperion" search result
babeliocli search hyperion | jq '.results[] | {title, author, avg_rating}'
```

Errors go to stderr, so stdout stays valid JSON:

```bash
babeliocli shelves 2>/dev/null | jq '.shelves | length'
```

## Commands

| Command | Description |
|---|---|
| `babeliocli login` | Username/password auth; stores session locally |
| `babeliocli session import` | Import browser cookies (for Google/Facebook SSO) |
| `babeliocli whoami` | Show the current session's user |
| `babeliocli logout` | Delete the local session file |
| `babeliocli shelves` | List built-in shelves with counts |
| `babeliocli books --shelf <key>` | List books in a shelf (paginated automatically) |
| `babeliocli search <query>` | Search books by title/author |
| `babeliocli book <url\|path\|slug/id>` | Show a book's details |
| `babeliocli reviews <url\|path\|slug/id>` | List reader reviews |

Supported shelf keys: `all`, `lus`, `a-lire`, `en-cours`, `pense-bete`, `abandonnes`, `critiques`, `non-critiques`.

## How It Works

Babelio has no public API, so `babeliocli`:

1. Posts credentials to `/connection.php?r=1` (or reuses imported cookies).
2. Keeps the resulting `PHPSESSID`, `bbac`, `bbacml`, and `id_user` cookies in a jar, persisted to `~/.config/babeliocli/session.json`.
3. Fetches the relevant HTML pages (`/mabibliotheque.php`, `/livres/…`, `/recherche.php`) and parses them with [goquery](https://github.com/PuerkitoBio/goquery), transcoding ISO-8859-1 on the fly.

Because parsing depends on Babelio's HTML structure, the CLI can break if the site is redesigned. Open an issue if a command stops returning data.

## Responsible use

- `babeliocli` is **read-only**. It does not write, rate, review, or modify your Babelio account.
- Respect Babelio's terms of service and rate limits. Do not hammer the service: commands fetch pages sequentially at human speed.
- This project is not affiliated with or endorsed by Babelio.

## AI agent skills

`babeliocli` ships with [skills](.agents/skills/) for AI coding agents (Claude Code, Cursor, Copilot). They teach the agent how to authenticate, query shelves, and produce grouped reports directly from your editor.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Bug reports and PRs welcome.

## License

MIT © Yoan Bernabeu
