# Contributing to babeliocli

Thanks for your interest in contributing! Here is everything you need to get started.

## Prerequisites

- Go 1.25+
- [golangci-lint](https://golangci-lint.run/) (optional, for linting)
- A Babelio account (for manual testing; mocks cover unit tests)

## Getting Started

```bash
git clone https://github.com/yoanbernabeu/babeliocli.git
cd babeliocli
make build
```

Run `make help` to see all available targets:

```
make build       Build the binary
make test        Run tests with the race detector
make lint        Run golangci-lint
make check       Run lint + test (same as CI)
make clean       Remove build artifacts
```

## Project Architecture

```
cmd/              Cobra commands (login, books, search, reviews, …)
internal/
  client/         HTTP client, cookie jar, session persistence
  babelio/        HTML parsers (shelves, books, search, reviews, book detail)
  output/         Format helpers (JSON / text)
main.go           Entrypoint
```

Because Babelio has no public API, every parser lives in `internal/babelio/` and is kept deliberately narrow: one file per concern, unit tests rely on saved HTML fixtures.

## Running Tests

```bash
make test
```

Tests run against HTML fixtures and do not require a live Babelio session.

## Linting

```bash
make lint
```

CI runs the linter on every push and PR. Active rules live in `.golangci.yml`.

## Before You Push

```bash
make check
```

This runs the linter and the test suite with the race detector — the same thing CI does.

## Code Style

- No obvious comments: explain the *why*, never the *what*
- Keep HTML parsers tolerant: Babelio's markup is quirky and evolves. Prefer structural selectors (class names, `href*=` matches) over regex on raw text
- Respect the JSON schema: don't rename a field in a non-breaking release — add a new one and deprecate instead
- Keep changes focused: one PR, one topic

## Updating HTML Parsers

When Babelio changes its markup, expect failures in the `internal/babelio` parsers. The repair loop is usually:

1. Save a fresh fixture from the broken page
2. Add or adjust a selector in the relevant parser
3. Add a regression test against the saved fixture

Always keep the parser defensive: return partial data rather than panicking if a field is missing.

## Submitting a Pull Request

1. Fork the repository and create a branch from `main`
2. Make your changes
3. Run `make check` to ensure tests pass and the linter is clean
4. Open a PR with a clear description

## Reporting Bugs

Please open an issue with:

- The exact `babeliocli` command you ran
- The JSON output (or the error on stderr)
- `babeliocli --version`
- Your OS and Go version

Do **not** paste your session file or cookies in a public issue.

## Requesting Features

Open an issue describing the use case. Read-only, agent-friendly workflows are preferred; the project deliberately avoids write operations on Babelio.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
