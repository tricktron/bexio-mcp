# Slice: List & Search Timesheets

## User Story
As a developer, I want to list and search my timesheet entries so that I can review what I've already logged.

## Outer Boundary
- Entry: MCP tools `list_timesheets`, `search_timesheets`
- Test file: `mcp_test.go`
- Framework: `go test`

## Shell Boundaries
- HTTP: BexioClient — calls `GET /2.0/timesheet` and `POST /2.0/timesheet/search`

## Functional Core
Discovered via TDD.

## Acceptance Criterion
Given a running MCP server  
When the client calls `list_timesheets`  
Then the tool returns a list of recent timesheet entries

Given a running MCP server  
When the client calls `search_timesheets` with a date range filter  
Then the tool returns matching timesheet entries

## API Reference
- `docs/bexio-api-timesheet.json` — GET /2.0/timesheet, POST /2.0/timesheet/search

## Uses
- BexioClient from slice 01
