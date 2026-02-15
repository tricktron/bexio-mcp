package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestBexioSearchTimesheetsRequestSearchFieldsIsOptional(t *testing.T) {
	field, ok := reflect.TypeFor[bexioSearchTimesheetsRequest]().FieldByName("SearchFields")
	if !ok {
		t.Fatal("SearchFields field is missing")
	}

	jsonTag := field.Tag.Get("json")
	if !strings.Contains(jsonTag, "omitempty") {
		t.Fatalf("expected SearchFields json tag to include omitempty, got %q", jsonTag)
	}
}
