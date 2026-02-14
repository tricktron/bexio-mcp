package main

func filterTimesheetsByDate(timesheets []bexioTimesheet, from, to string) []bexioTimesheet {
	filtered := make([]bexioTimesheet, 0, len(timesheets))

	for _, timesheet := range timesheets {
		if timesheet.Date >= from && timesheet.Date <= to {
			filtered = append(filtered, timesheet)
		}
	}

	return filtered
}
