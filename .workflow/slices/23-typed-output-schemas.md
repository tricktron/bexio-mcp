# Slice: Typed output schemas for timesheet tools

## User Story
As an LLM consuming bexio-mcp tools, I want timesheet tool responses to have a documented schema so I know the structure of what each tool returns.

## Outer Boundary
- Entry: `registerTool[TInput, TOutput]` in `main.go`
- Test file: `server_test.go` (acceptance), `schema_test.go` (schema assertions)
- Framework: `go test`

## Shell Boundaries
- `registerTimesheetTools` remains the MCP shell boundary for timesheet tool wiring

## Functional Core
- `registerTool` becomes `registerTool[TInput, TOutput any]`
- Handler signature changes from `func(ctx, input) (any, error)` to `func(ctx, input) (TOutput, error)`
- `mcp.AddTool[TInput, TOutput]` replaces `mcp.AddTool[TInput, any]`
- `marshalToolResult` is removed — SDK handles marshaling via `StructuredContent`
- New `deleteTimesheetResult` struct: `{Success bool}`
- New `searchTimesheetsResult` wrapper struct: `{Results []bexioTimesheet}` (SDK requires `outputSchema.type == "object"`)
- New `deleteTimesheet(ctx, bexio, id)` helper centralizes delete response decoding outside registration closures
- Lookup tools stay `TOutput=any` for now (see slice 26)

## Acceptance Criteria

### AC1: Timesheet tool output schemas appear in tools/list
Given the MCP server is running
When a client calls `tools/list`
Then `create_timesheet`, `edit_timesheet`, `delete_timesheet`, and `search_timesheets` each have a non-nil `outputSchema` with `type: "object"` and populated `properties`

### AC2: create/edit use bexioTimesheet as TOutput
Given `create_timesheet` or `edit_timesheet` is called
When the tool returns successfully
Then the response validates against the `bexioTimesheet` output schema

### AC3: delete uses deleteTimesheetResult as TOutput
Given `delete_timesheet` is called
When the tool returns successfully
Then the response is `{"success": true}` matching the `deleteTimesheetResult` output schema

### AC4: search wraps slice in object
Given `search_timesheets` is called
When the tool returns a list of timesheets
Then the response is wrapped in `{"results": [...]}` matching the `searchTimesheetsResult` output schema

### AC5: Lookup tools still work with TOutput=any
Given any lookup tool (`list_contacts`, `list_projects`, etc.) is called
When the tool returns successfully
Then the tool works as before (no `outputSchema`, no regression)

## Decision
See ADR `docs/decisions/0008-typed-output-schemas.md`

## Uses
- `bexioTimesheet` struct from `timesheet.go`
- Go MCP SDK `mcp.AddTool[In, Out]` with typed `Out`
- Test fixtures in `server_test.go` for expected response shapes

## Scope: Tools affected

| Tool                | Current return     | Target TOutput                         |
| ------------------- | ------------------ | -------------------------------------- |
| `create_timesheet`  | `bexioTimesheet`   | `bexioTimesheet`                       |
| `edit_timesheet`    | `bexioTimesheet`   | `bexioTimesheet`                       |
| `delete_timesheet`  | `json.RawMessage`  | `deleteTimesheetResult` (new)          |
| `search_timesheets` | `[]bexioTimesheet` | `searchTimesheetsResult` (new wrapper) |
| Lookup tools (6)    | `json.RawMessage`  | `any` (unchanged, deferred to slice 26)|

## Implemented
- `registerTool[TInput, TOutput]` now drives typed output wiring through `mcp.AddTool[TInput, TOutput]`
- Timesheet tool outputs use object-shaped typed results (`bexioTimesheet`, `deleteTimesheetResult`, `searchTimesheetsResult`)
- `delete_timesheet` registration now delegates decoding to `deleteTimesheet(ctx, bexio, id)` for consistent handler structure
