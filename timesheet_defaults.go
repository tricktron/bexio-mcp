package main

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
	userID := currentUserID
	if input.UserID != nil {
		userID = *input.UserID
	}

	return bexioCreateTimesheetRequest{
		UserID:          userID,
		StatusID:        resolveStatusID(input.StatusID, statuses),
		AllowableBill:   input.AllowableBill,
		ClientServiceID: input.ClientServiceID,
		Tracking:        input.Tracking,
		Text:            input.Text,
		ContactID:       input.ContactID,
		SubContactID:    input.SubContactID,
		PrProjectID:     input.PrProjectID,
	}
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
