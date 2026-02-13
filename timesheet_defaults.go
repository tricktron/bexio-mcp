package main

import (
	"encoding/json"
	"fmt"
)

type timesheetStatus struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

const completedStatusName = "Erledigt"

func resolveTimesheetDefaults(
	input createTimesheetInput,
	currentUserID int,
	statuses []timesheetStatus,
) bexioCreateTimesheetRequest {
	return bexioCreateTimesheetRequest{
		UserID:          resolveUserID(input.UserID, currentUserID),
		StatusID:        resolveStatusID(input.StatusID, statuses),
		AllowableBill:   input.AllowableBill,
		ClientServiceID: input.ClientServiceID,
		Tracking:        input.Tracking,
		Text:            input.Text,
		ContactID:       input.ContactID,
		PrProjectID:     input.PrProjectID,
	}
}

func resolveUserID(inputUserID *int, currentUserID int) int {
	if inputUserID != nil {
		return *inputUserID
	}

	return currentUserID
}

func resolveStatusID(inputStatusID *int, statuses []timesheetStatus) int {
	if inputStatusID != nil {
		return *inputStatusID
	}

	for _, status := range statuses {
		if status.Name == completedStatusName {
			return status.ID
		}
	}

	return 0
}

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
