package main

type trackingRange struct {
	Type  string `json:"type"`
	Date  string `json:"date,omitempty"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type bexioCreateTimesheetRequest struct {
	UserID          int           `json:"user_id"`
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

type bexioSearchTimesheetsRequest struct {
	SearchFields []bexioSearchField `json:"search_fields"`
}

type bexioTimesheet struct {
	ID              int           `json:"id"`
	UserID          int           `json:"user_id"`
	AllowableBill   bool          `json:"allowable_bill"`
	ClientServiceID int           `json:"client_service_id"`
	Text            string        `json:"text,omitempty"`
	ContactID       *int          `json:"contact_id,omitempty"`
	PrProjectID     *int          `json:"pr_project_id,omitempty"`
	Tracking        trackingRange `json:"tracking"`
}
