# babeliocli — AI Agent Skills

These skills help AI coding agents (Claude Code, Cursor, Copilot, …) drive `babeliocli` without having to re-discover the tool from scratch.

## Available Skills

| Skill | Description |
|-------|-------------|
| [babeliocli-setup](babeliocli-setup/SKILL.md) | Install, log in (password or SSO cookie import), manage the session |
| [babeliocli-library](babeliocli-library/SKILL.md) | Read shelves, list books, group by date, find titles |
| [babeliocli-discover](babeliocli-discover/SKILL.md) | Search the catalogue, inspect book details, read critiques |

## What Are Skills?

Skills are Markdown files that give AI coding agents specialized knowledge about a tool. Instead of the agent exploring the codebase to find answers, it gets curated, accurate instructions that describe the commands, the JSON schema, and common pitfalls.

Point your agent at the relevant `SKILL.md` (or install via [skills.sh](https://skills.sh/)) before asking it to run `babeliocli` on your behalf.
