# Slice: Project Setup & Create Timesheet

## User Story
As a developer, I want to create a timesheet entry in bexio via an MCP tool so that I can log my time through natural language.

## Outer Boundary
- Entry: MCP tool `create_timesheet` over stdio transport
- Test file: `mcp_test.go` (or `server_test.go`)
- Framework: `go test`

## Shell Boundaries
- CLI shell: `main.go` — JSON-RPC over stdio, request dispatch, env wiring
- HTTP shell: `bexio_client.go` — authenticated `POST /2.0/timesheet` to bexio API

## Functional Core
- No standalone pure core class yet in this slice.
- Shared domain data structures live in `timesheet.go` and are used by both shells.

## Acceptance Criterion
Given a running MCP server with a valid bexio API token  
When the client calls `create_timesheet` with text, contact_id, pr_project_id, client_service_id, tracking (range with date, start, end), and other required fields  
Then a timesheet entry is created in bexio and the tool returns the created entry

## Includes
- Go module init (`go.mod`)
- MCP server with stdio transport
- `create_timesheet` tool with struct-based input schema
- BexioClient with Bearer token auth (from env var `BEXIO_API_TOKEN`)
- HTTP call to `POST /2.0/timesheet`

## API Reference
- `docs/bexio-api-timesheet.json` — POST /2.0/timesheet schema

## Uses
None (greenfield)

## Implemented
- `main.go`
- `bexio_client.go`
- `timesheet.go`
- `server_test.go`
- `bexio_client_test.go`
