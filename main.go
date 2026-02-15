package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const datePatternYYYYMMDD = `^\d{4}-\d{2}-\d{2}$`

func main() {
	os.Exit(run(os.Getenv, os.Stderr, &mcp.StdioTransport{}))
}

func run(getenv func(string) string, stderr io.Writer, transport mcp.Transport) int {
	token := getenv("BEXIO_API_TOKEN")
	baseURL := getenv("BEXIO_API_BASE_URL")

	if err := validateConfig(token, baseURL); err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	if transport == nil {
		return 0
	}

	bexio := NewBexioClient(baseURL, token, http.DefaultClient)
	server, err := newMCPServer(bexio)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	session, err := server.Connect(context.Background(), transport, nil)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	if waitErr := session.Wait(); waitErr != nil {
		fmt.Fprintf(stderr, "%v\n", waitErr)
		return 1
	}

	return 0
}

func validateConfig(token, baseURL string) error {
	if token == "" {
		return errors.New("missing BEXIO_API_TOKEN")
	}

	if baseURL == "" {
		return errors.New("missing BEXIO_API_BASE_URL")
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
		currentUser, err := bexio.GetCurrentUser(ctx)
		if err != nil {
			return bexioTimesheet{}, fmt.Errorf("get current user: %w", err)
		}

		currentUserID = currentUser.ID
	}

	var statuses []timesheetStatus
	if input.StatusID == nil {
		resolvedStatuses, err := bexio.ListTimesheetStatuses(ctx)
		if err != nil {
			return bexioTimesheet{}, fmt.Errorf("list timesheet statuses: %w", err)
		}

		statuses = resolvedStatuses
	}

	request := resolveTimesheetDefaults(input, currentUserID, statuses)

	return bexio.CreateTimesheet(ctx, request)
}

func deleteTimesheet(ctx context.Context, bexio BexioClient, id int) (deleteTimesheetResult, error) {
	deleted, err := bexio.DeleteTimesheet(ctx, id)
	if err != nil {
		return deleteTimesheetResult{}, fmt.Errorf("delete timesheet: %w", err)
	}

	var result deleteTimesheetResult
	if parseErr := json.Unmarshal(deleted, &result); parseErr != nil {
		return deleteTimesheetResult{}, fmt.Errorf("parse delete result: %w", parseErr)
	}

	return result, nil
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
		func(ctx context.Context, input createTimesheetInput) (bexioTimesheet, error) {
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
		func(ctx context.Context, input bexioDeleteTimesheetRequest) (deleteTimesheetResult, error) {
			return deleteTimesheet(ctx, bexio, input.ID)
		},
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
		server,
		"edit_timesheet",
		"Edit a timesheet entry in Bexio",
		func(ctx context.Context, input bexioEditTimesheetRequest) (bexioTimesheet, error) {
			updated, err := bexio.EditTimesheet(ctx, input.ID, input.bexioCreateTimesheetRequest)
			if err != nil {
				return bexioTimesheet{}, fmt.Errorf("edit timesheet: %w", err)
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
		func(ctx context.Context, input bexioSearchTimesheetsRequest) (searchTimesheetsResult, error) {
			entries, err := listOrSearchTimesheets(ctx, bexio, input.SearchFields)
			if err != nil {
				return searchTimesheetsResult{}, err
			}

			entries = filterTimesheetsByOptionalDateRange(entries, input.DateFrom, input.DateTo)

			return searchTimesheetsResult{Results: entries}, nil
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
		func(ctx context.Context, _ struct{}) (listContactsResult, error) {
			contacts, err := bexio.ListContacts(ctx)
			if err != nil {
				return listContactsResult{}, fmt.Errorf("list contacts: %w", err)
			}

			return listContactsResult{Contacts: contacts}, nil
		},
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
		server,
		"list_projects",
		"List projects in Bexio. If contact_id is provided, filters projects by that contact.",
		func(ctx context.Context, input bexioListProjectsRequest) (listProjectsResult, error) {
			if input.ContactID == nil {
				projects, err := bexio.ListProjects(ctx)
				if err != nil {
					return listProjectsResult{}, fmt.Errorf("list projects: %w", err)
				}

				return listProjectsResult{Projects: projects}, nil
			}

			projects, err := bexio.SearchProjects(ctx, *input.ContactID)
			if err != nil {
				return listProjectsResult{}, fmt.Errorf("search projects: %w", err)
			}

			return listProjectsResult{Projects: projects}, nil
		},
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
		server,
		"list_client_services",
		"List client services in Bexio",
		func(ctx context.Context, _ struct{}) (listServicesResult, error) {
			services, err := bexio.ListClientServices(ctx)
			if err != nil {
				return listServicesResult{}, fmt.Errorf("list client services: %w", err)
			}

			return listServicesResult{Services: services}, nil
		},
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
		server,
		"list_packages",
		"List work packages for a project in Bexio",
		func(ctx context.Context, input bexioListProjectPackagesRequest) (listPackagesResult, error) {
			packages, err := bexio.ListPackages(ctx, input.ProjectID)
			if err != nil {
				return listPackagesResult{}, fmt.Errorf("list packages: %w", err)
			}

			return listPackagesResult{Packages: packages}, nil
		},
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
		server,
		"list_timesheet_statuses",
		"List timesheet statuses in Bexio",
		func(ctx context.Context, _ struct{}) (listStatusesResult, error) {
			statuses, err := bexio.ListTimesheetStatuses(ctx)
			if err != nil {
				return listStatusesResult{}, fmt.Errorf("list timesheet statuses: %w", err)
			}

			return listStatusesResult{Statuses: statuses}, nil
		},
		nil,
	); err != nil {
		return err
	}

	if err := registerTool(
		server,
		"get_current_user",
		"Get current authenticated user in Bexio",
		func(ctx context.Context, _ struct{}) (bexioUser, error) {
			currentUser, err := bexio.GetCurrentUser(ctx)
			if err != nil {
				return bexioUser{}, fmt.Errorf("get current user: %w", err)
			}

			return currentUser, nil
		},
		nil,
	); err != nil {
		return err
	}

	return nil
}

func registerTool[TInput, TOutput any](
	server *mcp.Server,
	name string,
	description string,
	handler func(ctx context.Context, input TInput) (TOutput, error),
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

	mcp.AddTool[TInput, TOutput](
		server,
		tool,
		func(ctx context.Context, _ *mcp.CallToolRequest, input TInput) (*mcp.CallToolResult, TOutput, error) {
			result, err := handler(ctx, input)
			if err != nil {
				var zero TOutput
				return nil, zero, err
			}

			return nil, result, nil
		},
	)

	return nil
}
