package main

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRegisterToolExposesOutputSchemaForTypedResult(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)

	server := mcp.NewServer(&mcp.Implementation{Name: "schema-unit-test", Version: "0.1.0"}, nil)
	err := registerTool(
		server,
		"typed_delete",
		"Unit test tool for typed output schema",
		func(context.Context, struct{}) (deleteTimesheetResult, error) {
			return deleteTimesheetResult{Success: true}, nil
		},
		nil,
	)
	assert.NoError(t, err)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	assert.NoError(t, err)
	t.Cleanup(func() { serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "schema-unit-client", Version: "0.1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	assert.NoError(t, err)
	t.Cleanup(func() { clientSession.Close() })

	listed, err := clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	assert.NoError(t, err)

	tool := mustFindToolByName(t, listed.Tools, "typed_delete")
	assert.True(t, tool.OutputSchema != nil, "tool should expose output schema for typed result")

	outputSchema, ok := tool.OutputSchema.(map[string]any)
	assert.True(t, ok, "output schema should be a map")

	properties, ok := outputSchema["properties"].(map[string]any)
	assert.True(t, ok, "output schema should include properties")

	_, hasSuccess := properties["success"]
	assert.True(t, hasSuccess, "output schema should include success property")
}

func TestRegisteredToolSchemasIncludeEnumConstraints(t *testing.T) {
	t.Parallel()

	env := newAcceptanceEnv(t)

	listed, err := env.listTools(&mcp.ListToolsParams{})
	assert.NoError(t, err)

	t.Run("create_timesheet tracking.type enum", func(t *testing.T) {
		t.Parallel()

		createTool := mustFindToolByName(t, listed.Tools, "create_timesheet")
		createSchema, ok := createTool.InputSchema.(map[string]any)
		assert.True(t, ok)

		trackingSchema := mustSchemaPropertyMap(t, createSchema, "tracking")
		trackingProps := mustSchemaProperties(t, trackingSchema)
		typeSchema, ok := trackingProps["type"].(map[string]any)
		assert.True(t, ok, "tracking.type should be a schema object")

		enum, ok := typeSchema["enum"].([]any)
		assert.True(t, ok, "tracking.type should have enum")
		assert.Equal(t, []any{"range", "duration"}, enum)
	})

	t.Run("search_timesheets criteria enum", func(t *testing.T) {
		t.Parallel()

		searchTool := mustFindToolByName(t, listed.Tools, "search_timesheets")
		searchSchema, ok := searchTool.InputSchema.(map[string]any)
		assert.True(t, ok)

		searchFieldsSchema := mustSchemaPropertyMap(t, searchSchema, "search_fields")
		items, ok := searchFieldsSchema["items"].(map[string]any)
		assert.True(t, ok, "search_fields should have items")

		itemProps := mustSchemaProperties(t, items)
		criteriaSchema, ok := itemProps["criteria"].(map[string]any)
		assert.True(t, ok, "criteria should be a schema object")

		enum, ok := criteriaSchema["enum"].([]any)
		assert.True(t, ok, "criteria should have enum")
		assert.Equal(t, []any{
			"=",
			"!=",
			">",
			">=",
			"<",
			"<=",
			"like",
			"not_like",
			"is_null",
			"not_null",
			"in",
			"not_in",
			"equal",
			"not_equal",
			"greater_than",
			"greater_equal",
			"less_than",
			"less_equal",
		}, enum)
	})

	t.Run("search_timesheets field enum", func(t *testing.T) {
		t.Parallel()

		searchTool := mustFindToolByName(t, listed.Tools, "search_timesheets")
		searchSchema, ok := searchTool.InputSchema.(map[string]any)
		assert.True(t, ok)

		searchFieldsSchema := mustSchemaPropertyMap(t, searchSchema, "search_fields")
		items, ok := searchFieldsSchema["items"].(map[string]any)
		assert.True(t, ok)

		itemProps := mustSchemaProperties(t, items)
		fieldSchema, ok := itemProps["field"].(map[string]any)
		assert.True(t, ok, "field should be a schema object")

		enum, ok := fieldSchema["enum"].([]any)
		assert.True(t, ok, "field should have enum")
		assert.Equal(t, []any{"id", "client_service_id", "contact_id", "user_id", "pr_project_id", "status_id"}, enum)
	})

	t.Run("search_timesheets date_from pattern", func(t *testing.T) {
		t.Parallel()

		searchTool := mustFindToolByName(t, listed.Tools, "search_timesheets")
		searchSchema, ok := searchTool.InputSchema.(map[string]any)
		assert.True(t, ok)

		dateFromSchema := mustSchemaPropertyMap(t, searchSchema, "date_from")
		pattern, ok := dateFromSchema["pattern"].(string)
		assert.True(t, ok, "date_from should have pattern")
		assert.Equal(t, `^\d{4}-\d{2}-\d{2}$`, pattern)
	})

	t.Run("search_timesheets date_to pattern", func(t *testing.T) {
		t.Parallel()

		searchTool := mustFindToolByName(t, listed.Tools, "search_timesheets")
		searchSchema, ok := searchTool.InputSchema.(map[string]any)
		assert.True(t, ok)

		dateToSchema := mustSchemaPropertyMap(t, searchSchema, "date_to")
		pattern, ok := dateToSchema["pattern"].(string)
		assert.True(t, ok, "date_to should have pattern")
		assert.Equal(t, `^\d{4}-\d{2}-\d{2}$`, pattern)
	})

	t.Run("create_timesheet tracking.date pattern", func(t *testing.T) {
		t.Parallel()

		createTool := mustFindToolByName(t, listed.Tools, "create_timesheet")
		createSchema, ok := createTool.InputSchema.(map[string]any)
		assert.True(t, ok)

		trackingSchema := mustSchemaPropertyMap(t, createSchema, "tracking")
		trackingProps := mustSchemaProperties(t, trackingSchema)
		dateSchema, ok := trackingProps["date"].(map[string]any)
		assert.True(t, ok, "tracking.date should be a schema object")

		pattern, ok := dateSchema["pattern"].(string)
		assert.True(t, ok, "tracking.date should have pattern")
		assert.Equal(t, `^\d{4}-\d{2}-\d{2}$`, pattern)
	})
}

func TestTimesheetToolsExposeOutputSchemasInListAcceptance(t *testing.T) {
	// Slice: Typed output schemas for timesheet tools
	// Given the MCP server is running, when a client calls tools/list, then timesheet tools expose expected output schema properties.
	t.Parallel()

	env := newAcceptanceEnv(t)

	listed, err := env.listTools(&mcp.ListToolsParams{})
	assert.NoError(t, err)

	testCases := []struct {
		name       string
		properties []string
	}{
		{
			name: "create_timesheet",
			properties: []string{
				"id",
				"user_id",
				"status_id",
				"allowable_bill",
				"client_service_id",
				"date",
				"duration",
				"running",
				"tracking",
			},
		},
		{
			name: "edit_timesheet",
			properties: []string{
				"id",
				"user_id",
				"status_id",
				"allowable_bill",
				"client_service_id",
				"date",
				"duration",
				"running",
				"tracking",
			},
		},
		{name: "delete_timesheet", properties: []string{"success"}},
		{name: "search_timesheets", properties: []string{"results"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tool := mustFindToolByName(t, listed.Tools, tc.name)
			assert.True(t, tool.OutputSchema != nil, "tool %q should expose output schema", tc.name)

			outputSchema, ok := tool.OutputSchema.(map[string]any)
			assert.True(t, ok, "tool %q output schema should be an object map", tc.name)
			assert.Equal(t, "object", outputSchema["type"])

			properties, ok := outputSchema["properties"].(map[string]any)
			assert.True(t, ok, "tool %q output schema should include properties", tc.name)

			for _, property := range tc.properties {
				_, hasProperty := properties[property]
				assert.True(t, hasProperty, "tool %q output schema should include property %q", tc.name, property)
			}
		})
	}
}
