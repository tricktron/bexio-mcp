# Slice: Adopt MCP SDK

## User Story
As a developer, I want the MCP server to use the official Go SDK so that it works correctly with real MCP clients.

## Outer Boundary
- Entry: MCP tool `create_timesheet` over stdio transport (same as slice 01, different implementation)
- Test file: `server_test.go`
- Framework: `go test`

## Shell Boundaries
- CLI: `main.go` — SDK-based MCP server with stdio transport, tool registration, env wiring
- HTTP: `bexio_client.go` — unchanged (keep as-is)

## Functional Core
- `timesheet.go` — unchanged (keep as-is)

## What Changes
- Replace hand-rolled JSON-RPC in `main.go` with `github.com/modelcontextprotocol/go-sdk/mcp` server + stdio transport
- Remove all hand-rolled protocol types (`rpcRequest`, `rpcResponse`, `rpcError`, `readPayload`, `writeResponse`)
- Register `create_timesheet` as an SDK tool with proper schema
- Update `server_test.go` to test through the SDK's server API instead of raw JSON-RPC

## What Stays
- `bexio_client.go` — no changes
- `bexio_client_test.go` — no changes
- `timesheet.go` — no changes
- `.golangci.yml` — no changes

## Acceptance Criterion
Given a built bexio-mcp binary configured in opencode
When a `create_timesheet` tool call is made via opencode
Then the server connects successfully (no timeout) and the timesheet is created in Bexio

Given the existing `server_test.go` acceptance test
When `go test ./...` is run
Then all tests pass with the SDK-based implementation

## Uses
- `bexio_client.go` from slice 01
- `timesheet.go` from slice 01

## Implemented
- Replaced hand-rolled MCP JSON-RPC handling in `main.go` with `github.com/modelcontextprotocol/go-sdk/mcp`
- Registered `create_timesheet` via SDK tool wiring and stdio transport
- Updated acceptance coverage to use SDK in-memory transport client/server flow
- Consolidated shared test helpers into `test_helpers_test.go`

## Status
Complete
