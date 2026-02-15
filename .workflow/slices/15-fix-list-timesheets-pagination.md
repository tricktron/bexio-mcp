# Slice: Fix List Timesheets Missing Recent Entries

## Bug
`list_timesheets` returns empty results for recent dates when the user has
>500 total timesheets, because `ListTimesheets` fetches the first page
(oldest-first, limit 500).

## Root Cause
`BexioClient.ListTimesheets` calls `GET /2.0/timesheet` with no query params.
Bexio defaults to `order_by=id` (oldest first), `limit=500`, `offset=0`.
The existing HTTP boundary test (`TestBexioClientListTimesheets`) didn't
assert query params — `fakeBexioCapturedRequest` had no query field and
`newRawResponseServer` only captured `r.URL.Path`, not `r.URL.RawQuery`.

## Fix

### 1. Strengthen HTTP boundary test (red)
- Add `Query string` field to `fakeBexioCapturedRequest` (`server_test.go`)
- Capture `r.URL.RawQuery` in `newRawResponseServer` (`bexio_client_test.go`)
- Update `TestBexioClientListTimesheets` to assert
  `Query: "limit=2000&order_by=date_desc"` — test fails (red)
- Update `TestBexioClientSearchTimesheets` to assert same query params — fails

### 2. Add query params to HTTP client (green)
- `BexioClient.ListTimesheets`: set `order_by=date_desc&limit=2000` query params
- `BexioClient.SearchTimesheets`: same
- Both tests pass

## Shell Boundaries
- HTTP: `BexioClient.ListTimesheets` — sends `order_by=date_desc&limit=2000`
- HTTP: `BexioClient.SearchTimesheets` — sends `order_by=date_desc&limit=2000`

## Functional Core
- No changes — existing `filterTimesheetsByDateRange` unchanged

## Implemented
- `bexio_client.go`: `ListTimesheets` and `SearchTimesheets` — added query params
- `server_test.go`: `fakeBexioCapturedRequest` — added `Query` field
- `bexio_client_test.go`: `newRawResponseServer` — captures `r.URL.RawQuery`;
  `TestBexioClientListTimesheets` and `TestBexioClientSearchTimesheets` — assert query params
