# bexio-mcp

`bexio-mcp` is a Go MCP server that exposes bexio timesheet operations as MCP tools so you can log time from natural language clients (for example Claude Desktop or opencode) while keeping bexio as the source of truth.

## Prerequisites

- Go `1.25.5` (see `go.mod`)
- A bexio API token with permissions to create timesheets

## Build and Run

Build binary:

```bash
go build -o bexio-mcp .
```

Run server over stdio transport:

```bash
BEXIO_API_TOKEN="your-token" \
BEXIO_API_BASE_URL="https://api.bexio.com" \
./bexio-mcp
```

You can also run without building first:

```bash
BEXIO_API_TOKEN="your-token" \
BEXIO_API_BASE_URL="https://api.bexio.com" \
go run .
```

## Configuration

The server reads configuration from environment variables:

- `BEXIO_API_TOKEN`: bexio bearer token (required)
- `BEXIO_API_BASE_URL`: bexio API base URL, usually `https://api.bexio.com` (required)

## MCP Client Configuration Examples

opencode (`opencode.json`):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "bexio": {
      "type": "local",
      "command": ["/absolute/path/to/bexio-mcp"],
      "environment": {
        "BEXIO_API_TOKEN": "{env:BEXIO_API_TOKEN}",
        "BEXIO_API_BASE_URL": "https://api.bexio.com"
      }
    }
  }
}
```

Claude Desktop (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "bexio": {
      "command": "/absolute/path/to/bexio-mcp",
      "env": {
        "BEXIO_API_TOKEN": "your-token",
        "BEXIO_API_BASE_URL": "https://api.bexio.com"
      }
    }
  }
}
```

## Available Tools

- `create_timesheet`: create one timesheet entry in bexio

## Development

Run tests:

```bash
go test ./...
```

Run lint (must pass before committing):

```bash
golangci-lint run
```
