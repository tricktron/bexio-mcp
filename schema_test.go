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
		assert.Equal(t, trackingTypeEnumValues(), enum)
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
		assert.Equal(t, searchCriteriaEnumValues(), enum)
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
		assert.Equal(t, searchFieldEnumValues(), enum)
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
