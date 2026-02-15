package main

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestFilterTimesheetsByDate(t *testing.T) {
	t.Parallel()

	timesheets := []bexioTimesheet{
		{Date: "2026-02-01"},
		{Date: "2026-02-02"},
		{Date: "2026-02-03"},
	}

	testCases := []struct {
		name          string
		from          string
		to            string
		expectedDates []string
	}{
		{
			name:          "returns only entries within inclusive date range",
			from:          "2026-02-02",
			to:            "2026-02-02",
			expectedDates: []string{"2026-02-02"},
		},
		{
			name:          "returns entries from date onward when to is empty",
			from:          "2026-02-02",
			to:            "",
			expectedDates: []string{"2026-02-02", "2026-02-03"},
		},
		{
			name:          "returns entries up to date when from is empty",
			from:          "",
			to:            "2026-02-02",
			expectedDates: []string{"2026-02-01", "2026-02-02"},
		},
		{
			name:          "returns all entries when both bounds are empty",
			from:          "",
			to:            "",
			expectedDates: []string{"2026-02-01", "2026-02-02", "2026-02-03"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filtered := filterTimesheetsByDateRange(timesheets, tc.from, tc.to)

			actualDates := make([]string, 0, len(filtered))
			for _, filteredTimesheet := range filtered {
				actualDates = append(actualDates, filteredTimesheet.Date)
			}

			assert.Equal(t, tc.expectedDates, actualDates)
		})
	}
}
