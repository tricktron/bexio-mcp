package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCreateTimesheetAcceptance(t *testing.T) {
	// Slice: Adopt MCP SDK
	// Given a configured MCP server, when create_timesheet is called over MCP, then the timesheet is created in Bexio without timeout.
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
	session, err := testClient.Connect(ctx, clientTransport, nil)
	assert.NoError(t, err)
	t.Cleanup(func() { session.Close() })

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
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
	}, fakeAPI.Received())

	assert.False(t, result.IsError)
	assert.True(t, len(result.Content) > 0, "result should have content")
}

type fakeBexioAPI struct {
	URL    string
	server *httptest.Server

	mu       sync.Mutex
	captured fakeBexioCapturedRequest
}

func startFakeBexioAPI(t *testing.T) *fakeBexioAPI {
	t.Helper()

	api := &fakeBexioAPI{}
	api.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

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
