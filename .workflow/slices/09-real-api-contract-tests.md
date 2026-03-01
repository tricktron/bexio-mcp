# Slice: Real API Contract Tests (Environment-Polymorphic)

## User Story
As a developer, I want my acceptance tests to run against both the fake and the real Bexio API so that fake drift is caught automatically and I trust my test suite.

## Outer Boundary
- Entry: `go test ./... -run TestAcceptance`
- Test file: `server_test.go`, `server_contract_test.go`
- Framework: `go test`

## Shell Boundaries
- MCP in-memory transport (`mcp.NewInMemoryTransports`) between test client and server
- HTTP: Real Bexio API (`api.bexio.com`) when `BEXIO_API_TOKEN` is set
- HTTP: `httptest.Server` fake when token is absent

## Functional Core
- `contractTestEnv` factory: returns either fake-backed or real-backed test environment
- Assertion helpers that work in both modes:
  - Fake mode: assert exact IDs, captured HTTP requests
  - Real mode: assert response shapes, field presence, test prefix
- `safeDeleteTestTimesheet`: fetch-then-verify-then-delete guard

## Test Timesheet Safety

### Identification
- All test timesheets use text prefix `[MCP-TEST]`
- Example: `[MCP-TEST] contract test create`

### Safe Delete Guard
```
func safeDeleteTestTimesheet(t *testing.T, client BexioClient, id int):
    1. GET /2.0/timesheet/{id}
    2. Assert response.text starts with "[MCP-TEST]"
    3. Only then DELETE /2.0/timesheet/{id}
    4. If prefix missing → t.Fatalf("refusing to delete non-test timesheet %d", id)
```

### Manual Fallback
If test cleanup fails, search Bexio UI for `[MCP-TEST]` and delete manually.

## Acceptance Criterion
Given `BEXIO_API_TOKEN` is set in the environment
When I run `go test ./... -run TestAcceptance`
Then each acceptance test runs against both the fake and the real Bexio API
And all test timesheets are created with `[MCP-TEST]` prefix
And cleanup only deletes timesheets verified to carry the `[MCP-TEST]` prefix
And all assertions pass in both modes

Given `BEXIO_API_TOKEN` is NOT set
When I run `go test ./...`
Then only fake-mode tests run (existing behavior, no change)

## Depends On
- Slice 10 (fix smoke test findings) — fakes must be realistic before polymorphic assertions can share expectations

## Uses
- `newAcceptanceEnv` from `server_test.go` (refactored to support env selection)
- `BexioClient` from `bexio_client.go`
- `startFakeBexioAPI()` from `server_test.go`

## Notes
- Read-only tools (list_contacts, list_projects, etc.) need no cleanup — just assert response shape
- Mutating tools (create, edit, delete) follow: create `[MCP-TEST]` entry → assert → safe-delete
- Real-mode assertions are necessarily looser: can't predict exact IDs, timestamps, or list contents
- `client_service_id` and other required IDs for real-mode create must be discovered at test setup (e.g., list client_services, pick first)

## Implemented
- Added environment-polymorphic test setup (`contractTestEnvs`, `newAcceptanceEnvWith`) for fake and optional real API runs
- Added contract acceptance tests for `create_timesheet`, `list_timesheets`, `list_client_services`, and `list_contacts`
- Added `safeDeleteTestTimesheet` guard (`GET` verify `[MCP-TEST]` prefix before `DELETE`)
- Added `GetTimesheet` in `BexioClient` and fake handler support for `GET /2.0/timesheet/{id}`
