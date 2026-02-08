# Slice: Lookup Tools (Contacts, Projects, Packages, Services)

## User Story
As a developer, I want to look up contacts, projects, packages, and services so that the LLM can resolve natural language ("BIT", "NOP project") to the correct bexio IDs when creating timesheet entries.

## Outer Boundary
- Entry: MCP tools `list_contacts`, `list_projects`, `list_packages`, `list_client_services`
- Test file: `mcp_test.go`
- Framework: `go test`

## Shell Boundaries
- HTTP: BexioClient — calls GET endpoints for contacts, projects, packages, client_services

## Functional Core
Discovered via TDD.

## Acceptance Criterion
Given a running MCP server  
When the client calls `list_contacts`  
Then the tool returns a list of contacts with id and name

Given a running MCP server  
When the client calls `list_projects` with an optional contact_id filter  
Then the tool returns projects associated with that contact

Given a running MCP server  
When the client calls `list_packages` with a project_id  
Then the tool returns work packages for that project

Given a running MCP server  
When the client calls `list_client_services`  
Then the tool returns available business activities (Tätigkeiten)

## API Reference
- `docs/bexio-api-contact.json` — GET /2.0/contact
- `docs/bexio-api-project.json` — GET /2.0/pr_project
- `docs/bexio-api-package.json` — GET /3.0/projects/{project_id}/packages
- `docs/bexio-api-client_service.json` — GET /2.0/client_service

## Uses
- BexioClient from slice 01
