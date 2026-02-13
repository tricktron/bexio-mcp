package main

import (
	"encoding/json"
	"fmt"
)

func parseCurrentUserID(raw json.RawMessage) (int, error) {
	var user struct {
		ID int `json:"id"`
	}

	if err := json.Unmarshal(raw, &user); err != nil {
		return 0, fmt.Errorf("parse current user payload: %w", err)
	}

	return user.ID, nil
}

func parseTimesheetStatuses(raw json.RawMessage) ([]timesheetStatus, error) {
	var statuses []timesheetStatus

	if err := json.Unmarshal(raw, &statuses); err != nil {
		return nil, fmt.Errorf("parse timesheet statuses payload: %w", err)
	}

	return statuses, nil
}
