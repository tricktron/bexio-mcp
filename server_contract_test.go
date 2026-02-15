package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type contractTestEnv struct {
	name    string
	env     acceptanceEnv
	cleanup func(t *testing.T, timesheetID int)
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

	servicesResult := callToolAndDecode[listServicesResult](t, env, &mcp.CallToolParams{Name: "list_client_services"})
	services := servicesResult.Services
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

func TestContractCreateTimesheet(t *testing.T) {
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

func TestContractSearchTimesheets(t *testing.T) {
	// Slice: Merge list_timesheets and search_timesheets
	// Given search_timesheets runs in fake and real environments, when the tool is called, then both modes return a list of timesheets with valid shape.
	for _, tc := range contractTestEnvs(t) {
		t.Run(tc.name, func(t *testing.T) {
			result := callToolAndDecode[searchTimesheetsResult](t, tc.env, &mcp.CallToolParams{
				Name: "search_timesheets",
			})

			timesheets := result.Results

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

func TestContractListClientServices(t *testing.T) {
	// Slice: Real API Contract Tests (Environment-Polymorphic)
	// Given list_client_services runs in fake and real environments, when the tool is called, then both modes return a list of client services with valid shape.
	for _, tc := range contractTestEnvs(t) {
		t.Run(tc.name, func(t *testing.T) {
			servicesResult := callToolAndDecode[listServicesResult](t, tc.env, &mcp.CallToolParams{
				Name: "list_client_services",
			})
			services := servicesResult.Services

			assert.True(t, len(services) > 0, "should return at least one client service")

			first := services[0]
			assert.NotEqual(t, 0, first.ID)
			assert.NotEqual(t, "", first.Name)
		})
	}
}

func TestContractListContacts(t *testing.T) {
	// Slice: Real API Contract Tests (Environment-Polymorphic)
	// Given list_contacts runs in fake and real environments, when the tool is called, then both modes return a list of contacts with valid shape.
	for _, tc := range contractTestEnvs(t) {
		t.Run(tc.name, func(t *testing.T) {
			contactsResult := callToolAndDecode[listContactsResult](t, tc.env, &mcp.CallToolParams{
				Name: "list_contacts",
			})
			contacts := contactsResult.Contacts

			assert.True(t, len(contacts) > 0, "should return at least one contact")

			first := contacts[0]
			assert.NotEqual(t, 0, first.ID)
			assert.NotEqual(t, "", first.Name1)
		})
	}
}

func safeDeleteTestTimesheet(ctx context.Context, t *testing.T, bexio BexioClient, id int) {
	t.Helper()

	timesheet, err := getTimesheetForCleanup(ctx, bexio, id)
	assert.NoError(t, err)

	if !strings.HasPrefix(timesheet.Text, "[MCP-TEST]") {
		t.Fatalf("refusing to delete non-test timesheet %d", id)
	}

	_, err = bexio.DeleteTimesheet(ctx, id)
	assert.NoError(t, err)
}

func getTimesheetForCleanup(ctx context.Context, bexio BexioClient, id int) (bexioTimesheet, error) {
	httpReq, err := bexio.newRequest(ctx, http.MethodGet, timesheetEndpoint+"/"+strconv.Itoa(id), nil)
	if err != nil {
		return bexioTimesheet{}, err
	}

	var timesheet bexioTimesheet
	if err = bexio.doAndDecode(httpReq, &timesheet); err != nil {
		return bexioTimesheet{}, err
	}

	return timesheet, nil
}
