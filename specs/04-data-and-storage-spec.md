# Data and Storage Spec

## Filesystem Layout

Root directory is configured by `STORAGE_PATH`.

Sharding rule:

- if filename length >= 2, path is `<root>/<first2>/<filename>`
- otherwise `<root>/<filename>`

Examples:

- `abcdef.txt` -> `/storage/ab/abcdef.txt`
- `x` -> `/storage/x`

## Candidate Path Resolution

Reads/deletes use fallback candidates generated from base path:

1. original path,
2. jpeg/jpg extension swap,
3. lowercase extension variant,
4. extensionless variant,
5. extensionless + `.bin`, `.php`, `.null`.

This supports legacy/object naming inconsistencies.

## Usage Accounting

State:

- in-memory atomic int64 (`totalSize`),
- persisted mirror in `./usage.json`.

Lifecycle:

- startup: `LoadUsage()` initializes from file or zero,
- upload/delete/download-url mutate via `SetUsage`/`AddUsage`,
- every mutation writes full JSON snapshot.

Current consistency model:

- process-local atomic safety,
- no cross-process locking around `usage.json`.

## Archive Extraction Model

Archive read path uses:

- `archivequery.ReadBestMatch` for orchestration,
- `ziparchive.Service` adapter for archive operations,
- `ziputil` as implementation detail.

Entry selection:

- `filequery.BestMatch` uses normalized Levenshtein similarity,
- exact case-insensitive equality wins,
- minimum similarity threshold: `0.4`.

Returned content is currently fully buffered into memory before response.

## IO Ownership Rules

Contract:

- any function returning `io.ReadCloser` transfers close responsibility to caller.
- caller must `defer Close()` as soon as ownership is received.

Implemented safeguards:

- `archivequery.ReadBestMatch` closes reader after `io.ReadAll`.
- `writeResponse` closes provided reader.
- tests assert close behavior on critical paths.
