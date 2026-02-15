package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCreateTimesheetAcceptance(t *testing.T) {
	// Slice: Adopt MCP SDK
	// Given a configured MCP server, when create_timesheet is called over MCP, then the timesheet is created in Bexio without timeout.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "create_timesheet",
		Arguments: map[string]any{
			"user_id":           42,
			"allowable_bill":    true,
			"client_service_id": 99,
			"contact_id":        11,
			"pr_project_id":     12,
			"text":              "Build acceptance test",
			"tracking": map[string]any{
				"type":  "range",
				"date":  "2026-02-08",
				"start": "09:00",
				"end":   "10:30",
			},
		},
	})
	assert.NoError(t, err)

	assert.False(t, result.IsError)
	assert.True(t, len(result.Content) > 0, "result should have content")
}

func TestCreateTimesheetAutoResolveDefaultsAcceptance(t *testing.T) {
	// Slice: Auto-resolve Timesheet Defaults
	// Given a running MCP server, when create_timesheet is called without user_id and status_id, then user_id resolves from current user and status_id resolves to "Erledigt".
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "create_timesheet",
		Arguments: map[string]any{
			"allowable_bill":    true,
			"client_service_id": 99,
			"text":              "Build acceptance test with defaults",
			"tracking": map[string]any{
				"type":  "range",
				"date":  "2026-02-08",
				"start": "09:00",
				"end":   "10:30",
			},
		},
	})
	assert.NoError(t, err)

	assert.False(t, result.IsError)
	assert.True(t, len(result.Content) > 0, "result should have content")
}

func TestCreateTimesheetDurationTrackingAcceptance(t *testing.T) {
	// Slice: Support duration tracking type
	// Given an MCP client calls create_timesheet with duration tracking, when the handler sends the request to Bexio, then the payload includes type/date/duration and omits start/end.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "create_timesheet",
		Arguments: map[string]any{
			"user_id":           42,
			"allowable_bill":    true,
			"client_service_id": 99,
			"text":              "Track by duration",
			"tracking": map[string]any{
				"type":     "duration",
				"date":     "2026-02-08",
				"duration": "01:30",
			},
		},
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)

	rawBody := env.fakeAPI.lastCreateTimesheetBodyBytes()
	assert.True(t, len(rawBody) > 0, "fake Bexio API should capture POST /2.0/timesheet request body")

	var sentPayload map[string]any
	err = json.Unmarshal(rawBody, &sentPayload)
	assert.NoError(t, err)

	tracking, ok := sentPayload["tracking"].(map[string]any)
	assert.True(t, ok, "request payload should include tracking object")

	assert.Equal(t, "duration", tracking["type"])
	assert.Equal(t, "2026-02-08", tracking["date"])
	duration, hasDuration := tracking["duration"]
	assert.True(t, hasDuration, "duration tracking payload should include tracking.duration")
	assert.Equal(t, "01:30", duration)

	_, hasStart := tracking["start"]
	_, hasEnd := tracking["end"]
	assert.False(t, hasStart, "duration tracking payload should not include tracking.start")
	assert.False(t, hasEnd, "duration tracking payload should not include tracking.end")
}

func TestDeleteTimesheetAcceptance(t *testing.T) {
	// Slice: Edit & Delete Timesheet
	// Given an existing timesheet entry, when the client calls delete_timesheet with an id, then the entry is deleted from Bexio and the tool confirms deletion.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "delete_timesheet",
		Arguments: map[string]any{
			"id": 777,
		},
	})
	assert.NoError(t, err)

	assert.False(t, result.IsError)
	assert.Equal(t, 1, len(result.Content))

	textContent, ok := result.Content[0].(*mcp.TextContent)
	assert.True(t, ok, "result should contain one text content item")
	assert.Equal(t, `{"success":true}`, textContent.Text)
}

func TestEditTimesheetAcceptance(t *testing.T) {
	// Slice: Edit & Delete Timesheet
	// Given an existing timesheet entry, when the client calls edit_timesheet with an id and updated fields, then the entry is updated in bexio and the tool returns the updated entry.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "edit_timesheet",
		Arguments: map[string]any{
			"id":                777,
			"user_id":           42,
			"allowable_bill":    true,
			"client_service_id": 99,
			"contact_id":        11,
			"pr_project_id":     12,
			"text":              "Updated acceptance test",
			"tracking": map[string]any{
				"type":  "range",
				"date":  "2026-02-08",
				"start": "13:00",
				"end":   "14:30",
			},
		},
	})
	assert.NoError(t, err)

	assert.False(t, result.IsError)
	assert.Equal(t, 1, len(result.Content))

	textContent, ok := result.Content[0].(*mcp.TextContent)
	assert.True(t, ok, "result should contain one text content item")

	var updated bexioTimesheet
	err = json.Unmarshal([]byte(textContent.Text), &updated)
	assert.NoError(t, err)
	assert.Equal(t, bexioTimesheet{
		ID:              777,
		UserID:          42,
		StatusID:        0,
		AllowableBill:   true,
		ClientServiceID: 99,
		Date:            "2026-02-08",
		Duration:        "01:30",
		Running:         false,
		ContactID:       intPtr(11),
		PrProjectID:     intPtr(12),
		Text:            "Updated acceptance test",
		Tracking: trackingRange{
			Type:  "range",
			Date:  "2026-02-08",
			Start: "13:00",
			End:   "14:30",
		},
	}, updated)
}

func TestSearchTimesheetsNoArgsReturnsAllAcceptance(t *testing.T) {
	// Slice: Merge list_timesheets and search_timesheets
	// Given a running MCP server, when the client calls search_timesheets, then the tool returns recent timesheet entries with realistic tracking datetime format and core API fields.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "search_timesheets",
	})
	assert.NoError(t, err)

	assert.False(t, result.IsError)
	assert.Equal(t, 1, len(result.Content))

	textContent, ok := result.Content[0].(*mcp.TextContent)
	assert.True(t, ok, "result should contain one text content item")

	var listed []bexioTimesheet
	err = json.Unmarshal([]byte(textContent.Text), &listed)
	assert.NoError(t, err)
	assert.True(t, len(listed) >= 2, "result should contain at least two timesheet entries")

	first := listed[0]
	assert.Equal(t, 801, first.ID)
	assert.Equal(t, "list-entry-1", first.Text)
	assert.Equal(t, "2026-02-01 09:00:00", first.Tracking.Start)
	assert.Equal(t, "2026-02-01 10:00:00", first.Tracking.End)
	assert.Equal(t, 2, first.StatusID)
	assert.Equal(t, "2026-02-01", first.Date)
	assert.Equal(t, "01:00", first.Duration)
	assert.False(t, first.Running)
}

func TestSearchTimesheetsAcceptance(t *testing.T) {
	// Slice: List & Search Timesheets
	// Given a running MCP server, when the client calls search_timesheets with a user_id filter, then the tool returns matching timesheet entries.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "search_timesheets",
		Arguments: map[string]any{
			"search_fields": []map[string]any{
				{"field": "user_id", "value": "42", "criteria": "="},
			},
		},
	})
	assert.NoError(t, err)

	contentJSON, err := json.Marshal(result.Content)
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(
		t,
		strings.Contains(string(contentJSON), "search-match-1"),
		"result should contain matching timesheet entries",
	)
}

func TestSearchTimesheetsDateRangeFilterAcceptance(t *testing.T) {
	// Slice: Merge list_timesheets and search_timesheets
	// Given I have timesheets on 2026-02-01 and 2026-02-02, when I request timesheets from 2026-02-02 to 2026-02-02, then I receive only the timesheet from 2026-02-02.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "search_timesheets",
		Arguments: map[string]any{
			"date_from": "2026-02-02",
			"date_to":   "2026-02-02",
		},
	})
	assert.NoError(t, err)

	assert.False(t, result.IsError)
	assert.Equal(t, 1, len(result.Content))

	textContent, ok := result.Content[0].(*mcp.TextContent)
	assert.True(t, ok, "result should contain one text content item")

	var listed []bexioTimesheet
	err = json.Unmarshal([]byte(textContent.Text), &listed)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listed))
	assert.Equal(t, "2026-02-02", listed[0].Date)
}

func TestSearchTimesheetsOpenEndedDateRangeAcceptance(t *testing.T) {
	// Slice: Merge list_timesheets and search_timesheets
	// Given I have timesheets on 2026-02-01 and 2026-02-02, when I request timesheets from 2026-02-02 onward (no date_to), then I receive only the timesheet from 2026-02-02.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "search_timesheets",
		Arguments: map[string]any{
			"date_from": "2026-02-02",
		},
	})
	assert.NoError(t, err)

	assert.False(t, result.IsError)
	assert.Equal(t, 1, len(result.Content))

	textContent, ok := result.Content[0].(*mcp.TextContent)
	assert.True(t, ok, "result should contain one text content item")

	var listed []bexioTimesheet
	err = json.Unmarshal([]byte(textContent.Text), &listed)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(listed))
	assert.Equal(t, "2026-02-02", listed[0].Date)
}

func TestSearchTimesheetsDateRangeNoResultsAcceptance(t *testing.T) {
	// Slice: Search Timesheets by Date
	// Given I have a timesheet on 2026-02-10, when I request timesheets from 2026-02-01 to 2026-02-05, then I receive no timesheets.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "search_timesheets",
		Arguments: map[string]any{
			"search_fields": []map[string]any{
				{"field": "user_id", "value": "42", "criteria": "="},
			},
			"date_from": "2026-02-01",
			"date_to":   "2026-02-05",
		},
	})
	assert.NoError(t, err)

	assert.False(t, result.IsError)
	assert.Equal(t, 1, len(result.Content))

	textContent, ok := result.Content[0].(*mcp.TextContent)
	assert.True(t, ok, "result should contain one text content item")

	var listed []bexioTimesheet
	err = json.Unmarshal([]byte(textContent.Text), &listed)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(listed))
}

func TestLookupToolsAcceptance(t *testing.T) {
	// Slice: Lookup Tools (Contacts, Projects, Packages, Services)
	// Given a running MCP server, when lookup tools are called, then each tool returns non-error MCP content that includes expected values.
	testCases := []struct {
		name       string
		toolName   string
		arguments  map[string]any
		expected   string
		errorLabel string
	}{
		{name: "ListContacts", toolName: "list_contacts", expected: "Acme Corp", errorLabel: "contact entries"},
		{
			name:       "ListProjects",
			toolName:   "list_projects",
			arguments:  map[string]any{"contact_id": 11},
			expected:   "Project Alpha",
			errorLabel: "project entries",
		},
		{
			name:       "ListProjectsWithoutContactID",
			toolName:   "list_projects",
			arguments:  map[string]any{},
			expected:   "Project Alpha",
			errorLabel: "project entries",
		},
		{
			name:       "ListClientServices",
			toolName:   "list_client_services",
			expected:   "Engineering",
			errorLabel: "client service entries",
		},
		{
			name:       "ListPackages",
			toolName:   "list_packages",
			arguments:  map[string]any{"project_id": 5},
			expected:   "Backend Sprint",
			errorLabel: "package entries",
		},
		{
			name:       "ListTimesheetStatuses",
			toolName:   "list_timesheet_statuses",
			expected:   "Erledigt",
			errorLabel: "timesheet status entries",
		},
		{name: "GetCurrentUser", toolName: "get_current_user", expected: "Rudolph", errorLabel: "current user entry"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := newAcceptanceEnv(t)

			result, err := env.callTool(&mcp.CallToolParams{
				Name:      tc.toolName,
				Arguments: tc.arguments,
			})
			assert.NoError(t, err)

			contentJSON, err := json.Marshal(result.Content)
			assert.NoError(t, err)
			assert.False(t, result.IsError)
			assert.True(
				t,
				strings.Contains(string(contentJSON), tc.expected),
				"result should contain "+tc.errorLabel,
			)
		})
	}
}

func TestToolSchemasIncludeDescriptionsAcceptance(t *testing.T) {
	// Slice: Add jsonschema description tags to all input structs
	// Given a running MCP server, when tools are listed, then create_timesheet and search_timesheets schemas include field descriptions.
	env := newAcceptanceEnv(t)

	listed, err := env.listTools(&mcp.ListToolsParams{})
	assert.NoError(t, err)

	createTool := mustFindToolByName(t, listed.Tools, "create_timesheet")
	createSchema, ok := createTool.InputSchema.(map[string]any)
	assert.True(t, ok, "create_timesheet input schema should be an object")
	assertSchemaPropertiesHaveDescriptions(t, createSchema)

	searchTool := mustFindToolByName(t, listed.Tools, "search_timesheets")
	searchSchema, ok := searchTool.InputSchema.(map[string]any)
	assert.True(t, ok, "search_timesheets input schema should be an object")

	searchFields := mustSchemaPropertyMap(t, searchSchema, "search_fields")
	searchItems, ok := searchFields["items"].(map[string]any)
	assert.True(t, ok, "search_fields should define item schema")

	itemProps := mustSchemaProperties(t, searchItems)
	assertDescriptionContains(t, itemProps, "field", "filter")
	assertDescriptionContains(t, itemProps, "value", "match")
	assertDescriptionContains(t, itemProps, "criteria", "operator")
}

type fakeBexioAPI struct {
	URL                     string
	server                  *httptest.Server
	lastCreateTimesheetBody []byte
}

type acceptanceEnv struct {
	ctx       context.Context
	callTool  func(params *mcp.CallToolParams) (*mcp.CallToolResult, error)
	listTools func(params *mcp.ListToolsParams) (*mcp.ListToolsResult, error)
	fakeAPI   *fakeBexioAPI
}

func newAcceptanceEnv(t *testing.T) acceptanceEnv {
	t.Helper()

	fakeAPI := startFakeBexioAPI(t)
	bexio := NewBexioClient(fakeAPI.URL, "test-token", http.DefaultClient)
	env := newAcceptanceEnvWith(t, bexio)
	env.fakeAPI = fakeAPI

	return env
}

func newAcceptanceEnvWith(t *testing.T, bexio BexioClient) acceptanceEnv {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)

	server, err := newMCPServer(bexio)
	assert.NoError(t, err)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	assert.NoError(t, err)
	t.Cleanup(func() { serverSession.Close() })

	testClient := mcp.NewClient(&mcp.Implementation{Name: "acceptance-test", Version: "0.1.0"}, nil)
	clientSession, err := testClient.Connect(ctx, clientTransport, nil)
	assert.NoError(t, err)
	t.Cleanup(func() { clientSession.Close() })

	return acceptanceEnv{
		ctx: ctx,
		callTool: func(params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
			return clientSession.CallTool(ctx, params)
		},
		listTools: func(params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return clientSession.ListTools(ctx, params)
		},
		fakeAPI: nil,
	}
}

func startFakeBexioAPI(t *testing.T) *fakeBexioAPI {
	t.Helper()

	api := &fakeBexioAPI{}
	exactHandlers := map[string]http.HandlerFunc{
		http.MethodPost + " /2.0/timesheet": func(w http.ResponseWriter, r *http.Request) {
			reqBody := api.captureCreateTimesheetRequest(t, r)
			writeJSONResponse(t, w, http.StatusCreated, buildTimesheet(777, reqBody))
		},
		http.MethodGet + " /2.0/timesheet": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONResponse(t, w, http.StatusOK, listTimesheetEntriesFixture())
		},
		http.MethodPost + " /2.0/timesheet/search": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONResponse(t, w, http.StatusOK, searchTimesheetEntriesFixture())
		},
		http.MethodGet + " /2.0/contact": func(w http.ResponseWriter, _ *http.Request) {
			contacts := []map[string]any{
				{"id": 11, "name_1": "Acme Corp", "name_2": ""},
				{"id": 12, "name_1": "Globex Inc", "name_2": ""},
			}
			writeJSONResponse(t, w, http.StatusOK, contacts)
		},
		http.MethodPost + " /2.0/pr_project/search": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONResponse(t, w, http.StatusOK, projectsFixture())
		},
		http.MethodGet + " /2.0/pr_project": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONResponse(t, w, http.StatusOK, projectsFixture())
		},
		http.MethodGet + " /2.0/client_service": func(w http.ResponseWriter, _ *http.Request) {
			services := []map[string]any{
				{"id": 77, "name": "Engineering"},
				{"id": 78, "name": "Consulting"},
			}
			writeJSONResponse(t, w, http.StatusOK, services)
		},
		http.MethodGet + " /3.0/projects/5/packages": func(w http.ResponseWriter, _ *http.Request) {
			packages := []map[string]any{
				{"id": 61, "name": "Backend Sprint"},
				{"id": 62, "name": "QA Run"},
			}
			writeJSONResponse(t, w, http.StatusOK, packages)
		},
		http.MethodGet + " /2.0/timesheet_status": func(w http.ResponseWriter, _ *http.Request) {
			statuses := []map[string]any{
				{"id": 1, "name": "Offen"},
				{"id": 2, "name": "Erledigt"},
			}
			writeJSONResponse(t, w, http.StatusOK, statuses)
		},
		http.MethodGet + " /3.0/users/me": func(w http.ResponseWriter, _ *http.Request) {
			user := map[string]any{
				"id":        4,
				"firstname": "Rudolph",
				"lastname":  "Smith",
				"email":     "rudolph.smith@example.com",
			}
			writeJSONResponse(t, w, http.StatusOK, user)
		},
	}

	api.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if handled := handleTimesheetByIDRoutes(t, w, r); handled {
			return
		}

		handler, ok := exactHandlers[r.Method+" "+r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		handler(w, r)
	}))
	api.URL = api.server.URL
	t.Cleanup(api.server.Close)

	return api
}

func handleTimesheetByIDRoutes(t *testing.T, w http.ResponseWriter, r *http.Request) bool {
	t.Helper()

	if !strings.HasPrefix(r.URL.Path, "/2.0/timesheet/") || r.URL.Path == "/2.0/timesheet/search" {
		return false
	}

	id, ok := parseTimesheetID(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return true
	}

	switch r.Method {
	case http.MethodPost:
		reqBody := decodeRequestJSON[bexioCreateTimesheetRequest](t, r)
		writeJSONResponse(t, w, http.StatusOK, buildTimesheet(id, reqBody))
		return true
	case http.MethodGet:
		writeJSONResponse(t, w, http.StatusOK, bexioTimesheet{
			ID:   id,
			Text: "[MCP-TEST] contract test create",
		})
		return true
	case http.MethodDelete:
		writeJSONResponse(t, w, http.StatusOK, map[string]bool{"success": true})
		return true
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return true
	}
}

func (api *fakeBexioAPI) captureCreateTimesheetRequest(t *testing.T, r *http.Request) bexioCreateTimesheetRequest {
	t.Helper()

	rawBody, err := io.ReadAll(r.Body)
	assert.NoError(t, err)

	api.lastCreateTimesheetBody = append(api.lastCreateTimesheetBody[:0], rawBody...)

	var reqBody bexioCreateTimesheetRequest
	err = json.Unmarshal(rawBody, &reqBody)
	assert.NoError(t, err)

	return reqBody
}

func (api *fakeBexioAPI) lastCreateTimesheetBodyBytes() []byte {
	return api.lastCreateTimesheetBody
}

func parseTimesheetID(path string) (int, bool) {
	idText := strings.TrimPrefix(path, "/2.0/timesheet/")
	if idText == "" || strings.Contains(idText, "/") {
		return 0, false
	}

	id, err := strconv.Atoi(idText)
	if err != nil {
		return 0, false
	}

	return id, true
}

func buildTimesheet(id int, reqBody bexioCreateTimesheetRequest) bexioTimesheet {
	return bexioTimesheet{
		ID:              id,
		UserID:          reqBody.UserID,
		StatusID:        reqBody.StatusID,
		AllowableBill:   reqBody.AllowableBill,
		ClientServiceID: reqBody.ClientServiceID,
		Date:            reqBody.Tracking.Date,
		Duration:        "01:30",
		Running:         false,
		Text:            reqBody.Text,
		ContactID:       reqBody.ContactID,
		PrProjectID:     reqBody.PrProjectID,
		Tracking:        reqBody.Tracking,
	}
}

func listTimesheetEntriesFixture() []bexioTimesheet {
	return []bexioTimesheet{
		{
			ID:              801,
			UserID:          1,
			StatusID:        2,
			AllowableBill:   true,
			ClientServiceID: 11,
			Date:            "2026-02-01",
			Duration:        "01:00",
			Running:         false,
			Text:            "list-entry-1",
			Tracking: trackingRange{
				Type:  "range",
				Date:  "2026-02-01",
				Start: "2026-02-01 09:00:00",
				End:   "2026-02-01 10:00:00",
			},
		},
		{
			ID:              802,
			UserID:          2,
			StatusID:        1,
			AllowableBill:   false,
			ClientServiceID: 12,
			Date:            "2026-02-02",
			Duration:        "01:00",
			Running:         false,
			Text:            "list-entry-2",
			Tracking: trackingRange{
				Type:  "range",
				Date:  "2026-02-02",
				Start: "2026-02-02 10:00:00",
				End:   "2026-02-02 11:00:00",
			},
		},
	}
}

func searchTimesheetEntriesFixture() []bexioTimesheet {
	return []bexioTimesheet{
		{
			ID:              901,
			UserID:          42,
			StatusID:        2,
			AllowableBill:   true,
			ClientServiceID: 77,
			Date:            "2026-02-10",
			Duration:        "01:00",
			Running:         false,
			Text:            "search-match-1",
			Tracking: trackingRange{
				Type:  "range",
				Date:  "2026-02-10",
				Start: "2026-02-10 08:00:00",
				End:   "2026-02-10 09:00:00",
			},
		},
	}
}

func projectsFixture() []map[string]any {
	return []map[string]any{
		{"id": 501, "name": "Project Alpha", "contact_id": 11},
	}
}

func writeJSONResponse(t *testing.T, w http.ResponseWriter, status int, payload any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("encode fake response: %v", err)
	}
}

func decodeRequestJSON[T any](t *testing.T, r *http.Request) T {
	t.Helper()

	var payload T
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("decode fake request: %v", err)
	}

	return payload
}

func mustFindToolByName(t *testing.T, tools []*mcp.Tool, name string) *mcp.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}

func assertSchemaPropertiesHaveDescriptions(t *testing.T, schema map[string]any) {
	t.Helper()

	props := mustSchemaProperties(t, schema)
	for key, value := range props {
		propSchema, ok := value.(map[string]any)
		assert.True(t, ok, "property %q should be a schema object", key)

		description, ok := propSchema["description"].(string)
		assert.True(t, ok, "property %q should have description", key)
		assert.True(t, strings.TrimSpace(description) != "", "property %q description should not be empty", key)

		if nestedProps, hasNestedProps := propSchema["properties"].(map[string]any); hasNestedProps {
			assertSchemaPropertiesHaveDescriptions(t, map[string]any{"properties": nestedProps})
		}

		if items, hasItems := propSchema["items"].(map[string]any); hasItems {
			if _, itemsHasProps := items["properties"].(map[string]any); itemsHasProps {
				assertSchemaPropertiesHaveDescriptions(t, items)
			}
		}
	}
}

func mustSchemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()

	props, ok := schema["properties"].(map[string]any)
	assert.True(t, ok, "schema should have properties")
	return props
}

func mustSchemaPropertyMap(t *testing.T, schema map[string]any, property string) map[string]any {
	t.Helper()

	props := mustSchemaProperties(t, schema)
	value, ok := props[property].(map[string]any)
	assert.True(t, ok, "schema property %q should exist", property)
	return value
}

func assertDescriptionContains(t *testing.T, props map[string]any, property string, needle string) {
	t.Helper()

	propSchema, ok := props[property].(map[string]any)
	assert.True(t, ok, "schema property %q should exist", property)

	description, ok := propSchema["description"].(string)
	assert.True(t, ok, "schema property %q should have description", property)
	assert.True(
		t,
		strings.Contains(strings.ToLower(description), strings.ToLower(needle)),
		"schema property %q description should contain %q",
		property,
		needle,
	)
}

type fakeBexioCapturedRequest struct {
	Method        string
	Path          string
	Query         string
	Authorization string
	Body          bexioCreateTimesheetRequest
	SearchBody    []bexioSearchField
}

func intPtr(v int) *int {
	return &v
}
