package main

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
		assert.Equal(t, 18, len(enum))
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
		assert.Equal(t, 6, len(enum))
	})
}
