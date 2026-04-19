# Security Policy

## Reporting a Vulnerability

If you discover a security issue in `babeliocli`, please report it privately.

**Do not open a public GitHub issue.**

Instead, email **yoan.bernabeu@pm.me** with:

- A description of the issue
- Steps to reproduce
- Potential impact (e.g., session leakage, arbitrary command execution)

I will acknowledge your report within a reasonable timeframe and coordinate a fix before any public disclosure.

## Scope

`babeliocli` is a read-only client that stores a Babelio session in a local JSON file. Relevant concerns include:

- Leakage of session cookies through logs, error messages, or telemetry
- Insecure permissions on the session file (it should always be `0600`)
- Command injection via CLI flags, filenames, or HTML parsing
- Unvalidated URLs or paths passed to the underlying HTTP client

## Out of Scope

- Changes to Babelio's HTML that break parsers (please open a regular issue instead)
- Rate-limiting or detection by Babelio (please respect the service)

## Supported Versions

Only the latest release is supported with fixes.
