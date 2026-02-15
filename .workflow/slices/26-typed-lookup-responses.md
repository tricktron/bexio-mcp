# Slice: Typed output schemas for lookup tools

## User Story
As an LLM consuming bexio-mcp tools, I want lookup tool responses (contacts, projects, services, etc.) to have a documented schema so I know the structure of what each tool returns.

## Prerequisites
- Slice 23 (typed output schemas for timesheet tools) must be completed first — it introduces the `registerTool[TInput, TOutput]` signature

## Outer Boundary
- Entry: Lookup tool registrations in `main.go`
- Test file: `server_test.go` (acceptance), `server_contract_test.go` (contract), `schema_test.go`
- Framework: `go test`

## Shell Boundaries
- **HTTP**: `BexioClient` methods change from returning `json.RawMessage` to typed structs

## Functional Core
- Define typed response structs for each lookup endpoint
- Change `BexioClient` methods to deserialize into typed structs
- Wire `registerTool[TInput, TOutput]` with concrete output types for all 6 lookup tools
- Wrapper structs for list endpoints (SDK requires `outputSchema.type == "object"`)

## Acceptance Criteria

### AC1: All lookup tools have output schemas in tools/list
Given the MCP server is running
When a client calls `tools/list`
Then every lookup tool has a non-nil `outputSchema` with `type: "object"` and populated `properties`

### AC2: Contract tests pass with typed responses
Given lookup tools run against fake and real environments
When the tools return successfully
Then responses deserialize into the typed structs without error

## Uses
- `registerTool[TInput, TOutput]` from slice 23
- Contract test pattern from `server_contract_test.go`
- Bexio API docs: `docs/bexio-api-contact.json`, `docs/bexio-api-project.json`, `docs/bexio-api-client_service.json`, `docs/bexio-api-package.json`, `docs/bexio-api-user.json`

## Scope: Tools affected

| Tool                      | Current return    | Target TOutput                     | API spec source                    |
| ------------------------- | ----------------- | ---------------------------------- | ---------------------------------- |
| `list_contacts`           | `json.RawMessage` | `listContactsResult` (new wrapper) | `bexio-api-contact.json`           |
| `list_projects`           | `json.RawMessage` | `listProjectsResult` (new wrapper) | `bexio-api-project.json`           |
| `list_client_services`    | `json.RawMessage` | `listServicesResult` (new wrapper) | `bexio-api-client_service.json`    |
| `list_packages`           | `json.RawMessage` | `listPackagesResult` (new wrapper) | `bexio-api-package.json`           |
| `list_timesheet_statuses` | `json.RawMessage` | `listStatusesResult` (new wrapper) | `bexio-api-timesheet.json` (reuse) |
| `get_current_user`        | `json.RawMessage` | `bexioUser` (new, single object)   | `bexio-api-user.json`              |

## Design note
Define minimal structs matching the fields the MCP server actually exposes to LLMs (id, name, and key fields), not full API response structs. The contract tests already assert on `id` + `name`/`name_1` — match that granularity. Fields can be added later if LLMs need them.

## Implemented
- Shell Boundaries: `BexioClient` methods now return typed structs
- Functional Core: `lookup.go` — `bexioContact`, `bexioProject`, `bexioService`, `bexioPackage`, `bexioUser` + wrapper types
- Files modified: `main.go`, `bexio_client.go`, `lookup.go`, `timesheet_defaults.go`
