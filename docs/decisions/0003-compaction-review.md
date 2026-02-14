# 0003 — Compaction Review

Date: 2026-02-14

## Status

Proposed

## Context

Compaction review asking "does this concept pay rent?" for every concept in the codebase. CodeScene health scores and independent code review were run in parallel.

### CodeScene Scores

| File                       | Health | Notes                                           |
|----------------------------|--------|-------------------------------------------------|
| `main.go`                  | 9.47   | Green                                           |
| `bexio_client.go`          | 9.38   | Green                                           |
| `timesheet_defaults.go`    | 10.0   | Optimal                                         |
| `server_test.go`           | 8.4    | Yellow — duplication in tests + large fake setup |
| `bexio_client_test.go`     | 9.09   | Green                                           |
| `server_contract_test.go`  | 9.38   | Green                                           |
| `timesheet.go`             | N/A    | Too small to score                              |
| `lookup_types.go`          | N/A    | Too small to score                              |

## Issues

### 1. `lookup_types.go` doesn't pay rent

`bexioClientService` and `bexioContact` are never used in production. The lookup endpoints return `json.RawMessage` by design — the server doesn't inspect or transform these payloads. The only consumers are contract tests that decode into them for shape assertions.

**Verdict:** Delete. Inline anonymous structs in tests.

**Rationale for not promoting to production:** The `json.RawMessage` pass-through is intentional. This server ferries lookup data to the LLM without understanding it. Adding typed structs would mean maintaining types that track Bexio's API shape for no benefit. The exceptions (`timesheetStatus`, current user parsing in `timesheet_defaults.go`) are justified because production logic actually inspects those values.

### 2. Acceptance tests mix two concerns

The acceptance tests in `server_test.go` assert on both:

1. **MCP results** — "I called `create_timesheet` and got back a timesheet" (user journey)
2. **HTTP internals** — "The request to Bexio had method POST, path `/2.0/timesheet`, auth header Bearer test-token" (integration contract)

This makes ~10 of 12 `bexio_client_test.go` tests redundant — they assert the same HTTP method/path/auth/body that the acceptance tests already verify.

Two options:

- **Option A:** Acceptance tests assert only on MCP results. Client tests own HTTP contract assertions. Clear separation.
- **Option B (current):** Acceptance tests assert both. Client tests are redundant.

Two `bexio_client_test.go` tests are unique regardless: `TestBexioClientReturnsErrorOnNon2xxStatus` and `TestBexioClientSendsAcceptHeader`.

**Verdict:** Resolved — Option A implemented (slice 11, ADR 0005). Acceptance tests assert only on MCP results. Client tests own all HTTP contract assertions. 13 `Received()` assertion blocks removed from `server_test.go`.

### 3. `BexioClient.GetTimesheet` doesn't pay rent in production

Zero production callers. Only used in `safeDeleteTestTimesheet` (contract test cleanup).

**Verdict:** Move to test file or delete.

### 4. ~~Three capture methods are duplication~~ — Resolved

`capture`, `captureWithTimesheet`, `captureWithSearch` on `fakeBexioAPI` were near-identical. Root cause: `fakeBexioCapturedRequest` has two body fields (`Body` and `SearchBody`) instead of one.

CodeScene flagged this as duplication.

**Verdict:** Resolved — all three methods deleted along with the `captured` field and `mu` mutex. After slice 11 removed all `Received()` callers from acceptance tests, the entire capture infrastructure became dead code. `fakeBexioAPI` is now just `{URL, server}`. `fakeBexioCapturedRequest` struct remains, owned by `bexio_client_test.go` for HTTP contract assertions.

### 5. Lookup acceptance tests are duplicated

`TestListContactsAcceptance`, `TestListProjectsAcceptance`, `TestListProjectsWithoutContactIDAcceptance`, `TestListClientServicesAcceptance`, `TestListPackagesAcceptance`, `TestListTimesheetStatusesAcceptance`, `TestGetCurrentUserAcceptance` all follow the same pattern: call tool, assert captured request, assert content contains a string.

CodeScene flagged this as the primary duplication cluster (~150 lines → ~50 with table-driven).

**Verdict:** Still valid but lower urgency. After slice 11, these tests only assert `result.IsError` and `strings.Contains(contentJSON, ...)` — no more `Received()` blocks. Table-driven refactor would collapse ~100 lines → ~40.

### 6. `createTimesheetInput` vs `bexioCreateTimesheetRequest` — near-duplicate structs

8 shared fields, differing by `UserID` optionality and `StatusID` presence. The mapping in `resolveTimesheetDefaults` manually copies every field.

The split is intentional (optional MCP input → resolved API request) but any new field must be added to both structs and the mapping function.

**Verdict:** Pays rent. No action now, but be aware when adding fields.
