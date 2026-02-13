package main

type timesheetStatus struct {
	ID   int
	Name string
}

func resolveTimesheetDefaults(input createTimesheetInput, currentUserID int, statuses []timesheetStatus) bexioCreateTimesheetRequest {
	request := bexioCreateTimesheetRequest{
		AllowableBill:   input.AllowableBill,
		ClientServiceID: input.ClientServiceID,
		Tracking:        input.Tracking,
		Text:            input.Text,
		ContactID:       input.ContactID,
		PrProjectID:     input.PrProjectID,
	}

	if input.UserID != nil {
		request.UserID = *input.UserID
	} else {
		request.UserID = currentUserID
	}

	if input.StatusID != nil {
		request.StatusID = *input.StatusID
		return request
	}

	for _, status := range statuses {
		if status.Name == "Erledigt" {
			request.StatusID = status.ID
			return request
		}
	}

	return request
}
