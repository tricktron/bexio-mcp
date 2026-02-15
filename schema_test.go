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

func mustSchemaProperty(t *testing.T, schema *jsonschema.Schema, name string) *jsonschema.Schema {
	t.Helper()

	property, ok := schema.Properties[name]
	assert.True(t, ok, "schema should include property %q", name)

	return property
}
