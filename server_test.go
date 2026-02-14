package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
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

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodPost,
		Path:          "/2.0/timesheet",
		Authorization: "Bearer test-token",
		Body: bexioCreateTimesheetRequest{
			UserID:          42,
			StatusID:        2,
			AllowableBill:   true,
			ClientServiceID: 99,
			ContactID:       intPtr(11),
			PrProjectID:     intPtr(12),
			Text:            "Build acceptance test",
			Tracking: trackingRange{
				Type:  "range",
				Date:  "2026-02-08",
				Start: "09:00",
				End:   "10:30",
			},
		},
	}, env.fakeAPI.Received())

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

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodPost,
		Path:          "/2.0/timesheet",
		Authorization: "Bearer test-token",
		Body: bexioCreateTimesheetRequest{
			UserID:          4,
			StatusID:        2,
			AllowableBill:   true,
			ClientServiceID: 99,
			Text:            "Build acceptance test with defaults",
			Tracking: trackingRange{
				Type:  "range",
				Date:  "2026-02-08",
				Start: "09:00",
				End:   "10:30",
			},
		},
	}, env.fakeAPI.Received())

	assert.False(t, result.IsError)
	assert.True(t, len(result.Content) > 0, "result should have content")
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

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodDelete,
		Path:          "/2.0/timesheet/777",
		Authorization: "Bearer test-token",
	}, env.fakeAPI.Received())

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

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodPost,
		Path:          "/2.0/timesheet/777",
		Authorization: "Bearer test-token",
		Body: bexioCreateTimesheetRequest{
			UserID:          42,
			AllowableBill:   true,
			ClientServiceID: 99,
			ContactID:       intPtr(11),
			PrProjectID:     intPtr(12),
			Text:            "Updated acceptance test",
			Tracking: trackingRange{
				Type:  "range",
				Date:  "2026-02-08",
				Start: "13:00",
				End:   "14:30",
			},
		},
	}, env.fakeAPI.Received())

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

func TestListTimesheetsAcceptance(t *testing.T) {
	// Slice: List & Search Timesheets
	// Given a running MCP server, when the client calls list_timesheets, then the tool returns recent timesheet entries with realistic tracking datetime format and core API fields.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "list_timesheets",
	})
	assert.NoError(t, err)

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/2.0/timesheet",
		Authorization: "Bearer test-token",
	}, env.fakeAPI.Received())

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

	searchFields := []bexioSearchField{
		{Field: "user_id", Value: "42", Criteria: "="},
	}

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "search_timesheets",
		Arguments: map[string]any{
			"search_fields": []map[string]any{
				{"field": "user_id", "value": "42", "criteria": "="},
			},
		},
	})
	assert.NoError(t, err)

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodPost,
		Path:          "/2.0/timesheet/search",
		Authorization: "Bearer test-token",
		SearchBody:    searchFields,
	}, env.fakeAPI.Received())

	contentJSON, err := json.Marshal(result.Content)
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(
		t,
		strings.Contains(string(contentJSON), "search-match-1"),
		"result should contain matching timesheet entries",
	)
}

func TestListContactsAcceptance(t *testing.T) {
	// Slice: Lookup Tools (Contacts, Projects, Packages, Services)
	// Given a running MCP server, when the client calls list_contacts, then the tool returns a list of contacts with id and name.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "list_contacts",
	})
	assert.NoError(t, err)

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/2.0/contact",
		Authorization: "Bearer test-token",
	}, env.fakeAPI.Received())

	contentJSON, err := json.Marshal(result.Content)
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(
		t,
		strings.Contains(string(contentJSON), "Acme Corp"),
		"result should contain contact entries",
	)
}

func TestListProjectsAcceptance(t *testing.T) {
	// Slice: Lookup Tools (Contacts, Projects, Packages, Services)
	// Given a running MCP server, when the client calls list_projects with an optional contact_id filter, then the tool returns projects associated with that contact.
	env := newAcceptanceEnv(t)

	searchFields := []bexioSearchField{
		{Field: "contact_id", Value: "11", Criteria: "="},
	}

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "list_projects",
		Arguments: map[string]any{
			"contact_id": 11,
		},
	})
	assert.NoError(t, err)

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodPost,
		Path:          "/2.0/pr_project/search",
		Authorization: "Bearer test-token",
		SearchBody:    searchFields,
	}, env.fakeAPI.Received())

	contentJSON, err := json.Marshal(result.Content)
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(
		t,
		strings.Contains(string(contentJSON), "Project Alpha"),
		"result should contain project entries",
	)
}

func TestListProjectsWithoutContactIDAcceptance(t *testing.T) {
	// Slice: Fix Smoke Test Findings
	// Given list_projects is called without contact_id, when the request reaches Bexio, then it uses GET /2.0/pr_project instead of POST search.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name:      "list_projects",
		Arguments: map[string]any{},
	})
	assert.NoError(t, err)

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/2.0/pr_project",
		Authorization: "Bearer test-token",
	}, env.fakeAPI.Received())

	contentJSON, err := json.Marshal(result.Content)
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(
		t,
		strings.Contains(string(contentJSON), "Project Alpha"),
		"result should contain project entries",
	)
}

func TestListClientServicesAcceptance(t *testing.T) {
	// Slice: Lookup Tools (Contacts, Projects, Packages, Services)
	// Given a running MCP server, when the client calls list_client_services, then the tool returns available business activities (Taetigkeiten).
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "list_client_services",
	})
	assert.NoError(t, err)

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/2.0/client_service",
		Authorization: "Bearer test-token",
	}, env.fakeAPI.Received())

	contentJSON, err := json.Marshal(result.Content)
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(
		t,
		strings.Contains(string(contentJSON), "Engineering"),
		"result should contain client service entries",
	)
}

func TestListPackagesAcceptance(t *testing.T) {
	// Slice: Lookup Tools (Contacts, Projects, Packages, Services)
	// Given a running MCP server, when the client calls list_packages with a project_id, then the tool returns work packages for that project.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "list_packages",
		Arguments: map[string]any{
			"project_id": 5,
		},
	})
	assert.NoError(t, err)

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/3.0/projects/5/packages",
		Authorization: "Bearer test-token",
	}, env.fakeAPI.Received())

	contentJSON, err := json.Marshal(result.Content)
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(
		t,
		strings.Contains(string(contentJSON), "Backend Sprint"),
		"result should contain package entries",
	)
}

func TestListTimesheetStatusesAcceptance(t *testing.T) {
	// Slice: Timesheet Status & Current User
	// Given a running MCP server, when the client calls list_timesheet_statuses, then the tool returns available statuses with id and name.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "list_timesheet_statuses",
	})
	assert.NoError(t, err)

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/2.0/timesheet_status",
		Authorization: "Bearer test-token",
	}, env.fakeAPI.Received())

	contentJSON, err := json.Marshal(result.Content)
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(
		t,
		strings.Contains(string(contentJSON), "Erledigt"),
		"result should contain timesheet status entries",
	)
}

func TestGetCurrentUserAcceptance(t *testing.T) {
	// Slice: Timesheet Status & Current User
	// Given a running MCP server, when the client calls get_current_user, then the tool returns the authenticated user's id and name.
	env := newAcceptanceEnv(t)

	result, err := env.callTool(&mcp.CallToolParams{
		Name: "get_current_user",
	})
	assert.NoError(t, err)

	assert.Equal(t, fakeBexioCapturedRequest{
		Method:        http.MethodGet,
		Path:          "/3.0/users/me",
		Authorization: "Bearer test-token",
	}, env.fakeAPI.Received())

	contentJSON, err := json.Marshal(result.Content)
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(
		t,
		strings.Contains(string(contentJSON), "Rudolph"),
		"result should contain current user entry",
	)
}

type fakeBexioAPI struct {
	URL    string
	server *httptest.Server

	mu       sync.Mutex
	captured fakeBexioCapturedRequest
}

type acceptanceEnv struct {
	ctx      context.Context
	fakeAPI  *fakeBexioAPI
	bexio    BexioClient
	callTool func(params *mcp.CallToolParams) (*mcp.CallToolResult, error)
}

type contractTestEnv struct {
	name    string
	env     acceptanceEnv
	cleanup func(t *testing.T, timesheetID int)
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

	server := newMCPServer(bexio)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	assert.NoError(t, err)
	t.Cleanup(func() { serverSession.Close() })

	testClient := mcp.NewClient(&mcp.Implementation{Name: "acceptance-test", Version: "0.1.0"}, nil)
	clientSession, err := testClient.Connect(ctx, clientTransport, nil)
	assert.NoError(t, err)
	t.Cleanup(func() { clientSession.Close() })

	return acceptanceEnv{
		ctx:   ctx,
		bexio: bexio,
		callTool: func(params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
			return clientSession.CallTool(ctx, params)
		},
	}
}

func contractTestEnvs(t *testing.T) []contractTestEnv {
	t.Helper()

	envs := []contractTestEnv{
		{
			name: "fake",
			env:  newAcceptanceEnv(t),
			cleanup: func(t *testing.T, _ int) {
				t.Helper()
			},
		},
	}

	token := os.Getenv("BEXIO_API_TOKEN")
	if token != "" {
		realBexio := NewBexioClient("https://api.bexio.com", token, http.DefaultClient)
		realEnv := newAcceptanceEnvWith(t, realBexio)
		envs = append(envs, contractTestEnv{
			name: "real",
			env:  realEnv,
			cleanup: func(t *testing.T, timesheetID int) {
				t.Helper()
				safeDeleteTestTimesheet(realEnv.ctx, t, realBexio, timesheetID)
			},
		})
	}

	return envs
}

func discoverClientServiceID(t *testing.T, env acceptanceEnv) int {
	t.Helper()

	type service struct {
		ID int `json:"id"`
	}
	services := callToolAndDecode[[]service](t, env, &mcp.CallToolParams{Name: "list_client_services"})
	assert.True(t, len(services) > 0, "need at least one client service")

	return services[0].ID
}

func callToolAndDecode[T any](t *testing.T, env acceptanceEnv, params *mcp.CallToolParams) T {
	t.Helper()

	result, err := env.callTool(params)
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, 1, len(result.Content))

	textContent, ok := result.Content[0].(*mcp.TextContent)
	assert.True(t, ok, "result should contain one text content item")

	var decoded T
	err = json.Unmarshal([]byte(textContent.Text), &decoded)
	assert.NoError(t, err)

	return decoded
}

func startFakeBexioAPI(t *testing.T) *fakeBexioAPI {
	t.Helper()

	api := &fakeBexioAPI{}
	exactHandlers := map[string]http.HandlerFunc{
		http.MethodPost + " /2.0/timesheet": func(w http.ResponseWriter, r *http.Request) {
			reqBody := decodeRequestJSON[bexioCreateTimesheetRequest](t, r)
			api.captureWithTimesheet(r, reqBody)
			writeJSONResponse(t, w, http.StatusCreated, buildTimesheet(777, reqBody))
		},
		http.MethodGet + " /2.0/timesheet": func(w http.ResponseWriter, r *http.Request) {
			api.capture(r)
			writeJSONResponse(t, w, http.StatusOK, listTimesheetEntriesFixture())
		},
		http.MethodPost + " /2.0/timesheet/search": func(w http.ResponseWriter, r *http.Request) {
			reqBody := decodeRequestJSON[[]bexioSearchField](t, r)
			api.captureWithSearch(r, reqBody)
			writeJSONResponse(t, w, http.StatusOK, searchTimesheetEntriesFixture())
		},
		http.MethodGet + " /2.0/contact": func(w http.ResponseWriter, r *http.Request) {
			api.capture(r)
			contacts := []map[string]any{
				{"id": 11, "name_1": "Acme Corp", "name_2": ""},
				{"id": 12, "name_1": "Globex Inc", "name_2": ""},
			}
			writeJSONResponse(t, w, http.StatusOK, contacts)
		},
		http.MethodPost + " /2.0/pr_project/search": func(w http.ResponseWriter, r *http.Request) {
			reqBody := decodeRequestJSON[[]bexioSearchField](t, r)
			api.captureWithSearch(r, reqBody)
			writeJSONResponse(t, w, http.StatusOK, projectsFixture())
		},
		http.MethodGet + " /2.0/pr_project": func(w http.ResponseWriter, r *http.Request) {
			api.capture(r)
			writeJSONResponse(t, w, http.StatusOK, projectsFixture())
		},
		http.MethodGet + " /2.0/client_service": func(w http.ResponseWriter, r *http.Request) {
			api.capture(r)
			services := []map[string]any{
				{"id": 77, "name": "Engineering"},
				{"id": 78, "name": "Consulting"},
			}
			writeJSONResponse(t, w, http.StatusOK, services)
		},
		http.MethodGet + " /3.0/projects/5/packages": func(w http.ResponseWriter, r *http.Request) {
			api.capture(r)
			packages := []map[string]any{
				{"id": 61, "name": "Backend Sprint"},
				{"id": 62, "name": "QA Run"},
			}
			writeJSONResponse(t, w, http.StatusOK, packages)
		},
		http.MethodGet + " /2.0/timesheet_status": func(w http.ResponseWriter, r *http.Request) {
			api.capture(r)
			statuses := []map[string]any{
				{"id": 1, "name": "Offen"},
				{"id": 2, "name": "Erledigt"},
			}
			writeJSONResponse(t, w, http.StatusOK, statuses)
		},
		http.MethodGet + " /3.0/users/me": func(w http.ResponseWriter, r *http.Request) {
			api.capture(r)
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

		if handled := handleTimesheetByIDRoutes(t, api, w, r); handled {
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

func (f *fakeBexioAPI) capture(r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.captured = fakeBexioCapturedRequest{
		Method:        r.Method,
		Path:          r.URL.Path,
		Authorization: r.Header.Get("Authorization"),
	}
}

func (f *fakeBexioAPI) captureWithTimesheet(r *http.Request, body bexioCreateTimesheetRequest) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.captured = fakeBexioCapturedRequest{
		Method:        r.Method,
		Path:          r.URL.Path,
		Authorization: r.Header.Get("Authorization"),
		Body:          body,
	}
}

func (f *fakeBexioAPI) captureWithSearch(r *http.Request, searchBody []bexioSearchField) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.captured = fakeBexioCapturedRequest{
		Method:        r.Method,
		Path:          r.URL.Path,
		Authorization: r.Header.Get("Authorization"),
		SearchBody:    searchBody,
	}
}

func handleTimesheetByIDRoutes(t *testing.T, api *fakeBexioAPI, w http.ResponseWriter, r *http.Request) bool {
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
		api.captureWithTimesheet(r, reqBody)
		writeJSONResponse(t, w, http.StatusOK, buildTimesheet(id, reqBody))
		return true
	case http.MethodGet:
		api.capture(r)
		writeJSONResponse(t, w, http.StatusOK, bexioTimesheet{
			ID:   id,
			Text: "[MCP-TEST] contract test create",
		})
		return true
	case http.MethodDelete:
		api.capture(r)
		writeJSONResponse(t, w, http.StatusOK, map[string]bool{"success": true})
		return true
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return true
	}
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

func (f *fakeBexioAPI) Received() fakeBexioCapturedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.captured
}

type fakeBexioCapturedRequest struct {
	Method        string
	Path          string
	Authorization string
	Body          bexioCreateTimesheetRequest
	SearchBody    []bexioSearchField
}

func intPtr(v int) *int {
	return &v
}

func TestContractCreateTimesheetAcceptance(t *testing.T) {
	// Slice: Real API Contract Tests (Environment-Polymorphic)
	// Given create_timesheet runs in fake and real environments, when the tool is called with valid payloads, then both modes satisfy the same observable contract.
	for _, tc := range contractTestEnvs(t) {
		t.Run(tc.name, func(t *testing.T) {
			serviceID := discoverClientServiceID(t, tc.env)

			created := callToolAndDecode[bexioTimesheet](t, tc.env, &mcp.CallToolParams{
				Name: "create_timesheet",
				Arguments: map[string]any{
					"allowable_bill":    true,
					"client_service_id": serviceID,
					"text":              "[MCP-TEST] contract test create",
					"tracking": map[string]any{
						"type":  "range",
						"date":  "2026-02-08",
						"start": "09:00",
						"end":   "10:30",
					},
				},
			})
			assert.NotEqual(t, 0, created.ID)
			assert.Equal(t, "[MCP-TEST] contract test create", created.Text)

			t.Cleanup(func() { tc.cleanup(t, created.ID) })
		})
	}
}

func TestContractListTimesheetsAcceptance(t *testing.T) {
	// Slice: Real API Contract Tests (Environment-Polymorphic)
	// Given list_timesheets runs in fake and real environments, when the tool is called, then both modes return a list of timesheets with valid shape.
	for _, tc := range contractTestEnvs(t) {
		t.Run(tc.name, func(t *testing.T) {
			timesheets := callToolAndDecode[[]bexioTimesheet](t, tc.env, &mcp.CallToolParams{
				Name: "list_timesheets",
			})

			assert.True(t, len(timesheets) > 0, "should return at least one timesheet")

			first := timesheets[0]
			assert.NotEqual(t, 0, first.ID)
			assert.NotEqual(t, 0, first.UserID)
			assert.NotEqual(t, 0, first.StatusID)
			assert.NotEqual(t, 0, first.ClientServiceID)
			assert.NotEqual(t, "", first.Date)
			assert.NotEqual(t, "", first.Duration)
			assert.NotEqual(t, "", first.Tracking.Type)
		})
	}
}

func TestContractListClientServicesAcceptance(t *testing.T) {
	// Slice: Real API Contract Tests (Environment-Polymorphic)
	// Given list_client_services runs in fake and real environments, when the tool is called, then both modes return a list of client services with valid shape.
	for _, tc := range contractTestEnvs(t) {
		t.Run(tc.name, func(t *testing.T) {
			services := callToolAndDecode[[]bexioClientService](t, tc.env, &mcp.CallToolParams{
				Name: "list_client_services",
			})

			assert.True(t, len(services) > 0, "should return at least one client service")

			first := services[0]
			assert.NotEqual(t, 0, first.ID)
			assert.NotEqual(t, "", first.Name)
		})
	}
}

func safeDeleteTestTimesheet(ctx context.Context, t *testing.T, bexio BexioClient, id int) {
	t.Helper()

	timesheet, err := bexio.GetTimesheet(ctx, id)
	assert.NoError(t, err)

	if !strings.HasPrefix(timesheet.Text, "[MCP-TEST]") {
		t.Fatalf("refusing to delete non-test timesheet %d", id)
	}

	_, err = bexio.DeleteTimesheet(ctx, id)
	assert.NoError(t, err)
}
