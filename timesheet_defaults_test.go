package main

import (
	"encoding/json"
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

func TestParseCurrentUserID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		raw     json.RawMessage
		wantID  int
		wantErr bool
	}{
		{
			name:    "parses user id from valid user payload",
			raw:     json.RawMessage(`{"id":4,"firstname":"Rudolph"}`),
			wantID:  4,
			wantErr: false,
		},
		{
			name:    "returns error for malformed user payload",
			raw:     json.RawMessage(`{"id":`),
			wantID:  0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotID, err := parseCurrentUserID(tc.raw)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.wantID, gotID)
		})
	}
}

func TestParseTimesheetStatuses(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		raw          json.RawMessage
		wantStatuses []timesheetStatus
		wantErr      bool
	}{
		{
			name: "parses status list from valid payload",
			raw:  json.RawMessage(`[{"id":1,"name":"Offen"},{"id":2,"name":"Erledigt"}]`),
			wantStatuses: []timesheetStatus{
				{ID: 1, Name: "Offen"},
				{ID: 2, Name: "Erledigt"},
			},
			wantErr: false,
		},
		{
			name:         "returns error for malformed status payload",
			raw:          json.RawMessage(`[{"id":1,"name":"Offen"}`),
			wantStatuses: nil,
			wantErr:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotStatuses, err := parseTimesheetStatuses(tc.raw)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatuses, gotStatuses)
		})
	}
}
