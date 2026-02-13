package main

type timesheetStatus struct {
	ID   int
	Name string
}

const completedStatusName = "Erledigt"

func resolveTimesheetDefaults(input createTimesheetInput, currentUserID int, statuses []timesheetStatus) bexioCreateTimesheetRequest {
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
