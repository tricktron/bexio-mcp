package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const datePatternYYYYMMDD = `^\d{4}-\d{2}-\d{2}$`

func main() {
	os.Exit(run(os.Getenv, os.Stderr))
}

func run(getenv func(string) string, stderr io.Writer) int {
	_ = getenv
	_ = stderr

	return 0
}

func validateConfig(token, baseURL string) error {
	_ = baseURL

	if token == "" {
		return errors.New("missing BEXIO_API_TOKEN")
	}

	return nil
}

func newMCPServer(bexio BexioClient) (*mcp.Server, error) {
	server := mcp.NewServer(&mcp.Implementation{Name: "bexio-mcp", Version: "0.0.0"}, nil)
	if err := registerTimesheetTools(server, bexio); err != nil {
		return nil, fmt.Errorf("register timesheet tools: %w", err)
	}
	if err := registerLookupTools(server, bexio); err != nil {
		return nil, fmt.Errorf("register lookup tools: %w", err)
	}

	return server, nil
}

func createTimesheet(ctx context.Context, bexio BexioClient, input createTimesheetInput) (bexioTimesheet, error) {
	var currentUserID int
	if input.UserID == nil {
		rawCurrentUser, err := bexio.GetCurrentUser(ctx)
		if err != nil {
			return bexioTimesheet{}, fmt.Errorf("get current user: %w", err)
		}

		currentUserID, err = parseCurrentUserID(rawCurrentUser)
		if err != nil {
			return bexioTimesheet{}, fmt.Errorf("parse current user id: %w", err)
		}
	}

	var statuses []timesheetStatus
	if input.StatusID == nil {
		rawStatuses, err := bexio.ListTimesheetStatuses(ctx)
		if err != nil {
			return bexioTimesheet{}, fmt.Errorf("list timesheet statuses: %w", err)
		}

		statuses, err = parseTimesheetStatuses(rawStatuses)
		if err != nil {
			return bexioTimesheet{}, fmt.Errorf("parse timesheet statuses: %w", err)
		}
	}

	request := resolveTimesheetDefaults(input, currentUserID, statuses)

	return bexio.CreateTimesheet(ctx, request)
}

func addTrackingDatePattern(schema *jsonschema.Schema) {
	tracking, hasTracking := schema.Properties["tracking"]
	if !hasTracking || tracking == nil {
		return
	}

	tracking.Required = filterRequiredProperties(tracking.Required, "start", "end")

	dateSchema, hasDate := tracking.Properties["date"]
	if !hasDate || dateSchema == nil {
		return
	}

	dateSchema.Pattern = datePatternYYYYMMDD
}

func filterRequiredProperties(required []string, skip ...string) []string {
	if len(required) == 0 || len(skip) == 0 {
		return required
	}

	skipSet := make(map[string]struct{}, len(skip))
	for _, key := range skip {
		skipSet[key] = struct{}{}
	}

	filtered := required[:0]
	for _, key := range required {
		if _, shouldSkip := skipSet[key]; shouldSkip {
			continue
		}
		filtered = append(filtered, key)
	}

	return filtered
}

func registerTimesheetTools(server *mcp.Server, bexio BexioClient) error {
	if err := registerTool(
		server,
		"create_timesheet",
		"Create a timesheet entry in Bexio. If user_id is omitted, defaults to the current authenticated user. If status_id is omitted, defaults to 'Erledigt' (completed).",
		func(ctx context.Context, input createTimesheetInput) (any, error) {
			return createTimesheet(ctx, bexio, input)
		},
		&jsonschema.ForOptions{TypeSchemas: timesheetTypeSchemas()},
		addTrackingDatePattern,
	); err != nil {
		return err
	}

	if err := registerTool(
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
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
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
		&jsonschema.ForOptions{TypeSchemas: timesheetTypeSchemas()},
		addTrackingDatePattern,
	); err != nil {
		return err
	}

	if err := registerTool(
		server,
		"search_timesheets",
		"Search and list timesheet entries in Bexio. If search_fields is omitted, lists all timesheets. If search_fields is provided, filters by the given field criteria. Optional date_from/date_to apply client-side date range filtering.",
		func(ctx context.Context, input bexioSearchTimesheetsRequest) (any, error) {
			entries, err := listOrSearchTimesheets(ctx, bexio, input.SearchFields)
			if err != nil {
				return nil, err
			}

			entries = filterTimesheetsByOptionalDateRange(entries, input.DateFrom, input.DateTo)

			return entries, nil
		},
		&jsonschema.ForOptions{TypeSchemas: searchTypeSchemas()},
		func(schema *jsonschema.Schema) {
			dateFromSchema, hasDateFrom := schema.Properties["date_from"]
			if hasDateFrom && dateFromSchema != nil {
				dateFromSchema.Pattern = datePatternYYYYMMDD
			}

			dateToSchema, hasDateTo := schema.Properties["date_to"]
			if hasDateTo && dateToSchema != nil {
				dateToSchema.Pattern = datePatternYYYYMMDD
			}
		},
	); err != nil {
		return err
	}

	return nil
}

func listOrSearchTimesheets(
	ctx context.Context,
	bexio BexioClient,
	searchFields []bexioSearchField,
) ([]bexioTimesheet, error) {
	if len(searchFields) == 0 {
		entries, err := bexio.ListTimesheets(ctx)
		if err != nil {
			return nil, fmt.Errorf("list timesheets: %w", err)
		}

		return entries, nil
	}

	entries, err := bexio.SearchTimesheets(ctx, searchFields)
	if err != nil {
		return nil, fmt.Errorf("search timesheets: %w", err)
	}

	return entries, nil
}

func filterTimesheetsByOptionalDateRange(entries []bexioTimesheet, dateFrom, dateTo *string) []bexioTimesheet {
	if dateFrom == nil && dateTo == nil {
		return entries
	}

	from, to := "", ""
	if dateFrom != nil {
		from = *dateFrom
	}
	if dateTo != nil {
		to = *dateTo
	}

	return filterTimesheetsByDateRange(entries, from, to)
}

//nolint:gocognit,funlen // Intentionally inlined tool registrations for locality.
func registerLookupTools(server *mcp.Server, bexio BexioClient) error {
	if err := registerTool(
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
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
		server,
		"list_projects",
		"List projects in Bexio. If contact_id is provided, filters projects by that contact.",
		func(ctx context.Context, input bexioListProjectsRequest) (any, error) {
			if input.ContactID == nil {
				projects, err := bexio.ListProjects(ctx)
				if err != nil {
					return nil, fmt.Errorf("list projects: %w", err)
				}

				return projects, nil
			}

			projects, err := bexio.SearchProjects(ctx, *input.ContactID)
			if err != nil {
				return nil, fmt.Errorf("search projects: %w", err)
			}

			return projects, nil
		},
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
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
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
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
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
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
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
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
		nil,
	); err != nil {
		return err
	}

	return nil
}

func registerTool[TInput any](
	server *mcp.Server,
	name string,
	description string,
	handler func(ctx context.Context, input TInput) (any, error),
	schemaOpts *jsonschema.ForOptions,
	schemaMutate ...func(*jsonschema.Schema),
) error {
	tool := &mcp.Tool{
		Name:        name,
		Description: description,
	}

	if schemaOpts != nil {
		schema, err := jsonschema.ForType(reflect.TypeFor[TInput](), schemaOpts)
		if err != nil {
			return fmt.Errorf("generate input schema for %s: %w", name, err)
		}

		for _, mutate := range schemaMutate {
			if mutate != nil {
				mutate(schema)
			}
		}
		tool.InputSchema = schema
	}

	mcp.AddTool[TInput, any](
		server,
		tool,
		func(ctx context.Context, _ *mcp.CallToolRequest, input TInput) (*mcp.CallToolResult, any, error) {
			payload, err := handler(ctx, input)
			if err != nil {
				return nil, nil, err
			}

			result, err := marshalToolResult(payload)
			if err != nil {
				return nil, nil, err
			}

			return result, nil, nil
		},
	)

	return nil
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
