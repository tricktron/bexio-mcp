package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type config struct {
	apiToken   string
	apiBaseURL string
}

func main() {
	cfg := config{
		apiToken:   os.Getenv("BEXIO_API_TOKEN"),
		apiBaseURL: os.Getenv("BEXIO_API_BASE_URL"),
	}

	bexio := NewBexioClient(cfg.apiBaseURL, cfg.apiToken, http.DefaultClient)
	server := newMCPServer(bexio)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func newMCPServer(bexio BexioClient) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "bexio-mcp", Version: "0.0.0"}, nil)
	mcp.AddTool[bexioCreateTimesheetRequest, any](server, &mcp.Tool{Name: "create_timesheet"}, func(ctx context.Context, req *mcp.CallToolRequest, input bexioCreateTimesheetRequest) (*mcp.CallToolResult, any, error) {
		_ = req

		created, err := bexio.CreateTimesheet(ctx, input)
		if err != nil {
			return nil, nil, err
		}

		jsonBody, err := json.Marshal(created)
		if err != nil {
			return nil, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBody)},
			},
		}, nil, nil
	})

	return server
}
