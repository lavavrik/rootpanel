# go-sm

A basic Go Web API for listing files in directories.

## Features

- Simple HTTP API for file system exploration
- JSON responses with file metadata
- Health check endpoint
- Built-in documentation page
- Safe path handling to prevent directory traversal

## Configuration

The application supports configuration via a JSON configuration file (`config.json`). If the file doesn't exist, default values are used.

### Configuration File Format

Create a `config.json` file in the same directory as the executable:

```json
{
  "server": {
    "port": "8080"
  }
}
```

### Configuration Options

- `server.port`: Server port (default: "8080")

### Environment Variable Override

The `PORT` environment variable takes precedence over the configuration file:

```bash
PORT=9090 ./go-sm  # Runs on port 9090 regardless of config file
```

## Quick Start

1. Build and run the application:
```bash
go run main.go
```

2. The server will start on port 8080 (or the port specified in the `PORT` environment variable)

3. Visit `http://localhost:8080` for API documentation

## API Endpoints

### `GET /`
Returns an HTML page with API documentation and usage examples.

### `GET /health`
Health check endpoint that returns service status.

**Response:**
```json
{
  "status": "healthy",
  "service": "go-sm file API"
}
```

### `GET /files`
Lists files in the current directory.

**Response:**
```json
{
  "path": ".",
  "files": [
    {
      "name": "main.go",
      "is_dir": false,
      "size": 1234,
      "mod_time": "2024-01-01 12:00:00"
    }
  ]
}
```

### `GET /files?path=<directory>`
Lists files in the specified directory.

**Parameters:**
- `path` (query parameter): Directory path to list files from

**Example:**
```bash
curl "http://localhost:8080/files?path=/tmp"
```

## Building

```bash
go build -o go-sm
./go-sm
```

## Environment Variables

- `PORT`: Server port (overrides configuration file, default: uses config file or "8080")

## Response Format

All file listing responses include:
- `name`: File or directory name
- `is_dir`: Boolean indicating if it's a directory
- `size`: File size in bytes
- `mod_time`: Last modification time in "YYYY-MM-DD HH:MM:SS" format