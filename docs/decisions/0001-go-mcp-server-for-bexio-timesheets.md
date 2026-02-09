# 0001: Go MCP Server for Bexio Timesheets

## Status
Accepted

## Context
We need to automate bexio timesheet logging via natural language. The developer logs time daily with mostly static fields (Tätigkeit, Ansprechpartner, Status, Abrechenbar) and a few dynamic fields (Bemerkung, Dauer, Datum, Kontakt → Kontaktperson → Projekt → Arbeitspaket).

An MCP server exposes bexio timesheet operations as tools, enabling any MCP client (Claude Desktop, VS Code, etc.) to interactively log time via natural language.

## Decision
Build the MCP server in **Go** using `github.com/modelcontextprotocol/go-sdk/mcp` for protocol handling and stdio transport.

> **Update (ADR 0002):** Tried hand-rolled JSON-RPC 2.0 first. It passed unit tests but failed against a real MCP client (opencode timeout). Reverted to SDK. See ADR 0002 for full history.

### Why Go over TypeScript
- Single binary distribution — no runtime dependency, easy to deploy behind corporate proxy
- `net/http` stdlib — no third-party HTTP library needed for bexio API calls
- Struct types serve as both MCP tool schemas and HTTP request/response bodies
- Fast compile + test cycle for iterative API integration work

### Why Go over Python
- Corporate MITM proxy (BIT Proxy CA) generates certificates missing the Authority Key Identifier extension
- Python 3.13+ added `VERIFY_X509_STRICT` as default SSL flag, which rejects these certificates
- Every Python HTTP library would need SSL context workarounds — fragile and error-prone
- Go's TLS implementation does not have this strictness issue

### Why Go over Rust
- I/O-bound workload (HTTP calls to bexio API) — Rust's zero-cost abstractions provide no benefit
- Faster compile times for iterative development
- Simpler error handling for REST API wrappers
- Lower learning curve for the same end result

## Alternatives Considered

| Alternative       | Why rejected                                          |
| ----------------- | ----------------------------------------------------- |
| TypeScript SDK    | v2 alpha, npm dependency chain, npx proxy issues      |
| Python (FastMCP)  | Python 3.13 SSL breakage with corporate MITM proxy    |
| Rust (rmcp)       | Unnecessary complexity for I/O-bound API wrapper      |
| CLI tool (no MCP) | No natural language interaction, no learning MCP goal |

## Diagram
```mermaid
graph LR
    User[Developer] -->|natural language| Client[MCP Client<br>Claude Desktop / VS Code]
    Client -->|MCP protocol<br>stdio| Server[bexio-mcp<br>Go binary]
    Server -->|REST API<br>Bearer token| Bexio[api.bexio.com]
```

## Consequences
- Go module using `github.com/modelcontextprotocol/go-sdk/mcp` for MCP protocol
- Single binary, distributed via `go install` or direct binary
- Bearer token auth (static API token from bexio settings)
- API spec files in `docs/bexio-api-*.json` for reference during implementation
- Struct types double as MCP schemas and HTTP request/response bodies
