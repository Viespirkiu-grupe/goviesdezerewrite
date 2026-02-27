# Product Spec

## Purpose

`goviesdeze` is a lightweight HTTP file service for:

1. storing files by client-provided filename,
2. retrieving files with byte-range support,
3. deleting files,
4. importing files from remote URLs,
5. reporting total storage usage.

It is optimized for simple deployment, local filesystem storage, and low operational complexity.

## Target Users

- Internal services that need stable file URLs and simple upload/download APIs.
- Automation jobs that ingest remote content and store it by digest.
- Consumers that need partial reads via HTTP Range for previews/streaming.

## Scope

In scope:

- REST-like HTTP endpoints on top of Gin.
- Optional API-key auth for non-GET operations.
- Local filesystem sharded storage.
- Archive entry extraction on download (`?extract=...`).
- Usage accounting persisted in `usage.json`.

Out of scope (current implementation):

- S3 or object storage backend (README mentions it, code currently does not implement it).
- Full IAM/authn/authz model beyond shared API key.
- Multi-tenant quotas and hard limits.
- Transactional metadata database.

## Functional Requirements

- `PUT /file/:filename` writes raw request body to storage path derived from `:filename`.
- `GET /file/:filename` returns full file by default.
- `GET /file/:filename` with `Range` must return `206` and valid `Content-Range` for valid single range.
- `GET /file/:filename?extract=<entry>` returns extracted archive entry bytes.
- `DELETE /file/:filename` removes file and updates usage accounting.
- `POST /download-url` downloads a URL, stores content by md5 filename, and returns digest/size.
- `GET /storage-usage` returns total size in bytes.

## Non-Functional Requirements

- Keep dependencies minimal and startup fast.
- Avoid memory leaks by closing all `io.ReadCloser` values in call sites.
- Support frequent reads and moderate write concurrency.
- Return deterministic JSON error payloads for client handling.

## Success Criteria

- Upload/download/delete cycle works for binary files.
- Range requests behave correctly for start-end, start-open, and suffix ranges.
- Archive extraction selects the closest matching entry and returns content.
- Usage endpoint reflects updates after writes/deletes/url-imports.
