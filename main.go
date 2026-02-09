package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	bexio := NewBexioClient(os.Getenv("BEXIO_API_BASE_URL"), os.Getenv("BEXIO_API_TOKEN"), http.DefaultClient)
	server := newMCPServer(bexio)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func newMCPServer(bexio BexioClient) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "bexio-mcp", Version: "0.0.0"}, nil)
	mcp.AddTool[bexioCreateTimesheetRequest, any](server, &mcp.Tool{
		Name:        "create_timesheet",
		Description: "Create a timesheet entry in Bexio",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input bexioCreateTimesheetRequest) (*mcp.CallToolResult, any, error) {
		created, err := bexio.CreateTimesheet(ctx, input)
		if err != nil {
			return nil, nil, fmt.Errorf("create timesheet: %w", err)
		}

		result, err := marshalTimesheetResult(created)
		if err != nil {
			return nil, nil, err
		}

		return result, nil, nil
	})

	return server
}

func marshalTimesheetResult(created bexioTimesheet) (*mcp.CallToolResult, error) {
	jsonBody, err := json.Marshal(created)
	if err != nil {
		return nil, fmt.Errorf("marshal tool result: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBody)},
		},
	}, nil
}
