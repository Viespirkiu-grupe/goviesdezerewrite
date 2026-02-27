# Goviesdeze

A Go implementation of the Viesdeze file storage service. This service provides HTTP endpoints for file upload, download, deletion, and URL-based file downloading with support for both local filesystem and S3 storage.

## Features

- **File Upload** (PUT /file/:filename) - Upload files to local storage or S3
- **File Download** (GET /file/:filename) - Download files with range request support
- **File Deletion** (DELETE /file/:filename) - Delete files from storage
- **URL Download** (POST /download-url) - Download files from URLs and store them
- **Storage Usage** (GET /storage-usage) - Get total storage usage statistics
- **API Key Authentication** (optional) - Secure endpoints with API key
- **Request Logging** - Log all requests with timing information
- **Sharded Storage** - Files stored in subdirectories based on filename prefix
- **Candidate Path Generation** - Smart file lookup with multiple path variations

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd goviesdeze
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o goviesdeze .
```

## Configuration

Copy `example.env` to `.env` and configure the following environment variables:

```bash
cp example.env .env
```

### Environment Variables

- `API_KEY` - API key for authentication (default: "super-secret-key")
- `REQUIRE_API_KEY` - Whether to require API key authentication (default: true)
- `PORT` - Server port (default: "3000")
- `STORAGE_PATH` - Local storage path (default: "./storage")
- `S3` - Enable S3 storage (default: false)
- `S3_ENDPOINT` - S3 endpoint URL
- `S3_ACCESS_KEY` - S3 access key
- `S3_SECRET_KEY` - S3 secret key
- `S3_REGION` - S3 region (default: "us-east-1")
- `S3_BUCKET` - S3 bucket name (default: "viespirkiai")

## Usage

### Start the server:
```bash
./goviesdeze
```

### API Endpoints

#### Upload File
```bash
curl -X PUT -H "X-API-Key: your-api-key" \
  --data-binary @file.txt \
  http://localhost:3000/file/example.txt
```

#### Download File
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:3000/file/example.txt
```

#### Delete File
```bash
curl -X DELETE -H "X-API-Key: your-api-key" \
  http://localhost:3000/file/example.txt
```

#### Download from URL
```bash
curl -X POST -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/file.jpg"}' \
  http://localhost:3000/download-url
```

#### Get Storage Usage
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:3000/storage-usage
```

## Storage Structure

Files are stored using a sharded directory structure where the first two characters of the filename become a subdirectory. For example:

- `abc123.txt` → `./storage/ab/abc123.txt`
- `def456.jpg` → `./storage/de/def456.jpg`

## Range Requests

The service supports HTTP range requests for efficient file streaming:

```bash
curl -H "Range: bytes=0-1023" \
  http://localhost:3000/file/example.txt
```

## Dependencies

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [AWS SDK for Go](https://github.com/aws/aws-sdk-go) - S3 integration
- [Filetype](https://github.com/h2non/filetype) - File type detection
