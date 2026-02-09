# Slice: List & Search Timesheets

## User Story
As a developer, I want to list and search my timesheet entries so that I can review what I've already logged.

## Outer Boundary
- Entry: MCP tools `list_timesheets`, `search_timesheets`
- Test file: `server_test.go`
- Framework: `go test`

## Shell Boundaries
- MCP shell: `newMCPServer` in `main.go` — tool registration and transport wiring
- HTTP shell: `BexioClient` in `bexio_client.go` — calls `GET /2.0/timesheet` and `POST /2.0/timesheet/search`

## Functional Core
- `bexioSearchField`: search filter contract shared by MCP and HTTP layers
- `bexioSearchTimesheetsRequest`: MCP request envelope for search filters
- `bexioTimesheet`: timesheet response model used across tool and HTTP boundaries

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

## Implemented
- Added MCP tool handlers for `list_timesheets` and `search_timesheets`
- Added `BexioClient.ListTimesheets` and `BexioClient.SearchTimesheets`
- Added acceptance coverage for list and search tool flows
- Added HTTP client unit coverage for list and search API calls
