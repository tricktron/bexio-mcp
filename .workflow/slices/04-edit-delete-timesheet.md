# Slice: Edit & Delete Timesheet

## User Story
As a developer, I want to edit or delete a timesheet entry so that I can fix mistakes in my time logs.

## Outer Boundary
- Entry: MCP tools `edit_timesheet`, `delete_timesheet`
- Test file: `mcp_test.go`
- Framework: `go test`

## Shell Boundaries
- HTTP: BexioClient — calls `POST /2.0/timesheet/{id}` (edit) and `DELETE /2.0/timesheet/{id}`

## Functional Core
Discovered via TDD.

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
