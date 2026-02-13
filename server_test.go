package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	}, updated)
}

func TestListTimesheetsAcceptance(t *testing.T) {
	// Slice: List & Search Timesheets
	// Given a running MCP server, when the client calls list_timesheets, then the tool returns a list of recent timesheet entries.
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

	contentJSON, err := json.Marshal(result.Content)
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(
		t,
		strings.Contains(string(contentJSON), "list-entry-1"),
		"result should contain timesheet entries",
	)
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

type fakeBexioAPI struct {
	URL    string
	server *httptest.Server

	mu       sync.Mutex
	captured fakeBexioCapturedRequest
}

type acceptanceEnv struct {
	ctx      context.Context
	fakeAPI  *fakeBexioAPI
	callTool func(params *mcp.CallToolParams) (*mcp.CallToolResult, error)
}

func newAcceptanceEnv(t *testing.T) acceptanceEnv {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)

	fakeAPI := startFakeBexioAPI(t)
	bexio := NewBexioClient(fakeAPI.URL, "test-token", http.DefaultClient)
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
		ctx:     ctx,
		fakeAPI: fakeAPI,
		callTool: func(params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
			return clientSession.CallTool(ctx, params)
		},
	}
}

func startFakeBexioAPI(t *testing.T) *fakeBexioAPI {
	t.Helper()

	api := &fakeBexioAPI{}
	api.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/2.0/timesheet":
			var reqBody bexioCreateTimesheetRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Fatal(err)
			}

			api.mu.Lock()
			api.captured = fakeBexioCapturedRequest{
				Method:        r.Method,
				Path:          r.URL.Path,
				Authorization: r.Header.Get("Authorization"),
				Body:          reqBody,
			}
			api.mu.Unlock()

			created := bexioTimesheet{
				ID:              777,
				UserID:          reqBody.UserID,
				AllowableBill:   reqBody.AllowableBill,
				ClientServiceID: reqBody.ClientServiceID,
				Text:            reqBody.Text,
				ContactID:       reqBody.ContactID,
				PrProjectID:     reqBody.PrProjectID,
				Tracking:        reqBody.Tracking,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(created); err != nil {
				t.Fatal(err)
			}
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/timesheet":
			api.mu.Lock()
			api.captured = fakeBexioCapturedRequest{
				Method:        r.Method,
				Path:          r.URL.Path,
				Authorization: r.Header.Get("Authorization"),
			}
			api.mu.Unlock()

			entries := []bexioTimesheet{
				{
					ID:              801,
					UserID:          1,
					AllowableBill:   true,
					ClientServiceID: 11,
					Text:            "list-entry-1",
					Tracking: trackingRange{
						Type:  "range",
						Date:  "2026-02-01",
						Start: "09:00",
						End:   "10:00",
					},
				},
				{
					ID:              802,
					UserID:          2,
					AllowableBill:   false,
					ClientServiceID: 12,
					Text:            "list-entry-2",
					Tracking: trackingRange{
						Type:  "range",
						Date:  "2026-02-02",
						Start: "10:00",
						End:   "11:00",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(entries); err != nil {
				t.Fatal(err)
			}
		case r.Method == http.MethodPost && r.URL.Path == "/2.0/timesheet/search":
			var reqBody []bexioSearchField
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Fatal(err)
			}

			api.mu.Lock()
			api.captured = fakeBexioCapturedRequest{
				Method:        r.Method,
				Path:          r.URL.Path,
				Authorization: r.Header.Get("Authorization"),
				SearchBody:    reqBody,
			}
			api.mu.Unlock()

			entries := []bexioTimesheet{
				{
					ID:              901,
					UserID:          42,
					AllowableBill:   true,
					ClientServiceID: 77,
					Text:            "search-match-1",
					Tracking: trackingRange{
						Type:  "range",
						Date:  "2026-02-10",
						Start: "08:00",
						End:   "09:00",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(entries); err != nil {
				t.Fatal(err)
			}
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/2.0/timesheet/") && r.URL.Path != "/2.0/timesheet/search":
			idText := strings.TrimPrefix(r.URL.Path, "/2.0/timesheet/")
			id, err := strconv.Atoi(idText)
			if err != nil {
				t.Fatal(err)
			}

			var reqBody bexioCreateTimesheetRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Fatal(err)
			}

			api.mu.Lock()
			api.captured = fakeBexioCapturedRequest{
				Method:        r.Method,
				Path:          r.URL.Path,
				Authorization: r.Header.Get("Authorization"),
				Body:          reqBody,
			}
			api.mu.Unlock()

			updated := bexioTimesheet{
				ID:              id,
				UserID:          reqBody.UserID,
				AllowableBill:   reqBody.AllowableBill,
				ClientServiceID: reqBody.ClientServiceID,
				Text:            reqBody.Text,
				ContactID:       reqBody.ContactID,
				PrProjectID:     reqBody.PrProjectID,
				Tracking:        reqBody.Tracking,
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(updated); err != nil {
				t.Fatal(err)
			}
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/2.0/timesheet/"):
			idText := strings.TrimPrefix(r.URL.Path, "/2.0/timesheet/")
			if _, err := strconv.Atoi(idText); err != nil {
				t.Fatal(err)
			}

			api.mu.Lock()
			api.captured = fakeBexioCapturedRequest{
				Method:        r.Method,
				Path:          r.URL.Path,
				Authorization: r.Header.Get("Authorization"),
			}
			api.mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]bool{"success": true}); err != nil {
				t.Fatal(err)
			}
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/contact":
			api.mu.Lock()
			api.captured = fakeBexioCapturedRequest{
				Method:        r.Method,
				Path:          r.URL.Path,
				Authorization: r.Header.Get("Authorization"),
			}
			api.mu.Unlock()

			contacts := []map[string]any{
				{"id": 11, "name_1": "Acme Corp", "name_2": ""},
				{"id": 12, "name_1": "Globex Inc", "name_2": ""},
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(contacts); err != nil {
				t.Fatal(err)
			}
		case r.Method == http.MethodPost && r.URL.Path == "/2.0/pr_project/search":
			var reqBody []bexioSearchField
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Fatal(err)
			}

			api.mu.Lock()
			api.captured = fakeBexioCapturedRequest{
				Method:        r.Method,
				Path:          r.URL.Path,
				Authorization: r.Header.Get("Authorization"),
				SearchBody:    reqBody,
			}
			api.mu.Unlock()

			projects := []map[string]any{
				{"id": 501, "name": "Project Alpha", "contact_id": 11},
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(projects); err != nil {
				t.Fatal(err)
			}
		case r.Method == http.MethodGet && r.URL.Path == "/2.0/client_service":
			api.mu.Lock()
			api.captured = fakeBexioCapturedRequest{
				Method:        r.Method,
				Path:          r.URL.Path,
				Authorization: r.Header.Get("Authorization"),
			}
			api.mu.Unlock()

			services := []map[string]any{
				{"id": 77, "name": "Engineering"},
				{"id": 78, "name": "Consulting"},
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(services); err != nil {
				t.Fatal(err)
			}
		case r.Method == http.MethodGet && r.URL.Path == "/3.0/projects/5/packages":
			api.mu.Lock()
			api.captured = fakeBexioCapturedRequest{
				Method:        r.Method,
				Path:          r.URL.Path,
				Authorization: r.Header.Get("Authorization"),
			}
			api.mu.Unlock()

			packages := []map[string]any{
				{"id": 61, "name": "Backend Sprint"},
				{"id": 62, "name": "QA Run"},
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(packages); err != nil {
				t.Fatal(err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	api.URL = api.server.URL
	t.Cleanup(api.server.Close)

	return api
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
