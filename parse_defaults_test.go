package main

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestParseCurrentUserID(t *testing.T) {
	t.Parallel()

	currentUserCases := []struct {
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

	for _, tc := range currentUserCases {
		tc := tc
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

	statusCases := []struct {
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

	for _, tc := range statusCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotStatuses, err := parseTimesheetStatuses(tc.raw)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assertStatusesEqual(t, tc.wantStatuses, gotStatuses)
		})
	}
}

func assertStatusesEqual(t *testing.T, want []timesheetStatus, got []timesheetStatus) {
	t.Helper()
	assert.Equal(t, want, got)
}
