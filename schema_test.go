package main

import (
	"reflect"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/google/jsonschema-go/jsonschema"
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
