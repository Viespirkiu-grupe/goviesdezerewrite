# HTTP API Spec

Base URL: `http://<host>:<port>`

Auth header for mutating endpoints when enabled:

- `X-API-Key: <key>`

## `PUT /file/:filename`

Upload raw body to storage.

Request:

- Path param: `filename`.
- Body: binary stream.

Response `200`:

```json
{
  "uploaded": "example.bin",
  "replaced": true,
  "oldSize": 120,
  "newSize": 240,
  "totalSize": 10240
}
```

Errors:

- `500` when directory/file write fails.

## `GET /file/:filename`

Download stored file, optional extraction and range.

Query:

- `extract=<entry-path>` optional archive entry path hint.

Headers:

- `Range` supports single ranges:
  - `bytes=0-99`
  - `bytes=500-`
  - `bytes=-200`

Response:

- `200` full content.
- `206` partial content with headers:
  - `Content-Range`
  - `Content-Length`
  - `Accept-Ranges: bytes`
- `404` if file or archive entry not found.
- `400` if archive is invalid for extraction mode.
- `416` for invalid range.

## `DELETE /file/:filename`

Delete stored file.

Response `200`:

```json
{
  "deleted": "example.bin",
  "sizeFreed": 240
}
```

Errors:

- `404` if file not found.
- `500` if remove fails.

## `POST /download-url`

Fetch remote URL and store content by md5 digest filename.

Request JSON:

```json
{ "url": "https://example.com/file.pdf" }
```

Response `200`:

```json
{
  "md5": "<digest>",
  "size": 12345
}
```

Errors:

- `400` malformed JSON/missing url.
- `500` fetch/write/move/stat errors.

## `GET /storage-usage`

Return total usage snapshot.

Response `200`:

```json
{
  "totalSizeBytes": 123456
}
```

## Error Format

Current handlers return simple JSON object:

```json
{ "error": "message" }
```

No formal machine-readable error code exists yet; this is a candidate improvement.
