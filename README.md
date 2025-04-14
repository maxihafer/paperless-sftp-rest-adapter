# paperless-sftp-rest-adapter

A lightweight adapter that monitors a directory for new files and automatically uploads them to your Paperless instance via the REST API. 
This allows uploading documents from your local network to paperless instances running on cloud providers or other services, where SFTP access is not available.

## Features

- Uses inotify to watch for new files
- Automatically uploads discovered files to Paperless
- Cleans up files after successful upload
- Runs as a Docker container with minimal footprint
- Cross-platform support

## Installation

### Docker (Recommended)

```bash
docker run -d \
  -v /path/to/documents:/consume \
  -e PAPERLESS_HOST=http://paperless:8000 \
  -e PAPERLESS_API_KEY=your_api_key \
  ghcr.io/maxihafer/paperless-sftp-rest-adapter:latest
```

### Binary Releases

Download the latest binary for your platform from the [Releases page](https://github.com/maxihafer/paperless-sftp-rest-adapter/releases).

## Configuration

Configure the application using the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `WATCH_DIR` | Directory to watch for new files | `/consume` |
| `PAPERLESS_HOST` | Paperless instance URL | `localhost:8000` |
| `PAPERLESS_API_KEY` | Paperless API token (required) | - |

## Usage

1. Configure the application with your Paperless instance details
2. Place files in the watched directory (mount the same path as a share in [atmoz/sftp](https://github.com/atmoz/sftp) or use a local directory)
3. Files will be automatically uploaded to Paperless and removed upon successful upload

## Development

### Prerequisites

- Go 1.24 or later
- Docker (for container builds)

### Building

```bash
# Build binary
go build -o paperless-sftp-rest-adapter

# Build Docker image
docker build -t paperless-sftp-rest-adapter .
```

## License

MIT License