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
	registerTimesheetTools(server, bexio)
	registerLookupTools(server, bexio)

	return server
}

func registerTimesheetTools(server *mcp.Server, bexio BexioClient) {
	registerTool(
		server,
		"create_timesheet",
		"Create a timesheet entry in Bexio",
		func(ctx context.Context, input createTimesheetInput) (any, error) {
			request := bexioCreateTimesheetRequest{
				AllowableBill:   input.AllowableBill,
				ClientServiceID: input.ClientServiceID,
				Tracking:        input.Tracking,
				Text:            input.Text,
				ContactID:       input.ContactID,
				PrProjectID:     input.PrProjectID,
			}
			if input.UserID != nil {
				request.UserID = *input.UserID
			}
			if input.StatusID != nil {
				request.StatusID = *input.StatusID
			}

			created, err := bexio.CreateTimesheet(ctx, request)
			if err != nil {
				return nil, fmt.Errorf("create timesheet: %w", err)
			}

			return created, nil
		},
	)
	registerTool(
		server,
		"delete_timesheet",
		"Delete a timesheet entry in Bexio",
		func(ctx context.Context, input bexioDeleteTimesheetRequest) (any, error) {
			deleted, err := bexio.DeleteTimesheet(ctx, input.ID)
			if err != nil {
				return nil, fmt.Errorf("delete timesheet: %w", err)
			}

			return deleted, nil
		},
	)
	registerTool(
		server,
		"edit_timesheet",
		"Edit a timesheet entry in Bexio",
		func(ctx context.Context, input bexioEditTimesheetRequest) (any, error) {
			updated, err := bexio.EditTimesheet(ctx, input.ID, input.bexioCreateTimesheetRequest)
			if err != nil {
				return nil, fmt.Errorf("edit timesheet: %w", err)
			}

			return updated, nil
		},
	)
	registerTool(
		server,
		"list_timesheets",
		"List timesheet entries in Bexio",
		func(ctx context.Context, _ struct{}) (any, error) {
			entries, err := bexio.ListTimesheets(ctx)
			if err != nil {
				return nil, fmt.Errorf("list timesheets: %w", err)
			}

			return entries, nil
		},
	)
	registerTool(
		server,
		"search_timesheets",
		"Search timesheet entries in Bexio",
		func(ctx context.Context, input bexioSearchTimesheetsRequest) (any, error) {
			entries, err := bexio.SearchTimesheets(ctx, input.SearchFields)
			if err != nil {
				return nil, fmt.Errorf("search timesheets: %w", err)
			}

			return entries, nil
		},
	)
}

func registerLookupTools(server *mcp.Server, bexio BexioClient) {
	registerTool(
		server,
		"list_contacts",
		"List contacts in Bexio",
		func(ctx context.Context, _ struct{}) (any, error) {
			contacts, err := bexio.ListContacts(ctx)
			if err != nil {
				return nil, fmt.Errorf("list contacts: %w", err)
			}

			return contacts, nil
		},
	)
	registerTool(
		server,
		"list_projects",
		"List projects in Bexio",
		func(ctx context.Context, input bexioListProjectsRequest) (any, error) {
			projects, err := bexio.SearchProjects(ctx, input.ContactID)
			if err != nil {
				return nil, fmt.Errorf("list projects: %w", err)
			}

			return projects, nil
		},
	)
	registerTool(
		server,
		"list_client_services",
		"List client services in Bexio",
		func(ctx context.Context, _ struct{}) (any, error) {
			services, err := bexio.ListClientServices(ctx)
			if err != nil {
				return nil, fmt.Errorf("list client services: %w", err)
			}

			return services, nil
		},
	)
	registerTool(
		server,
		"list_packages",
		"List work packages for a project in Bexio",
		func(ctx context.Context, input bexioListProjectPackagesRequest) (any, error) {
			packages, err := bexio.ListPackages(ctx, input.ProjectID)
			if err != nil {
				return nil, fmt.Errorf("list packages: %w", err)
			}

			return packages, nil
		},
	)
	registerTool(
		server,
		"list_timesheet_statuses",
		"List timesheet statuses in Bexio",
		func(ctx context.Context, _ struct{}) (any, error) {
			statuses, err := bexio.ListTimesheetStatuses(ctx)
			if err != nil {
				return nil, fmt.Errorf("list timesheet statuses: %w", err)
			}

			return statuses, nil
		},
	)
	registerTool(
		server,
		"get_current_user",
		"Get current authenticated user in Bexio",
		func(ctx context.Context, _ struct{}) (any, error) {
			currentUser, err := bexio.GetCurrentUser(ctx)
			if err != nil {
				return nil, fmt.Errorf("get current user: %w", err)
			}

			return currentUser, nil
		},
	)
}

func registerTool[TInput any](
	server *mcp.Server,
	name string,
	description string,
	handler func(ctx context.Context, input TInput) (any, error),
) {
	mcp.AddTool[TInput, any](server, &mcp.Tool{
		Name:        name,
		Description: description,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input TInput) (*mcp.CallToolResult, any, error) {
		payload, err := handler(ctx, input)
		if err != nil {
			return nil, nil, err
		}

		result, err := marshalToolResult(payload)
		if err != nil {
			return nil, nil, err
		}

		return result, nil, nil
	})
}

func marshalToolResult(payload any) (*mcp.CallToolResult, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal tool result: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBody)},
		},
	}, nil
}
