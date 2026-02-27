# Testing and Quality Spec

## Quality Goal

Protect behavior while refactoring toward hexagonal architecture.

## Current Test Coverage Focus

- `internal/core/filequery/filequery_test.go`
  - `ParseRange` behavior (valid/invalid/suffix/clamped).
  - `BestMatch` behavior and threshold handling.
- `internal/core/archivequery/archivequery_test.go`
  - maps archive errors,
  - verifies chosen filename,
  - verifies opened readers are closed.
- `internal/handlers/file/download_test.go`
  - verifies response helper closes reader,
  - verifies HTTP range behavior for suffix and clamped ranges.

## TDD Workflow

For each change:

1. Add/adjust failing test that captures target behavior.
2. Implement the minimal code to make test pass.
3. Run full package tests (`go test ./...`).
4. Refactor for readability while keeping tests green.

## Regression Checklist

- Range requests:
  - `bytes=start-end`
  - `bytes=start-`
  - `bytes=-suffix`
  - invalid/multi-range rejection.
- Archive extraction:
  - invalid archive -> `400`,
  - missing entry -> `404`,
  - successful extraction streams expected bytes.
- Reader lifecycle:
  - all `io.ReadCloser` call paths close resources.

## CI Quality Gates

Minimum recommended gates:

- `gofmt -w` clean,
- `go test ./...` pass,
- optional: `go vet ./...`,
- optional: staticcheck.

## Known Testing Gaps

- No tests for upload/delete/download-url handlers.
- No tests for middleware auth behavior.
- No end-to-end tests against real archive samples.
- No stress/concurrency tests for usage accounting updates.
