# Slice: Lint Fix & Smoke Test

## User Story
As a developer, I want to validate that the MCP server works with a real client and passes linting so that I have confidence before building more slices.

## Outer Boundary
- Entry: `golangci-lint run` + manual opencode smoke test
- Test file: N/A (manual validation)
- Framework: `golangci-lint`, `go build`

## Shell Boundaries
- FS: `.golangci.yml` — fix local-prefixes placeholder
- CLI: `go build .` — produce binary for smoke test

## Functional Core
None — this is a validation/config slice.

## Tasks
1. Fix `.golangci.yml`: replace `github.com/my/project` with `github.com/tricktron/bexio-mcp`
2. Run `golangci-lint run` — fix any issues it surfaces in existing code
3. Build binary: `go build -o bexio-mcp .`
4. Configure opencode to use the built binary as an MCP server
5. Smoke test: try `create_timesheet` through opencode against real Bexio
6. Fix any protocol issues discovered

## Acceptance Criterion
Given the golangci-lint config with correct local-prefixes  
When `golangci-lint run` is executed  
Then no lint errors are reported (or all are intentionally suppressed)

Given a built bexio-mcp binary configured in opencode  
When a `create_timesheet` tool call is made via opencode  
Then the timesheet is created in Bexio (validates both MCP protocol and Bexio API integration)

## Uses
- All slice 01 code
