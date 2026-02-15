package main

import (
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
)

func trackingTypeEnumValues() []any {
	return []any{"range", "duration"}
}

func searchCriteriaEnumValues() []any {
	return []any{
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
	}
}

func searchFieldEnumValues() []any {
	return []any{
		"id",
		"client_service_id",
		"contact_id",
		"user_id",
		"pr_project_id",
		"status_id",
	}
}

func timesheetSchemaTypeOverrides() map[reflect.Type]*jsonschema.Schema {
	return map[reflect.Type]*jsonschema.Schema{
		reflect.TypeFor[TrackingType](): {
			Type: "string",
			Enum: trackingTypeEnumValues(),
		},
	}
}

func searchSchemaTypeOverrides() map[reflect.Type]*jsonschema.Schema {
	return map[reflect.Type]*jsonschema.Schema{
		reflect.TypeFor[SearchCriteria](): {
			Type: "string",
			Enum: searchCriteriaEnumValues(),
		},
		reflect.TypeFor[SearchField](): {
			Type: "string",
			Enum: searchFieldEnumValues(),
		},
	}
}
