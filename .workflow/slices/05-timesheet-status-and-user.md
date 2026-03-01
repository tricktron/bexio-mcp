# Slice: Timesheet Status & Current User

## User Story
As a developer, I want to look up timesheet statuses and my own user ID so that defaults (status: Erledigt, user: me) can be resolved automatically.

## Outer Boundary
- Entry: MCP tools `list_timesheet_statuses`, `get_current_user`
- Test file: `server_test.go`
- Framework: `go test`

## Shell Boundaries
- HTTP: BexioClient.ListTimesheetStatuses - GET /2.0/timesheet_status
- HTTP: BexioClient.GetCurrentUser - GET /3.0/users/me
- MCP: list_timesheet_statuses tool registration
- MCP: get_current_user tool registration

## Functional Core
None - both tools are pure HTTP passthrough (getRaw pattern).

## Acceptance Criterion
Given a running MCP server  
When the client calls `list_timesheet_statuses`  
Then the tool returns available statuses with id and name

Given a running MCP server  
When the client calls `get_current_user`  
Then the tool returns the authenticated user's id and name

## API Reference
- `docs/bexio-api-timesheet.json` — GET /2.0/timesheet_status
- `docs/bexio-api-user.json` — GET /3.0/users/me

## Uses
- BexioClient from slice 01

## Implemented
- BexioClient.ListTimesheetStatuses (bexio_client.go)
- BexioClient.GetCurrentUser (bexio_client.go)
- MCP tool: list_timesheet_statuses (main.go)
- MCP tool: get_current_user (main.go)
- TestListTimesheetStatusesAcceptance (server_test.go)
- TestGetCurrentUserAcceptance (server_test.go)
