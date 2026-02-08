workspace {
    model {
        developer = person "Developer" "Logs time to bexio via natural language"

        mcpClient = softwareSystem "MCP Client" "Claude Desktop, VS Code, etc." "Existing System"

        bexioMcp = softwareSystem "bexio-mcp" "MCP server exposing bexio timesheet operations as tools" {
            mcpServer = container "MCP Server" "Handles MCP protocol, tool registration, stdio transport" "Go (hand-rolled JSON-RPC)"
            bexioClient = container "Bexio API Client" "HTTP client for bexio REST API v2.0" "Go (net/http)"
        }

        bexioApi = softwareSystem "Bexio API" "api.bexio.com — REST API for timesheets, contacts, projects" "Existing System"

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
