# AGENTS.md

This file describes how LLM agents should work in this repository.

## Architecture Summary

- MCP server implemented in Go and running over stdio
- Functional core / imperative shell split:
  - Shell: MCP transport + HTTP integration
  - Core: request/response domain structs for timesheets
- Current protocol implementation uses `github.com/modelcontextprotocol/go-sdk/mcp`
- Historical note: hand-rolled JSON-RPC was attempted and then superseded after real-client smoke test failure (see `docs/decisions/0002-hand-rolled-mcp-and-project-hygiene.md`)

## File Conventions

- `main.go`: CLI/MCP shell entrypoint, tool registration, stdio runtime
- `bexio_client.go`: HTTP shell for bexio REST API calls
- `timesheet.go`: domain/request/response types shared by MCP and HTTP layers
- `server_test.go`: acceptance-style MCP test using in-memory transport + fake bexio API
- `bexio_client_test.go`: unit tests for HTTP client behavior with `httptest`

## Testing Patterns

- Acceptance test pattern: exercise MCP tool calls end-to-end in-process
- HTTP boundary test pattern: `httptest.Server` captures request and returns realistic JSON
- Keep tests table-driven where helpful, and favor explicit request/response assertions

## Key Decisions and Constraints

- Use Go MCP SDK for protocol correctness (`docs/decisions/0001-go-mcp-server-for-bexio-timesheets.md`)
- Hand-rolled MCP remains documented as a rejected/superseded path (`docs/decisions/0002-hand-rolled-mcp-and-project-hygiene.md`)
- Grow interfaces per slice (do not preemptively design large interfaces)
- `golangci-lint run` must pass before committing

## Pointers

- ADRs: `docs/decisions/0001-go-mcp-server-for-bexio-timesheets.md`, `docs/decisions/0002-hand-rolled-mcp-and-project-hygiene.md`
- C4 model: `docs/architecture.dsl`
- Slice plan: `.workflow/slices/`
