package main

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestFilterTimesheetsByDate(t *testing.T) {
	t.Parallel()

	t.Run("returns only entries within inclusive date range", func(t *testing.T) {
		t.Parallel()

		timesheets := []bexioTimesheet{
			{Date: "2026-02-01"},
			{Date: "2026-02-02"},
		}

		filtered := filterTimesheetsByDate(timesheets, "2026-02-02", "2026-02-02")

		assert.Equal(t, 1, len(filtered))
		assert.Equal(t, "2026-02-02", filtered[0].Date)
	})

	t.Run("returns entries from date onward when to is empty", func(t *testing.T) {
		t.Parallel()

		timesheets := []bexioTimesheet{
			{Date: "2026-02-01"},
			{Date: "2026-02-02"},
		}

		filtered := filterTimesheetsByDate(timesheets, "2026-02-02", "")

		assert.Equal(t, 1, len(filtered))
		assert.Equal(t, "2026-02-02", filtered[0].Date)
	})
}
