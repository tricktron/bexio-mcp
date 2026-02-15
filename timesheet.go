package main

import (
	"encoding/json"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
)

type TrackingType string
type SearchField string
type SearchCriteria string

const durationTrackingType TrackingType = "duration"

type durationTrackingJSON struct {
	Type     TrackingType `json:"type"`
	Date     string       `json:"date,omitempty"`
	Duration string       `json:"duration,omitempty"`
}

type fullTrackingJSON trackingRange

type trackingRange struct {
	Type     TrackingType `json:"type"               jsonschema:"Tracking mode: range or duration"`
	Date     string       `json:"date,omitempty"     jsonschema:"Date of the entry in YYYY-MM-DD format"`
	Duration string       `json:"duration,omitempty" jsonschema:"Duration in HH:MM format (for duration tracking)"`
	Start    string       `json:"start,omitempty"    jsonschema:"Start time in HH:MM format (for range tracking)"`
	End      string       `json:"end,omitempty"      jsonschema:"End time in HH:MM format (for range tracking)"`
}

func (t trackingRange) MarshalJSON() ([]byte, error) {
	if t.Type == durationTrackingType {
		return json.Marshal(durationTrackingJSON{
			Type:     t.Type,
			Date:     t.Date,
			Duration: t.Duration,
		})
	}

	return json.Marshal(fullTrackingJSON(t))
}

type createTimesheetInput struct {
	UserID          *int          `json:"user_id,omitempty"       jsonschema:"Bexio user ID. If omitted, defaults to the current authenticated user. Use get_current_user to find the ID"`
	StatusID        *int          `json:"status_id,omitempty"     jsonschema:"Timesheet status ID. If omitted, defaults to the 'Erledigt' (completed) status. Use list_timesheet_statuses to find valid IDs"`
	AllowableBill   bool          `json:"allowable_bill"          jsonschema:"Whether this entry is billable"`
	ClientServiceID int           `json:"client_service_id"       jsonschema:"ID of the client service (activity type). Use list_client_services to find the ID by name"`
	Tracking        trackingRange `json:"tracking"                jsonschema:"Tracked time details (type plus date/time fields)"`
	Text            string        `json:"text,omitempty"          jsonschema:"Description text for the timesheet entry"`
	ContactID       *int          `json:"contact_id,omitempty"    jsonschema:"Bexio contact ID to associate with this entry. Use list_contacts to find the ID by name"`
	PrProjectID     *int          `json:"pr_project_id,omitempty" jsonschema:"Bexio project ID to associate with this entry. Use list_projects to find the ID by name"`
}

type bexioCreateTimesheetRequest struct {
	UserID          int           `json:"user_id"                 jsonschema:"Bexio user ID. Use get_current_user to find the ID"`
	StatusID        int           `json:"status_id,omitempty"     jsonschema:"Timesheet status ID. Use list_timesheet_statuses to find valid IDs"`
	AllowableBill   bool          `json:"allowable_bill"          jsonschema:"Whether this entry is billable"`
	ClientServiceID int           `json:"client_service_id"       jsonschema:"ID of the client service (activity type). Use list_client_services to find the ID by name"`
	Tracking        trackingRange `json:"tracking"                jsonschema:"Tracked time details (type plus date/time fields)"`
	Text            string        `json:"text,omitempty"          jsonschema:"Description text for the timesheet entry"`
	ContactID       *int          `json:"contact_id,omitempty"    jsonschema:"Bexio contact ID to associate with this entry. Use list_contacts to find the ID by name"`
	PrProjectID     *int          `json:"pr_project_id,omitempty" jsonschema:"Bexio project ID to associate with this entry. Use list_projects to find the ID by name"`
}

type bexioEditTimesheetRequest struct {
	bexioCreateTimesheetRequest

	ID int `json:"id" jsonschema:"ID of the timesheet entry to edit. Use search_timesheets to find the ID"`
}

type bexioDeleteTimesheetRequest struct {
	ID int `json:"id" jsonschema:"ID of the timesheet entry to delete. Use search_timesheets to find the ID"`
}

type bexioSearchField struct {
	Field    SearchField    `json:"field"              jsonschema:"Bexio timesheet field name to filter on (e.g. 'user_id', 'contact_id', 'pr_project_id')"`
	Value    string         `json:"value"              jsonschema:"Value to match against the field"`
	Criteria SearchCriteria `json:"criteria,omitempty" jsonschema:"Match operator: '=' for exact match, 'like' for partial, '!=' for not equal, '>' for greater, '<' for less"`
}

type timesheetDateRangeFilter struct {
	DateFrom *string `json:"date_from,omitempty" jsonschema:"Start of date range filter in YYYY-MM-DD format (client-side filtering)"`
	DateTo   *string `json:"date_to,omitempty"   jsonschema:"End of date range filter in YYYY-MM-DD format (client-side filtering)"`
}

type bexioSearchTimesheetsRequest struct {
	timesheetDateRangeFilter

	SearchFields []bexioSearchField `json:"search_fields,omitempty" jsonschema:"Optional Bexio field filters. If omitted, all timesheets are listed"`
}

type bexioListProjectsRequest struct {
	ContactID *int `json:"contact_id,omitempty" jsonschema:"Filter projects by contact ID. Use list_contacts to find the ID by name. If omitted, lists all projects"`
}

type bexioListProjectPackagesRequest struct {
	ProjectID int `json:"project_id" jsonschema:"Bexio project ID to list work packages for. Use list_projects to find the ID by name"`
}

type bexioTimesheet struct {
	ID              int           `json:"id"`
	UserID          int           `json:"user_id"`
	StatusID        int           `json:"status_id"`
	AllowableBill   bool          `json:"allowable_bill"`
	ClientServiceID int           `json:"client_service_id"`
	Date            string        `json:"date"`
	Duration        string        `json:"duration"`
	Running         bool          `json:"running"`
	Text            string        `json:"text,omitempty"`
	ContactID       *int          `json:"contact_id,omitempty"`
	PrProjectID     *int          `json:"pr_project_id,omitempty"`
	Tracking        trackingRange `json:"tracking"`
}

func timesheetTypeSchemas() map[reflect.Type]*jsonschema.Schema {
	return map[reflect.Type]*jsonschema.Schema{
		reflect.TypeFor[TrackingType](): {
			Type: "string",
			Enum: []any{"range", "duration"},
		},
	}
}

func searchTypeSchemas() map[reflect.Type]*jsonschema.Schema {
	return map[reflect.Type]*jsonschema.Schema{
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
		reflect.TypeFor[SearchField](): {
			Type: "string",
			Enum: []any{
				"id",
				"client_service_id",
				"contact_id",
				"user_id",
				"pr_project_id",
				"status_id",
			},
		},
	}
}
