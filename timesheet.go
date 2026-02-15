package main

type TrackingType string
type SearchField string
type SearchCriteria string

type trackingRange struct {
	Type  TrackingType `json:"type"           jsonschema:"Tracking mode: 'range' for start/end times, or 'duration' for a duration value"`
	Date  string       `json:"date,omitempty" jsonschema:"Date of the entry in YYYY-MM-DD format"`
	Start string       `json:"start"          jsonschema:"Start time in HH:MM format (for range) or duration value"`
	End   string       `json:"end"            jsonschema:"End time in HH:MM format (for range tracking)"`
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
