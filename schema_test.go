package main

import (
	"reflect"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTrackingRangeSchemaTypeEnum(t *testing.T) {
	t.Parallel()

	schema, err := jsonschema.ForType(reflect.TypeFor[trackingRange](), &jsonschema.ForOptions{
		TypeSchemas: map[reflect.Type]*jsonschema.Schema{
			reflect.TypeFor[TrackingType](): {Type: "string", Enum: []any{"range", "duration"}},
		},
	})
	assert.NoError(t, err)

	typeProperty := mustSchemaProperty(t, schema, "type")
	assert.Equal(t, []any{"range", "duration"}, typeProperty.Enum)
}

func TestSearchFieldSchemaCriteriaEnum(t *testing.T) {
	t.Parallel()

	schema, err := jsonschema.ForType(reflect.TypeFor[bexioSearchField](), &jsonschema.ForOptions{
		TypeSchemas: map[reflect.Type]*jsonschema.Schema{
			reflect.TypeFor[SearchCriteria](): {
				Type: "string",
				Enum: []any{
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
				},
			},
		},
	})
	assert.NoError(t, err)

	criteriaProperty := mustSchemaProperty(t, schema, "criteria")
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
	}, criteriaProperty.Enum)
}

func TestSearchFieldSchemaFieldEnum(t *testing.T) {
	t.Parallel()

	schema, err := jsonschema.ForType(reflect.TypeFor[bexioSearchField](), &jsonschema.ForOptions{
		TypeSchemas: map[reflect.Type]*jsonschema.Schema{
			reflect.TypeFor[SearchField](): {Type: "string", Enum: []any{
				"id", "client_service_id", "contact_id", "user_id", "pr_project_id", "status_id",
			}},
			reflect.TypeFor[SearchCriteria](): {Type: "string", Enum: []any{
				"=", "!=", ">", ">=", "<", "<=", "like", "not_like", "is_null", "not_null", "in", "not_in",
				"equal", "not_equal", "greater_than", "greater_equal", "less_than", "less_equal",
			}},
		},
	})
	assert.NoError(t, err)

	fieldProperty := mustSchemaProperty(t, schema, "field")
	assert.Equal(t, []any{
		"id", "client_service_id", "contact_id", "user_id", "pr_project_id", "status_id",
	}, fieldProperty.Enum)
}

func mustSchemaProperty(t *testing.T, schema *jsonschema.Schema, name string) *jsonschema.Schema {
	t.Helper()

	property, ok := schema.Properties[name]
	assert.True(t, ok, "schema should include property %q", name)

	return property
}

func TestRegisteredToolSchemasIncludeEnumConstraints(t *testing.T) {
	t.Parallel()

	env := newAcceptanceEnv(t)

	listed, err := env.listTools(&mcp.ListToolsParams{})
	assert.NoError(t, err)

	t.Run("create_timesheet tracking.type enum", func(t *testing.T) {
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
