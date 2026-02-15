# 0007: Enrich MCP tool schemas with descriptions, enums, and patterns

## Status
Accepted

## Context
All MCP tool input schemas are auto-inferred from Go structs via `jsonschema.ForType()` but carry zero field descriptions, no enum constraints, and no format/pattern validation. LLM clients see field names and types but must guess what values like `tracking.type`, `criteria`, or date strings should contain. Invalid guesses pass schema validation, hit the Bexio API, and return opaque errors.

The go-sdk (v1.2.0) + jsonschema-go already support:
- `jsonschema:"..."` struct tags for per-field descriptions
- `TypeSchemas` map for custom type → schema overrides (enums, patterns)
- Automatic runtime validation before handler execution when using `mcp.AddTool[T]`

We are not using any of these capabilities.

## Decision
Enrich tool schemas in two layers:

1. **Descriptions**: Add `jsonschema:"..."` tags to all input struct fields and improve tool-level `Description` strings.
2. **Constraints**: Define typed enums (`TrackingType`, `SearchCriteria`) and date patterns using `TypeSchemas` + post-inference schema mutation. Extend `registerTool` to accept optional `jsonschema.ForOptions` so schemas carry enums and patterns that the SDK validates at runtime.

Do **not** split any existing tools. ADR-0006 already decided `search_timesheets` merges list+search; the other dual-mode behaviors (`list_projects` with optional `contact_id`, `create_timesheet` with auto-defaults) are natural optional-parameter patterns, not SRP violations.

## Alternatives Considered
**Descriptions only (no constraints)**: Lower effort but leaves the validation gap. LLMs know what fields mean but still guess wrong on values. Rejected because the SDK validates constraints for free — the incremental cost is small.

## Consequences
- **Better LLM tool use**: Descriptions and enums in the schema directly improve tool call accuracy.
- **Fail-fast validation**: Invalid inputs rejected with `InvalidParams` before hitting Bexio.
- **New types**: `TrackingType`, `SearchCriteria`, `DateString` (or similar) in `timesheet.go`.
- **`registerTool` signature change**: Accepts optional schema options. Existing callers pass `nil`.
- **Bexio enum research required**: Must verify valid values for `tracking.type` and `search_field.criteria` against Bexio API docs.
- **Risk**: If enum values are wrong or incomplete, valid inputs get rejected. Mitigated by verifying against docs and keeping enums conservative.
