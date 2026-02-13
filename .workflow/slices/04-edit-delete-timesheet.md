# Slice: Edit & Delete Timesheet

## User Story
As a developer, I want to edit or delete a timesheet entry so that I can fix mistakes in my time logs.

## Outer Boundary
- Entry: MCP tools `edit_timesheet`, `delete_timesheet`
- Test file: `server_test.go`
- Framework: `go test`

## Shell Boundaries
- MCP transport shell: `main.go` wires `edit_timesheet` and `delete_timesheet` tool handlers to `BexioClient`
- HTTP shell: `BexioClient` calls `POST /2.0/timesheet/{id}` (edit) and `DELETE /2.0/timesheet/{id}`

## Functional Core
- `bexioCreateTimesheetRequest`: shared request model for create/edit payload fields
- `bexioEditTimesheetRequest`: MCP input envelope combining `id` and editable fields
- `bexioDeleteTimesheetRequest`: MCP input envelope for delete by `id`
- `bexioTimesheet`: shared response model returned by API and MCP tools

## Acceptance Criterion
Given an existing timesheet entry  
When the client calls `edit_timesheet` with an id and updated fields  
Then the entry is updated in bexio and the tool returns the updated entry

Given an existing timesheet entry  
When the client calls `delete_timesheet` with an id  
Then the entry is deleted from bexio and the tool confirms deletion

## API Reference
- `docs/bexio-api-timesheet.json` — POST /2.0/timesheet/{id}, DELETE /2.0/timesheet/{id}

## Uses
- BexioClient from slice 01

## Implemented
- Added MCP tools: `edit_timesheet`, `delete_timesheet`
- Added HTTP client methods: `EditTimesheet`, `DeleteTimesheet`
- Added acceptance coverage: `TestEditTimesheetAcceptance`, `TestDeleteTimesheetAcceptance`
- Added unit coverage: `TestBexioClientEditTimesheet`, `TestBexioClientDeleteTimesheet`
