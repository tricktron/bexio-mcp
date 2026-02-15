package main

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestResolveTimesheetDefaults(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         createTimesheetInput
		currentUserID int
		statuses      []timesheetStatus
		wantUserID    int
		wantStatusID  int
	}{
		{
			name:          "resolves both user and status when both are missing",
			input:         createTimesheetInput{},
			currentUserID: 4,
			statuses: []timesheetStatus{
				{ID: 1, Name: "Offen"},
				{ID: 2, Name: "Erledigt"},
			},
			wantUserID:   4,
			wantStatusID: 2,
		},
		{
			name: "keeps explicit user and resolves status",
			input: createTimesheetInput{
				UserID: intPtr(9),
			},
			currentUserID: 4,
			statuses: []timesheetStatus{
				{ID: 2, Name: "Erledigt"},
			},
			wantUserID:   9,
			wantStatusID: 2,
		},
		{
			name: "keeps explicit status and resolves user",
			input: createTimesheetInput{
				StatusID: intPtr(8),
			},
			currentUserID: 4,
			statuses: []timesheetStatus{
				{ID: 2, Name: "Erledigt"},
			},
			wantUserID:   4,
			wantStatusID: 8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := resolveTimesheetDefaults(tc.input, tc.currentUserID, tc.statuses)

			assert.Equal(t, tc.wantUserID, got.UserID)
			assert.Equal(t, tc.wantStatusID, got.StatusID)
		})
	}
}
