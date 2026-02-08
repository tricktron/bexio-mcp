# Slice: README and AGENTS.md

## User Story
As a developer, I want a README so that humans know what this project does and how to use it.
As an LLM agent, I want an AGENTS.md so that I understand the project conventions and can contribute effectively.

## Outer Boundary
- Entry: N/A (documentation files)
- Test file: N/A
- Framework: N/A

## Shell Boundaries
- FS: `README.md`, `AGENTS.md` — new files at project root

## Functional Core
None — documentation only.

## Tasks

### README.md
- What the project does (one paragraph)
- Prerequisites (Go version, Bexio API token)
- Build & run instructions
- Configuration (env vars: BEXIO_API_TOKEN, BEXIO_API_BASE_URL)
- MCP client configuration example (opencode, Claude Desktop)
- Available tools (currently: create_timesheet)
- Development: how to run tests, lint

### AGENTS.md
- Project architecture summary (hand-rolled JSON-RPC, functional core/imperative shell)
- File conventions (main.go = CLI shell, bexio_client.go = HTTP shell, timesheet.go = domain types)
- Testing patterns (subprocess acceptance test, httptest unit tests)
- Key decisions (hand-rolled MCP, no SDK, interfaces grow per-slice)
- golangci-lint: must pass before committing
- Pointers to ADRs, architecture.dsl, workflow slices

## Acceptance Criterion
Given a new developer (or LLM agent) visiting the repo  
When they read README.md  
Then they can build, configure, and run the MCP server

Given an LLM agent working on this codebase  
When it reads AGENTS.md  
Then it understands the conventions, patterns, and constraints

## Uses
- All existing project knowledge from slice 01 + ADRs
