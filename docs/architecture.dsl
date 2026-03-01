workspace {
    model {
        developer = person "Developer" "Logs time to bexio via natural language"

        mcpClient = softwareSystem "MCP Client" "Claude Desktop, VS Code, etc." "Existing System"

        bexioMcp = softwareSystem "bexio-mcp" "MCP server exposing bexio timesheet and lookup operations as tools" {
            mcpServer = container "MCP Server" "Handles MCP protocol, generic typed tool registration via registerTool[TInput,TOutput], handlers for timesheet tools (create/search/edit/delete) and lookup tools, and resolves default user_id/status_id from API lookups. SDK-generated output schemas and marshaling are used for typed timesheet outputs; search_timesheets dispatches to GET or POST based on search_fields presence." "Go (go-sdk/mcp)"
            bexioClient = container "Bexio API Client" "HTTP client wrapping bexio REST API: timesheets (v2.0), contacts (v2.0), projects (v2.0), packages (v3.0), client_services (v2.0)" "Go (net/http)"
        }

        bexioApi = softwareSystem "Bexio API" "api.bexio.com — REST API v2.0/v3.0/v4.0 for timesheets, contacts, projects, packages, client_services" "Existing System"

        developer -> mcpClient "Logs time via natural language"
        mcpClient -> mcpServer "MCP protocol over stdio" "JSON-RPC"
        mcpServer -> bexioClient "Calls bexio operations"
        bexioClient -> bexioApi "REST API calls" "HTTPS + Bearer token"
    }

    views {
        systemContext bexioMcp "SystemContext" {
            include *
            autoLayout
        }

        container bexioMcp "Containers" {
            include *
            autoLayout
        }
    }
}
