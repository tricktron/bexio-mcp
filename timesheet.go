package main

type trackingRange struct {
	Type  string `json:"type"`
	Date  string `json:"date,omitempty"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type createTimesheetInput struct {
	UserID          *int          `json:"user_id,omitempty"`
	StatusID        *int          `json:"status_id,omitempty"`
	AllowableBill   bool          `json:"allowable_bill"`
	ClientServiceID int           `json:"client_service_id"`
	Tracking        trackingRange `json:"tracking"`
	Text            string        `json:"text,omitempty"`
	ContactID       *int          `json:"contact_id,omitempty"`
	PrProjectID     *int          `json:"pr_project_id,omitempty"`
}

type bexioCreateTimesheetRequest struct {
	UserID          int           `json:"user_id"`
	StatusID        int           `json:"status_id,omitempty"`
	AllowableBill   bool          `json:"allowable_bill"`
	ClientServiceID int           `json:"client_service_id"`
	Tracking        trackingRange `json:"tracking"`
	Text            string        `json:"text,omitempty"`
	ContactID       *int          `json:"contact_id,omitempty"`
	PrProjectID     *int          `json:"pr_project_id,omitempty"`
}

type bexioEditTimesheetRequest struct {
	bexioCreateTimesheetRequest

	ID int `json:"id"`
}

type bexioDeleteTimesheetRequest struct {
	ID int `json:"id"`
}

type bexioSearchField struct {
	Field    string `json:"field"`
	Value    string `json:"value"`
	Criteria string `json:"criteria,omitempty"`
}

type timesheetDateRangeFilter struct {
	DateFrom *string `json:"date_from,omitempty"`
	DateTo   *string `json:"date_to,omitempty"`
}

type bexioSearchTimesheetsRequest struct {
	timesheetDateRangeFilter

	SearchFields []bexioSearchField `json:"search_fields,omitempty"`
}

type bexioListProjectsRequest struct {
	ContactID *int `json:"contact_id,omitempty"`
}

type bexioListProjectPackagesRequest struct {
	ProjectID int `json:"project_id"`
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
